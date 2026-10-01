package login

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"user/pkg/zjwt"
)

func (c *LoginController) CheckLogin(ctx *gin.Context) {
	accessToken := ctx.DefaultQuery("access_token", "")
	if accessToken == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{})
		return
	}
	hs := zjwt.NewHs(c.config.Jwt.HashKey, zjwt.HS256)
	userClaims := &zjwt.UserClaims{}
	err := hs.Verify(accessToken, userClaims)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusUnauthorized, gin.H{})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":         userClaims.UserID,
		"name":       userClaims.Name,
		"avatar_url": userClaims.Avatar,
	})
}
