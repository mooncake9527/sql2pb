package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type App struct {
	Out string `yaml:"out"`
	Tpl string `yaml:"tpl"`
	DB  *DB    `yaml:"db"`
}

type DB struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Schema   string `yaml:"schema"`
	Tables   string `yaml:"tables"`
}

var AppConfig *App

func Parse(configYml string) {
	if configYml == "" {
		panic("找不到配置文件")
	}
	v := viper.New()
	v.SetConfigFile(configYml)
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Fatal error config file: %v \n", err))
	}

	cfg := App{}
	if err = v.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	AppConfig = &cfg
}
