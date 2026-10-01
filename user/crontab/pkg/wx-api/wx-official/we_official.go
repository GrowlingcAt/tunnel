package wx_official

import (
	"crontab/pkg/wx-api"
	"fmt"
)

type wxOfficial struct {
	*wx_api.DefaultToken
}

func NewWxOfficial(id, secret string) wx_api.Token {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", id, secret)
	return &wxOfficial{
		DefaultToken: &wx_api.DefaultToken{
			Id:     id,
			Secret: secret,
			Url:    url,
			App:    "",
		},
	}
}
