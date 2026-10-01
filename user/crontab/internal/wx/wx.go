package wx

import (
	"crontab/pkg/config"
	"crontab/pkg/log"
	"crontab/pkg/wx-api/wx-official"
)

func GetWxAccessToken(cnf *config.Config) func() {
	return func() {
		for _, item := range cnf.WxOfficials {
			official := wx_official.NewWxOfficial(item.AppId, item.Secret)
			err := official.RefreshToken()
			if err != nil {
				log.Error(err)
				continue
			}

		}
	}
}
