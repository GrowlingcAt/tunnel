package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"tunnel/pkg/config"
	"tunnel/pkg/log"
	"tunnel/pkg/zerror"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer ")
		fmt.Println(token)
		if token == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		user, err := checkAuth(token)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Error(err)
			return
		}
		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("User.ID", user.ID)
		c.Set("User.Name", user.Name)
		c.Set("User.AvatarUrl", user.AvatarUrl)
		c.Next()
	}
}

type userInfo struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
}

func checkAuth(token string) (*userInfo, error) {
	conf := config.GetConfig()
	path := "/api/v1/login/check/auth"
	url := fmt.Sprintf("%s%s?access_token=%s", conf.DependOnServices.User.Address, path, token)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 401 {
		return nil, nil
	}
	if res.StatusCode == 500 {
		err = zerror.NewByMsg("服务器内部错误")
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	user := &userInfo{}
	err = json.Unmarshal(body, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
