package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	BootstrapToken string `mapstructure:"BOOTSTRAP_TOKEN"`
	ListenPort     string `mapstructure:"LISTEN_PORT"`
	LogLevel       string `mapstructure:"LOG_LEVEL"`
	DBHost         string `mapstructure:"DB_HOST"`
	DBPort         string `mapstructure:"DB_PORT"`
	DBUser         string `mapstructure:"DB_USER"`
	DBPwd          string `mapstructure:"DB_PWD"`
}

var GlobalConfig *Config

func init() {
	Setup("config.yml")
}

func getDefaultConfig() Config {
	return Config{
		ListenPort:     "9988",
		BootstrapToken: "",
		LogLevel:       "error",
		DBHost:         "localhost",
		DBPort:         "5432",
		DBUser:         "postgres",
		DBPwd:          "postgres",
	}
}

func loadConfigFromFile(path string, conf *Config) {
	var err error
	_, err = os.Stat(path)
	if err == nil {
		fileViper := viper.New()
		fileViper.SetConfigFile(path)
		if err = fileViper.ReadInConfig(); err == nil {
			if err = fileViper.Unmarshal(conf); err == nil {
				log.Printf("Load config from %s success\n", path)
				return
			}
		}
	}
}

func Setup(configPath string) {
	var conf = getDefaultConfig()
	loadConfigFromFile(configPath, &conf)
	GlobalConfig = &conf
	log.Printf("%+v\n", GlobalConfig)
}

func GetConf() Config {
	if GlobalConfig == nil {
		return getDefaultConfig()
	}
	return *GlobalConfig
}
