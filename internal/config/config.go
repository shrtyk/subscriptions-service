package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

var path string

func init() {
	flag.StringVar(&path, "cfg_path", "", "Path to config file")
}

type Config struct {
	PostgresCfg PostgresCfg `yaml:"postgres"`
	RepoCfg     RepoConfig  `yaml:"repository"`
}

type RepoConfig struct {
	DefaultPageSize int `yaml:"default_page_size" env:"REPO_DEFAULT_PAGE_SIZE" env-default:"10"`
	MaxPageSize     int `yaml:"max_page_size" env:"REPO_MAX_PAGE_SIZE" env-default:"100"`
}

type PostgresCfg struct {
	User     string `yaml:"user" env:"PG_USER" env-default:"user"`
	Password string `yaml:"password" env:"PG_PASSWORD" env-default:"password"`
	Host     string `yaml:"host" env:"PG_HOST" env-default:"postgres"`
	Port     string `yaml:"port" env:"PG_PORT" env-default:"5432"`
	DBName   string `yaml:"db_name" env:"PG_DBNAME" env-default:"subscriptions-db"`
	SSLMode  string `yaml:"sslmode" env:"PG_SSLMODE" env-default:"disable"`

	MaxOpenConns    int           `yaml:"max_open_conns" env:"PG_MAX_OPEN_CONNS" env-default:"20"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"PG_MAX_IDLE_CONS" env-default:"10"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetine" env:"PG_CONN_MAX_LIFETIME" env-default:"30m"`
	ConnMaxIdletime time.Duration `yaml:"conn_max_idletime" env:"PG_CONN_MAX_IDLETIME" env-default:"5m"`
}

func MustInitConfig() *Config {
	cfgPath := cfgPath()
	cfg := new(Config)

	if cfgPath != "" {
		err := cleanenv.ReadConfig(cfgPath, cfg)
		if err != nil && !os.IsNotExist(err) {
			panic(fmt.Sprintf("failed to read config file: %s", err))
		}
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		panic(fmt.Sprintf("failed to read environment variables: %s", err))
	}

	return cfg
}

func cfgPath() string {
	if !flag.Parsed() {
		flag.Parse()
	}

	if path == "" {
		return os.Getenv("CONFIG_PATH")
	}

	return path
}
