package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Server   ServerConfig
	DB       DBConfig
	Redis    RedisConfig
	S3       S3Config
	JWT      JWTConfig
	Telegram TelegramConfig
	CORS     CORSConfig
	SMTP     SMTPConfig
	Payment  PaymentConfig
	Bitrix   BitrixConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DBConfig struct {
	URL             string
	MaxConnections  int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	URL      string
	Password string
}

type S3Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	PublicURL string
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type TelegramConfig struct {
	BotToken    string
	WebhookURL  string
	WebAppURL   string
	AdminChatID int64
}

type CORSConfig struct {
	AllowedOrigins []string
}

type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

// BitrixConfig holds Bitrix24 integration settings.
// BITRIX_PORTAL: e.g. "company.bitrix24.ru"
// BITRIX_USER_ID: numeric user ID from the inbound webhook URL
// BITRIX_TOKEN: webhook access token
type BitrixConfig struct {
	Portal string
	UserID int
	Token  string
}

func (b *BitrixConfig) IsConfigured() bool {
	return b.Portal != "" && b.Token != "" && b.UserID > 0
}

// PaymentConfig holds payment gateway settings.
// PAYMENT_PROVIDER controls which provider is active: "mock" (default) | "yookassa" | "tinkoff".
// APP_URL is the public frontend URL used to build return and webhook URLs.
type PaymentConfig struct {
	Provider string // "mock" by default; switch when integrating a real gateway
	AppURL   string // e.g. "https://avangard-print.ru"
}

func (p *PaymentConfig) IsMock() bool {
	return p.Provider == "" || p.Provider == "mock"
}

func Load() (*Config, error) {
	viper.AutomaticEnv()

	// Try .env in current dir, then parent (monorepo root)
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		viper.SetConfigFile("../.env")
		_ = viper.ReadInConfig()
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: getStringOrDefault("PORT", "8080"),
			Env:  getStringOrDefault("ENV", "development"),
		},
		DB: DBConfig{
			URL:             viper.GetString("DATABASE_URL"),
			MaxConnections:  getIntOrDefault("DB_MAX_CONNECTIONS", 25),
			MaxIdleConns:    getIntOrDefault("DB_MAX_IDLE", 5),
			ConnMaxLifetime: 5 * time.Minute,
		},
		Redis: RedisConfig{
			URL:      getStringOrDefault("REDIS_URL", "redis://localhost:6379"),
			Password: viper.GetString("REDIS_PASSWORD"),
		},
		S3: S3Config{
			Endpoint:  viper.GetString("S3_ENDPOINT"),
			Region:    getStringOrDefault("S3_REGION", "us-east-1"),
			Bucket:    viper.GetString("S3_BUCKET"),
			AccessKey: viper.GetString("S3_ACCESS_KEY"),
			SecretKey: viper.GetString("S3_SECRET_KEY"),
			PublicURL: viper.GetString("S3_PUBLIC_URL"),
		},
		JWT: JWTConfig{
			Secret:        viper.GetString("JWT_SECRET"),
			AccessExpiry:  getDurationOrDefault("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshExpiry: getDurationOrDefault("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
		},
		Telegram: TelegramConfig{
			BotToken:    viper.GetString("TELEGRAM_BOT_TOKEN"),
			WebhookURL:  viper.GetString("TELEGRAM_WEBHOOK_URL"),
			WebAppURL:   getStringOrDefault("TELEGRAM_WEBAPP_URL", "https://avangard-print.ru"),
			AdminChatID: viper.GetInt64("TELEGRAM_ADMIN_CHAT_ID"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getStringSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		},
		SMTP: SMTPConfig{
			Host:      viper.GetString("SMTP_HOST"),
			Port:      getIntOrDefault("SMTP_PORT", 465),
			Username:  viper.GetString("SMTP_USERNAME"),
			Password:  viper.GetString("SMTP_PASSWORD"),
			FromEmail: getStringOrDefault("SMTP_FROM_EMAIL", "noreply@avangard-print.ru"),
			FromName:  getStringOrDefault("SMTP_FROM_NAME", "АВАНГАРД"),
		},
		Payment: PaymentConfig{
			Provider: getStringOrDefault("PAYMENT_PROVIDER", "mock"),
			AppURL:   getStringOrDefault("APP_URL", "https://avangard-print.ru"),
		},
		Bitrix: BitrixConfig{
			Portal: viper.GetString("BITRIX_PORTAL"),
			UserID: getIntOrDefault("BITRIX_USER_ID", 0),
			Token:  viper.GetString("BITRIX_TOKEN"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DB.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}

func (s *SMTPConfig) IsConfigured() bool {
	return s.Host != "" && s.Username != "" && s.Password != ""
}

func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// LogConfig logs the loaded configuration with secrets masked.
func (c *Config) LogConfig(log *zap.Logger) {
	log.Info("configuration loaded",
		zap.String("server.port", c.Server.Port),
		zap.String("server.env", c.Server.Env),
		zap.String("db.url", maskDSN(c.DB.URL)),
		zap.Int("db.maxConnections", c.DB.MaxConnections),
		zap.Int("db.maxIdleConns", c.DB.MaxIdleConns),
		zap.String("redis.url", c.Redis.URL),
		zap.String("s3.endpoint", c.S3.Endpoint),
		zap.String("s3.bucket", c.S3.Bucket),
		zap.String("s3.publicURL", c.S3.PublicURL),
		zap.Duration("jwt.accessExpiry", c.JWT.AccessExpiry),
		zap.Duration("jwt.refreshExpiry", c.JWT.RefreshExpiry),
		zap.Bool("telegram.configured", c.Telegram.BotToken != ""),
		zap.Strings("cors.allowedOrigins", c.CORS.AllowedOrigins),
		zap.Bool("smtp.configured", c.SMTP.Host != ""),
		zap.String("payment.provider", c.Payment.Provider),
		zap.String("payment.appURL", c.Payment.AppURL),
		zap.Bool("bitrix.configured", c.Bitrix.IsConfigured()),
		zap.String("bitrix.portal", c.Bitrix.Portal),
	)
}

// maskDSN hides password in a DSN string like postgresql://user:pass@host/db
func maskDSN(dsn string) string {
	atIdx := strings.Index(dsn, "@")
	if atIdx == -1 {
		return dsn
	}
	protoEnd := strings.Index(dsn, "://")
	if protoEnd == -1 {
		return "***"
	}
	userInfo := dsn[protoEnd+3 : atIdx]
	colonIdx := strings.Index(userInfo, ":")
	if colonIdx == -1 {
		return dsn
	}
	user := userInfo[:colonIdx]
	return dsn[:protoEnd+3] + user + ":***@" + dsn[atIdx+1:]
}

func getStringOrDefault(key, defaultVal string) string {
	val := viper.GetString(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getIntOrDefault(key string, defaultVal int) int {
	val := viper.GetInt(key)
	if val == 0 {
		return defaultVal
	}
	return val
}

func getDurationOrDefault(key string, defaultVal time.Duration) time.Duration {
	val := viper.GetString(key)
	if val == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return d
}

func getStringSlice(key string, defaultVal []string) []string {
	val := viper.GetString(key)
	if val == "" {
		return defaultVal
	}
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
