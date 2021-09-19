package config

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	defaultHTTPPort        = 5000
	defaultReadTimeout     = 5 * time.Second
	defaultWriteTimeout    = 5 * time.Second
	defaultShutdownTimeout = 5 * time.Second

	defaultMogoDBName    = "ynab-helper"
	defaultMongoURI      = "mongodb://mongo:27017"
	defaultMongoUser     = "root"
	defaultMongoPassword = "root"
)

// Config is application configuration
type Config struct {
	HTTP  HTTPConfig
	Mongo MongoConfig
}

// HTTPConfig is web related configuration
type HTTPConfig struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"readTimeOut"`
	WriteTimeout    time.Duration `mapstructure:"writeTimeOut"`
	ShutdownTimeout time.Duration `mapstructure:"shutdownTimeout"`
}

// MongoConfig is a configuration for connecting to a MongoDB
type MongoConfig struct {
	Name     string `mapstructure:"name"`
	URI      string `mapstructure:"uri"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// Init initialize Config struct
func Init(configPath string) (*Config, error) {
	populateDefaults()

	if err := parseConfigFile(configPath); err != nil {
		return nil, errors.Wrapf(err, "config.Init() ")
	}

	var cfg Config
	if err := unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil

}

func populateDefaults() {
	viper.SetDefault("http.port", defaultHTTPPort)
	viper.SetDefault("http.readTimeOut", defaultReadTimeout)
	viper.SetDefault("http.writeTimeOut", defaultWriteTimeout)
	viper.SetDefault("http.shutdownTimeout", defaultShutdownTimeout)

	viper.SetDefault("mongo.name", defaultMogoDBName)
	viper.SetDefault("mongo.uri", defaultMongoURI)
	viper.SetDefault("mongo.user", defaultMongoUser)
	viper.SetDefault("mongo.password", defaultMongoPassword)
}

func parseConfigFile(filePath string) error {

	path := strings.Split(filePath, "/")
	if len(path) < 2 {
		return errors.New("invalid path to the configuration file. Should contain '/'")
	}

	viper.AddConfigPath(path[0])
	viper.SetConfigName(path[1])

	return viper.ReadInConfig()
}

func unmarshal(cfg *Config) error {
	if err := viper.UnmarshalKey("http", &cfg.HTTP); err != nil {
		return errors.Wrap(err, "config.unmarshal() Web ")
	}

	if err := viper.UnmarshalKey("mongo", &cfg.Mongo); err != nil {
		return errors.Wrap(err, "config.unmarshal() mongoDB ")
	}
	return nil
}
