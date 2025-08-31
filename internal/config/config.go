package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"gopkg.in/yaml.v3"
)

var ErrConfigPathIsEmpty = errors.New("config path is empty")

type Config struct {
	App        `yaml:"app"`
	Logger     `yaml:"log"`
	Database   `yaml:"database"`
	Redis      `yaml:"redis"`
	Kafka      `yaml:"kafka"`
	HTTPServer `yaml:"http_server"`
}

type App struct {
	ServiceName string `yaml:"service_name"`
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

type Redis struct {
	Enable   bool   `yaml:"enable"`
	Host     string `yaml:"host"`
	Port     uint16 `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type Kafka struct {
	Brokers    []string   `yaml:"brokers"`
	Subscriber Subscriber `yaml:"subscriber"`
	Producer   Producer   `yaml:"producer"`
}

type Subscriber struct {
	Name             string           `yaml:"name"`
	WorkerCount      int              `yaml:"worker_count"`
	OrdersSubscriber OrdersSubscriber `yaml:"orders_subscriber"`
}

type OrdersSubscriber struct {
	BufferSize int    `yaml:"buffer_size"`
	Topic      string `yaml:"topic"`
	GroupID    string `yaml:"group_id"`
}

type Producer struct {
	Name           string         `yaml:"name"`
	WorkerCount    int            `yaml:"worker_count"`
	OrdersProducer OrdersProducer `yaml:"orders_producer"`
}

type OrdersProducer struct {
	Topic string `yaml:"topic"`
}

type HTTPServer struct {
	Host     string  `yaml:"host"`
	Port     uint16  `yaml:"port"`
	BasePath string  `yaml:"base_path"`
	Timeout  Timeout `yaml:"timeout"`
}

type Timeout struct {
	Request time.Duration `yaml:"request"`
	Read    time.Duration `yaml:"read"`
	Write   time.Duration `yaml:"write"`
	Idle    time.Duration `yaml:"idle"`
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

	if result == "" {
		result = os.Getenv("CONFIG_PATH")
	}

	return result
}
