package config

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	Http struct {
		IP   string
		Port int
		Mode string
	}
	Mysql struct {
		DSN         string
		MaxLifeTime int
		MaxOpenConn int
		MaxIdleConn int
	}
	Redis struct {
		Host string
		Port int
		Pwd  string `mapstructure:"pwd"`
	}
	Log struct {
		Level   string
		LogPath string `mapstructure:"logPath"`
	} `mapstructure:"log"`
	Cos struct {
		SecretId  string
		SecretKey string
		BucketUrl string
		CDNDomain string
	}
	DependOn struct {
		Crontab struct {
			Address     string
			AccessToken string
		}
	}
	Gitlab struct {
		ClientID     string
		ClientSecret string
		RedirectUri  string
		Domain       string
	}
	Github struct {
		ClientID     string
		ClientSecret string
		RedirectUri  string
	}
	Jwt struct {
		HashKey string
	}
	InternalSystemEntry map[string]string
	WxOfficial          struct {
		AppID    string
		MsgToken string
	}
}

var conf *Config

func InitConfig(filePath string, typ ...string) {
	v := viper.New()
	v.SetConfigFile(filePath)
	if len(typ) > 0 {
		v.SetConfigType(typ[0])
	}
	err := v.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
	conf = &Config{}
	err = v.Unmarshal(conf)
	if err != nil {
		log.Fatal(err)
	}

}

func GetConfig() *Config {
	return conf
}
