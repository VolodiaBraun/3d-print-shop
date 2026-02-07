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
	BotToken   string
	WebhookURL string
}

type CORSConfig struct {
	AllowedOrigins []string
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
			BotToken:   viper.GetString("TELEGRAM_BOT_TOKEN"),
			WebhookURL: viper.GetString("TELEGRAM_WEBHOOK_URL"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getStringSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
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
