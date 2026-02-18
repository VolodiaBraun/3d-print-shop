package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/bitrix"
	"github.com/brown/3d-print-shop/internal/cache"
	"github.com/brown/3d-print-shop/internal/config"
	mockdelivery "github.com/brown/3d-print-shop/internal/delivery/mock"
	"github.com/brown/3d-print-shop/internal/handler"
	"github.com/brown/3d-print-shop/internal/middleware"
	mockpayment "github.com/brown/3d-print-shop/internal/payment/mock"
	"github.com/brown/3d-print-shop/internal/repository/postgres"
	"github.com/brown/3d-print-shop/internal/service"
	"github.com/brown/3d-print-shop/internal/storage"
	tgbot "github.com/brown/3d-print-shop/internal/telegram"
	jwtpkg "github.com/brown/3d-print-shop/pkg/jwt"
	"github.com/brown/3d-print-shop/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Server.Env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	cfg.LogConfig(log)

	// Connect to database
	db, err := postgres.NewDB(cfg.DB, log)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer postgres.CloseDB(db, log)

	// Connect to S3
	s3Client, err := storage.NewS3Client(cfg.S3, log)
	if err != nil {
		log.Warn("failed to connect to s3, image uploads disabled", zap.Error(err))
	}

	// Repositories
	userRepo := postgres.NewUserRepo(db)
	categoryRepo := postgres.NewCategoryRepo(db)
	productRepo := postgres.NewProductRepo(db)
	productImageRepo := postgres.NewProductImageRepo(db)
	cartRepo := postgres.NewCartRepo(db)
	promoRepo := postgres.NewPromoRepo(db)

	// Connect to Redis
	redisClient, err := cache.NewRedis(cfg.Redis, log)
	if err != nil {
		log.Fatal("failed to connect to redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize JWT manager and auth token service
	jwtManager := jwtpkg.NewManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	authTokenService := service.NewAuthTokenService(jwtManager, redisClient, log)

	// Cache store
	cacheStore := cache.NewStore(redisClient)

	// Services
	authService := service.NewAuthService(userRepo, authTokenService, cfg.Telegram.BotToken, log)
	categoryService := service.NewCategoryService(categoryRepo, cacheStore, log)
	productService := service.NewProductService(productRepo, categoryRepo, cacheStore, log)
	imageService := service.NewImageService(productImageRepo, productRepo, s3Client, log)
	cartService := service.NewCartService(cartRepo, productRepo, log)
	promoService := service.NewPromoService(promoRepo, log)
	orderRepo := postgres.NewOrderRepo(db)
	customOrderRepo := postgres.NewCustomOrderRepo(db)
	orderService := service.NewOrderService(orderRepo, productRepo, userRepo, promoService, db, log)
	customOrderService := service.NewCustomOrderService(orderRepo, customOrderRepo, userRepo, db, log)

	// Delivery
	deliveryZoneRepo := postgres.NewDeliveryZoneRepo(db)
	pickupPointRepo := postgres.NewPickupPointRepo(db)
	mockProvider := mockdelivery.New(deliveryZoneRepo)
	deliveryService := service.NewDeliveryService(mockProvider, deliveryZoneRepo, pickupPointRepo, log)
	orderService.SetDeliveryService(deliveryService)

	// Payment (provider-agnostic; swap mockpayment for yookassa/tinkoff when ready)
	paymentProvider := mockpayment.New(cfg.Payment.AppURL)
	paymentService := service.NewPaymentService(paymentProvider, orderRepo, db, log, cfg.Payment.AppURL)
	orderService.SetPaymentService(paymentService)
	customOrderService.SetPaymentService(paymentService)
	if s3Client != nil {
		customOrderService.SetS3Client(s3Client)
	}
	paymentHandler := handler.NewPaymentHandler(paymentService)
	log.Info("payment provider initialized", zap.String("provider", paymentProvider.Name()))

	// Telegram bot (optional)
	var telegramBot *tgbot.Bot
	if cfg.Telegram.BotToken != "" {
		telegramBot, err = tgbot.New(cfg.Telegram, orderService, userRepo, log)
		if err != nil {
			log.Warn("telegram bot failed to initialize", zap.Error(err))
		} else {
			log.Info("telegram bot initialized", zap.String("username", telegramBot.Username()))
			orderService.SetNotifier(telegramBot)
			customOrderService.SetNotifier(telegramBot)
		}
	}

	// Email (optional)
	var emailService *service.EmailService
	if cfg.SMTP.IsConfigured() {
		es, err := service.NewEmailService(cfg.SMTP, log)
		if err != nil {
			log.Warn("email service failed to initialize", zap.Error(err))
		} else {
			log.Info("email service initialized", zap.String("from", cfg.SMTP.FromEmail))
			emailService = es
			orderService.SetEmailService(emailService)
			customOrderService.SetEmailService(emailService)
		}
	}

	// Loyalty
	bonusTransactionRepo := postgres.NewBonusTransactionRepo(db)
	loyaltySettingsRepo := postgres.NewLoyaltySettingsRepo(db)
	loyaltyService := service.NewLoyaltyService(userRepo, bonusTransactionRepo, loyaltySettingsRepo, db, log)
	authService.SetLoyaltyService(loyaltyService)
	orderService.SetLoyaltyService(loyaltyService)

	// User and Review services
	userService := service.NewUserService(userRepo, log)
	userService.SetCache(cacheStore)
	if emailService != nil {
		userService.SetEmailService(emailService)
	}
	reviewRepo := postgres.NewReviewRepo(db)
	reviewService := service.NewReviewService(reviewRepo, orderRepo, productRepo, db, log)

	// Content blocks
	contentBlockRepo := postgres.NewContentBlockRepo(db)
	contentService := service.NewContentService(contentBlockRepo, cacheStore, log)

	// Analytics
	analyticsRepo := postgres.NewAnalyticsRepo(db)
	analyticsService := service.NewAnalyticsService(analyticsRepo, db, log)

	// Bitrix24 service (optional; route registered after router is created)
	var bitrixHandler *handler.BitrixHandler
	if cfg.Bitrix.IsConfigured() {
		bitrixClient := bitrix.NewClient(cfg.Bitrix.Portal, cfg.Bitrix.UserID, cfg.Bitrix.Token)
		bitrixService := service.NewBitrixService(bitrixClient, orderRepo, customOrderRepo, log)
		customOrderService.SetBitrixService(bitrixService)
		bitrixHandler = handler.NewBitrixHandler(bitrixService)
		log.Info("bitrix24 integration enabled", zap.String("portal", cfg.Bitrix.Portal))
	} else {
		log.Info("bitrix24 integration disabled (BITRIX_PORTAL/BITRIX_USER_ID/BITRIX_TOKEN not set)")
	}

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	productHandler := handler.NewProductHandler(productService)
	imageHandler := handler.NewImageHandler(imageService)
	cartHandler := handler.NewCartHandler(cartService)
	promoHandler := handler.NewPromoHandler(promoService)
	orderHandler := handler.NewOrderHandler(orderService)
	deliveryHandler := handler.NewDeliveryHandler(deliveryService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	loyaltyHandler := handler.NewLoyaltyHandler(loyaltyService)
	contentHandler := handler.NewContentHandler(contentService)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)
	customOrderHandler := handler.NewCustomOrderHandler(customOrderService)

	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())
	router.MaxMultipartMemory = 52 << 20 // 52 MB (для 3D-моделей до 50 MB)

	// CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	authHandler.RegisterRoutes(v1)
	categoryHandler.RegisterPublicRoutes(v1)
	productHandler.RegisterPublicRoutes(v1)
	promoHandler.RegisterPublicRoutes(v1)
	orderHandler.RegisterPublicRoutes(v1)
	deliveryHandler.RegisterPublicRoutes(v1)
	reviewHandler.RegisterPublicRoutes(v1)
	contentHandler.RegisterPublicRoutes(v1)
	customOrderHandler.RegisterPublicRoutes(v1)
	authMw := middleware.AuthRequired(jwtManager)
	userHandler.RegisterProtectedRoutes(v1.Group("", authMw))
	reviewHandler.RegisterProtectedRoutes(v1.Group("", authMw))
	orderHandler.RegisterProtectedRoutes(v1.Group("", authMw))
	loyaltyHandler.RegisterProtectedRoutes(v1.Group("", authMw))
	cartHandler.RegisterRoutes(v1, authMw)

	// Protected admin routes
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthRequired(jwtManager))
	admin.Use(middleware.RequireRole("admin"))
	admin.GET("/me", func(c *gin.Context) {
		userID, _ := middleware.GetUserID(c)
		role, _ := middleware.GetRole(c)
		c.JSON(http.StatusOK, gin.H{"userID": userID, "role": role})
	})
	categoryHandler.RegisterAdminRoutes(admin)
	productHandler.RegisterAdminRoutes(admin)
	imageHandler.RegisterAdminRoutes(admin)
	promoHandler.RegisterAdminRoutes(admin)
	orderHandler.RegisterAdminRoutes(admin)
	deliveryHandler.RegisterAdminRoutes(admin)
	reviewHandler.RegisterAdminRoutes(admin)
	loyaltyHandler.RegisterAdminRoutes(admin)
	contentHandler.RegisterAdminRoutes(admin)
	analyticsHandler.RegisterAdminRoutes(admin)
	customOrderHandler.RegisterAdminRoutes(admin)

	// Payment routes
	paymentHandler.RegisterWebhookRoute(router)        // POST /webhook/payment
	paymentHandler.RegisterMockRoutes(v1)              // GET  /api/v1/payment/mock/:orderNumber
	paymentHandler.RegisterAdminRoutes(admin)          // POST /admin/orders/:id/regenerate-payment

	// Bitrix24 incoming webhook route
	if bitrixHandler != nil {
		bitrixHandler.RegisterWebhookRoute(router)     // POST /webhook/bitrix
	}

	// Register Telegram webhook route
	if telegramBot != nil {
		telegramBot.RegisterWebhook(router)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// Background stats aggregation
	aggCtx, aggCancel := context.WithCancel(context.Background())
	go analyticsService.StartBackgroundAggregation(aggCtx)

	// Start server in goroutine
	go func() {
		log.Info("listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	aggCancel()

	if telegramBot != nil {
		telegramBot.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", zap.Error(err))
	}

	log.Info("server stopped")
}
