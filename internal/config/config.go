// Package config: application configuration from environment variables.
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	App      App
	HTTP     HTTP
	Database Database
	JWT      JWT
	SMTP     SMTP
	OTP      OTP
	Log      Log
}

type App struct {
	Name    string `env:"APP_NAME" envDefault:"ledgerline-backend"`
	Env     string `env:"APP_ENV" envDefault:"development"`
	Version string `env:"APP_VERSION" envDefault:"0.1.0"`
}

func (a App) IsProduction() bool { return a.Env == "production" }

type HTTP struct {
	Host            string        `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	Port            int           `env:"HTTP_PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout     time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"15s"`
	RequestTimeout  time.Duration `env:"HTTP_REQUEST_TIMEOUT" envDefault:"10s"`
	AllowedOrigins  []string      `env:"HTTP_ALLOWED_ORIGINS" envSeparator:"," envDefault:"*"`
	RateLimitRPS    int           `env:"HTTP_RATE_LIMIT_RPS" envDefault:"20"`
	RateLimitBurst  int           `env:"HTTP_RATE_LIMIT_BURST" envDefault:"40"`
}

func (h HTTP) Addr() string { return fmt.Sprintf("%s:%d", h.Host, h.Port) }

type Database struct {
	Host            string        `env:"DB_HOST" envDefault:"localhost"`
	Port            int           `env:"DB_PORT" envDefault:"5432"`
	User            string        `env:"DB_USER" envDefault:"postgres"`
	Password        string        `env:"DB_PASSWORD" envDefault:"postgres"`
	Name            string        `env:"DB_NAME" envDefault:"ledgerline"`
	SSLMode         string        `env:"DB_SSLMODE" envDefault:"disable"`
	MaxConns        int32         `env:"DB_MAX_CONNS" envDefault:"20"`
	MinConns        int32         `env:"DB_MIN_CONNS" envDefault:"2"`
	MaxConnLifetime time.Duration `env:"DB_MAX_CONN_LIFETIME" envDefault:"1h"`
	MaxConnIdleTime time.Duration `env:"DB_MAX_CONN_IDLE_TIME" envDefault:"30m"`
}

// DSN for GORM & golang-migrate.
func (d Database) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

type JWT struct {
	Secret          string        `env:"JWT_SECRET,required"`
	Issuer          string        `env:"JWT_ISSUER" envDefault:"ledgerline"`
	AccessTokenTTL  time.Duration `env:"JWT_ACCESS_TOKEN_TTL" envDefault:"15m"`
	RefreshTokenTTL time.Duration `env:"JWT_REFRESH_TOKEN_TTL" envDefault:"168h"`
	ResetTokenTTL   time.Duration `env:"JWT_RESET_TOKEN_TTL" envDefault:"10m"`
}

type OTP struct {
	Length int           `env:"OTP_LENGTH" envDefault:"6"`
	TTL    time.Duration `env:"OTP_TTL" envDefault:"10m"`
}

type SMTP struct {
	Host        string        `env:"SMTP_HOST" envDefault:"localhost"`
	Port        int           `env:"SMTP_PORT" envDefault:"1025"`
	Username    string        `env:"SMTP_USERNAME"`
	Password    string        `env:"SMTP_PASSWORD"`
	FromName    string        `env:"SMTP_FROM_NAME" envDefault:"LedgerLine"`
	FromAddress string        `env:"SMTP_FROM_ADDRESS" envDefault:"no-reply@ledgerline.local"`
	TLS         string        `env:"SMTP_TLS" envDefault:"opportunistic"`
	Timeout     time.Duration `env:"SMTP_TIMEOUT" envDefault:"10s"`
	Enabled     bool          `env:"SMTP_ENABLED" envDefault:"false"`
}

type Log struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

// Load reads .env (optional) then parses environment variables.
func Load() (*Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}
	return &cfg, nil
}
