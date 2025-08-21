package config

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"gopkg.in/yaml.v3"
)

var ErrConfigPathIsEmpty = errors.New("config path is empty")

type Config struct {
	App      `yaml:"app"`
	Logger   `yaml:"log"`
	Database `yaml:"database"`
}

type App struct {
	Name string `yaml:"name"`
}

type Logger struct {
	Level      string   `yaml:"level"`
	FormatJSON bool     `yaml:"format_json"`
	Rotation   Rotation `yaml:"rotation"`
}

type Rotation struct {
	File       string `json:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

type Database struct {
	Host      string    `yaml:"host"`
	Port      uint16    `yaml:"port"`
	User      string    `yaml:"user"`
	Password  string    `yaml:"password"`
	Name      string    `yaml:"name"`
	SSLMode   string    `yaml:"ssl_mode"`
	MaxConns  int32     `yaml:"max_conns"`
	MinConns  int32     `yaml:"min_conns"`
	Migration Migration `yaml:"migration"`
}

type Migration struct {
	Path      string `yaml:"path"`
	AutoApply bool   `yaml:"auto_apply"`
}

func MustLoadConfig() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	return cfg
}

func LoadConfig() (*Config, error) {
	path := fetchConfigPath()
	if path == "" {
		return nil, ErrConfigPathIsEmpty
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}

	var config Config

	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &config, nil
}

func MustPrintConfig(cfg *Config) {
	if err := PrintConfig(cfg); err != nil {
		panic(err)
	}
}

func PrintConfig(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}

func fetchConfigPath() string {
	var result string

	flag.StringVar(&result, "config", "", "Path to config file")
	flag.Parse()

	return result
}
