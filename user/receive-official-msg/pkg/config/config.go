package config

import (
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	Http struct {
		IP   string
		Port int
		Mode string
	}
	WxOfficial struct {
		AppId    string `mapstructure:"appid"`
		Secret   string `mapstructure:"secret"`
		MsgToken string `mapstructure:"msgToken"`
	} `mapstructure:"wx_official"`
}

var cfg *Config

func InitConf(configPath string) {
	if configPath == "" {
		panic("请指定应用程序配置文件")
	}
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		panic("配置文件不存在")
	}
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(configPath)
	v.ReadInConfig()
	cfg = &Config{}
	err = v.Unmarshal(cfg)
	if err != nil {
		panic(err.Error())
	}
}

func GetConf() *Config {
	return cfg
}
