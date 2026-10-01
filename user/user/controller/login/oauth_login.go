package login

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"net/http"
	"time"
	"user/data"
	"user/pkg/config"
	"user/pkg/constants"
	"user/pkg/db/redis"
	github2 "user/pkg/github"
	gitlab2 "user/pkg/gitlab"
	"user/pkg/log"
	"user/pkg/storage"
	"user/pkg/utils"
	"user/pkg/zerror"
	"user/pkg/zjwt"
)

type LoginController struct {
	log       log.ILogger
	config    *config.Config
	data      data.IData
	redisPool redis.RedisPool
	sf        storage.StorageFactory
}

func NewLoginController(log log.ILogger, config *config.Config, redisPool redis.RedisPool, sf storage.StorageFactory, data data.IData) *LoginController {
	return &LoginController{
		log:       log,
		config:    config,
		data:      data,
		redisPool: redisPool,
		sf:        sf,
	}
}

func (c *LoginController) GetLoginMethods(ctx *gin.Context) {
	sys := ctx.DefaultQuery("sys", string(constants.SYS_MEDIAHUB))
	if sys == "" {
		err := zerror.NewByMsg("请指定需要登录的系统")
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	gitlab := gitlab2.NewGitlabOAuth(
		c.config.Gitlab.Domain,
		c.config.Gitlab.ClientID,
		c.config.Gitlab.ClientSecret,
		c.log,
	)
	gitlabConf := gitlab.GetOAuth2Config(
		c.config.Gitlab.RedirectUri,
		map[string]string{"sys": sys},
		[]string{"read_user"},
	)
	gitlabAuthUrl := gitlabConf.AuthCodeURL(
		utils.GenerateRandomString(8),
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
	)

	github := github2.NewGithubOAuth(
		c.config.Github.ClientID,
		c.config.Github.ClientSecret,
		c.log,
	)
	githubConf := github.GetOAuth2Config(
		c.config.Github.RedirectUri,
		map[string]string{"sys": sys},
		[]string{"read:user", "user:email"},
	)
	githubAuthUrl := githubConf.AuthCodeURL(
		utils.GenerateRandomString(8),
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
	)

	qrcode, err := c.getWxMpQrCode()
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"gitlab":    gitlabAuthUrl,
		"github":    githubAuthUrl,
		"wx_qrcode": qrcode,
	})
}

func (c *LoginController) OAuthCallback(ctx *gin.Context) {
	code := ctx.DefaultQuery("code", "")
	if code == "" {
		err := zerror.NewByMsg("授权失败")
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	sys := ctx.DefaultQuery("sys", "")
	if sys == "" {
		err := zerror.NewByMsg("请指定需要登录的系统")
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	gitlab := gitlab2.NewGitlabOAuth(
		c.config.Gitlab.Domain,
		c.config.Gitlab.ClientID,
		c.config.Gitlab.ClientSecret,
		c.log,
	)
	gitlabConf := gitlab.GetOAuth2Config(
		c.config.Gitlab.RedirectUri,
		map[string]string{"sys": sys},
		[]string{"read_user"},
	)

	token, err := gitlabConf.Exchange(
		context.Background(),
		code,
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
	)

	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	// 获取 GitLab 用户信息
	gitlabUser, err := gitlab.GetUser(token)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	gitlabUserData := c.data.NewGitlabUserData()
	gitlabEntity, err := gitlabUserData.GetByGitlabID(gitlabUser.ID)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	if gitlabEntity == nil {
		gitlabEntity = &data.GitlabUser{
			GitlabID:  gitlabUser.ID,
			UserName:  gitlabUser.UserName,
			Name:      gitlabUser.Name,
			Email:     gitlabUser.Email,
			AvatarUrl: gitlabUser.AvatarUrl,
			CreateAt:  time.Now().Unix(),
		}

		err = gitlabUserData.AddUser(gitlabEntity)
		if err != nil {
			c.log.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}

	user, err := c.data.NewUserData().GetByID(gitlabEntity.UserID)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	hs := zjwt.NewHs(c.config.Jwt.HashKey, zjwt.HS256)
	now := time.Now()

	userClaims := &zjwt.UserClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Second * time.Duration(zjwt.EXPIRES_IN))),
		},
		UserID: user.ID,
		Name:   user.Name,
		Avatar: user.AvatarUrl,
	}

	accessToken, err := hs.Sign(userClaims)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	cookies := c.getAcrossSubdomainCookie(accessToken)
	for _, cookie := range cookies {
		http.SetCookie(ctx.Writer, cookie)
	}

	redirectUrl := c.config.InternalSystemEntry[sys]
	ctx.Redirect(http.StatusFound, redirectUrl)
}

func (c *LoginController) GithubOAuthCallback(ctx *gin.Context) {
	code := ctx.DefaultQuery("code", "")
	if code == "" {
		err := zerror.NewByMsg("授权失败")
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	sys := ctx.DefaultQuery("sys", "")
	if sys == "" {
		err := zerror.NewByMsg("请指定需要登录的系统")
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	github := github2.NewGithubOAuth(
		c.config.Github.ClientID,
		c.config.Github.ClientSecret,
		c.log,
	)

	githubConf := github.GetOAuth2Config(
		c.config.Github.RedirectUri,
		map[string]string{"sys": sys},
		[]string{"read:user", "user:email"},
	)

	token, err := githubConf.Exchange(
		context.Background(),
		code,
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
	)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	// 获取 GitHub 用户信息
	githubUser, err := github.GetUser(token)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	githubUserData := c.data.NewGithubUserData()

	githubEntity, err := githubUserData.GetByGithubID(githubUser.ID)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	if githubEntity == nil {
		githubEntity = &data.GithubUser{
			GithubID:  githubUser.ID,
			UserName:  githubUser.UserName,
			Name:      githubUser.Name,
			Email:     githubUser.Email,
			AvatarUrl: githubUser.AvatarUrl,
			CreateAt:  time.Now().Unix(),
		}

		err = githubUserData.AddUser(githubEntity)
		if err != nil {
			c.log.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}

	user, err := c.data.NewUserData().GetByID(githubEntity.UserID)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	// 生成 JWT
	hs := zjwt.NewHs(c.config.Jwt.HashKey, zjwt.HS256)
	now := time.Now()

	userClaims := &zjwt.UserClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Second * time.Duration(zjwt.EXPIRES_IN))),
		},
		UserID: user.ID,
		Name:   user.Name,
		Avatar: user.AvatarUrl,
	}

	accessToken, err := hs.Sign(userClaims)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	// 构造 Cookie，并写入响应头
	cookies := c.getAcrossSubdomainCookie(accessToken)
	for _, cookie := range cookies {
		http.SetCookie(ctx.Writer, cookie)
	}

	redirectUrl := c.config.InternalSystemEntry[sys]
	ctx.Redirect(http.StatusFound, redirectUrl)
}
