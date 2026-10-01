package login

import (
	"net/http"
	"time"
	"user/pkg/zjwt"
)

func (c *LoginController) getAcrossSubdomainCookie(accessToken string) []*http.Cookie {
	accessTokenCookie := &http.Cookie{
		Name:    "sso_0voice_access_token",
		Value:   accessToken,
		Path:    "/",
		Domain:  ".growlingcat.cn",
		Expires: time.Now().Add(time.Second * time.Duration(zjwt.EXPIRES_IN)),
	}
	localAccessTokenCookie := &http.Cookie{
		Name:    "sso_0voice_access_token",
		Value:   accessToken,
		Path:    "/",
		Expires: time.Now().Add(time.Second * time.Duration(zjwt.EXPIRES_IN)),
	}
	return []*http.Cookie{accessTokenCookie, localAccessTokenCookie}
}
