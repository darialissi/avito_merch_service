package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	App   AppConfig   `yaml:"app"`
	DB    DBConfig    `yaml:"db"`
	Redis RedisConfig `yaml:"redis"`
}

type AppConfig struct {
	Port      string     `yaml:"port" env:"APP_PORT"`
	Timeout   AppTimeout `yaml:"timeout"`
	JWTConfig JWTConfig  `yaml:"jwt"`
}

type AppTimeout struct {
	Request  string `yaml:"request"  env:"APP_TIMEOUT_REQUEST"`
	Graceful string `yaml:"graceful" env:"APP_TIMEOUT_GRACEFUL"`
}

type JWTConfig struct {
	AccessToken  AccessJWT  `yaml:"access"`
	RefreshToken RefreshJWT `yaml:"refresh"`
}

type AccessJWT struct {
	Exp    string `yaml:"exp"  env:"ACCESS_TOKEN_EXP"`
	Secret string `yaml:"secret"  env:"ACCESS_TOKEN_SECRET"`
}

type RefreshJWT struct {
	Exp    string `yaml:"exp"  env:"REFRESH_TOKEN_EXP"`
	Secret string `yaml:"secret"  env:"REFRESH_TOKEN_SECRET"`
}

type RedisConfig struct {
	Host     string `yaml:"host"     env:"REDIS_STORAGE_HOST"`
	Port     int32  `yaml:"port"     env:"REDIS_STORAGE_PORT"`
	Password string `yaml:"password" env:"REDIS_STORAGE_PASSWORD"`
}

type DBConfig struct {
	Port     int32  `yaml:"port"     env:"DB_PORT"`
	Host     string `yaml:"host"     env:"DB_HOST"`
	Name     string `yaml:"name"     env:"DB_NAME"`
	User     string `yaml:"user"     env:"DB_USER"`
	Password string `yaml:"password" env:"DB_PASSWORD"`

	Pool DBPool `yaml:"pool"`
}

type DBPool struct {
	MaxConnIdleTime     string `yaml:"max_conn_idletime" env:"DB_POOL_MAX_CONN_IDLE_TIME"`
	MaxConnLifeTime     string `yaml:"max_conn_lifetime" env:"DB_POOL_MAX_CONN_LIFETIME"`
	MinConnectionsCount int32  `yaml:"min_conns"         env:"DB_POOL_MIN_CONNS"`
	MaxConnectionsCount int32  `yaml:"max_conns"         env:"DB_POOL_MAX_CONNS"`
}

func (cfg *DBConfig) getConnURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
}

func (cfg *DBConfig) CreatePool(ctx context.Context) (*pgxpool.Pool, error) {
	connUrl := cfg.getConnURL()

	// Парсим URL в конфигурацию
	poolConfig, err := pgxpool.ParseConfig(connUrl)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Применяем настройки пула
	if err := cfg.applyPoolSettings(poolConfig); err != nil {
		return nil, fmt.Errorf("apply pool settings: %w", err)
	}

	// Создаем пул с конфигурацией
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Выводим информацию о пуле
	log.Printf("Database pool configured: min=%d, max=%d, lifetime=%s, idle=%s",
		poolConfig.MinConns,
		poolConfig.MaxConns,
		poolConfig.MaxConnLifetime,
		poolConfig.MaxConnIdleTime,
	)

	return pool, nil
}

func (cfg *DBConfig) applyPoolSettings(poolConfig *pgxpool.Config) error {
	pool := &cfg.Pool

	// Минимальное количество соединений
	if pool.MinConnectionsCount > 0 {
		poolConfig.MinConns = pool.MinConnectionsCount
	} else {
		poolConfig.MinConns = 1 // значение по умолчанию
	}

	// Максимальное количество соединений
	if pool.MaxConnectionsCount > 0 {
		poolConfig.MaxConns = pool.MaxConnectionsCount
	} else {
		poolConfig.MaxConns = 10 // значение по умолчанию
	}

	// Максимальное время жизни соединения
	if pool.MaxConnLifeTime != "" {
		lifetime, err := time.ParseDuration(pool.MaxConnLifeTime)
		if err != nil {
			return fmt.Errorf("parse max_conn_lifetime: %w", err)
		}
		poolConfig.MaxConnLifetime = lifetime
	} else {
		poolConfig.MaxConnLifetime = 1 * time.Hour // значение по умолчанию
	}

	// Максимальное время простоя соединения
	if pool.MaxConnIdleTime != "" {
		idleTime, err := time.ParseDuration(pool.MaxConnIdleTime)
		if err != nil {
			return fmt.Errorf("parse max_conn_idle_time: %w", err)
		}
		poolConfig.MaxConnIdleTime = idleTime
	} else {
		poolConfig.MaxConnIdleTime = 10 * time.Minute // значение по умолчанию
	}

	// Дополнительные настройки
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	return nil
}

func GetConfig(path string) (Config, error) {
	var cfg Config

	log.Printf("Loading config from %s", path)

	err := cleanenv.ReadConfig(path, &cfg)

	return cfg, err
}
