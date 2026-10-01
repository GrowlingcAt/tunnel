package controller

import (
	"github.com/gin-gonic/gin"
	"time"
	app_cache "tunnel/app-cache"
	"tunnel/constants"
	"tunnel/data"
	"tunnel/pkg/config"
	domain_resolution "tunnel/pkg/domain-resolution"
	"tunnel/pkg/log"
	"tunnel/pkg/utils"
	"tunnel/pkg/zerror"
)

type App struct {
	config     *config.Config
	log        log.ILogger
	data       data.IData
	dnsFactory domain_resolution.IDNSFactory
	appCache   *app_cache.AppCache
}

func NewApp(config *config.Config, log log.ILogger, data data.IData, dnsFactory domain_resolution.IDNSFactory) *App {
	appCache := app_cache.NewAppCache(config, log, data)
	return &App{
		config:     config,
		log:        log,
		data:       data,
		dnsFactory: dnsFactory,
		appCache:   appCache,
	}
}

type AppForm struct {
	Name      string `form:"name" binding:"required"`
	Type      string `form:"type" binding:"required"`
	LocalPort int    `form:"local_port" binding:"required,number"`
	LocalIP   string `form:"local_ip" binding:"required,ipv4"`
}
type AddForm struct {
	AppForm
}
type UpdateForm struct {
	AppID int64 `form:"app_id" binding:"required,number"`
	AppForm
}

func (c *App) Add(ctx *gin.Context) {
	value, exists := ctx.Get("User.ID")
	if !exists {
		ctx.JSON(401, nil)
		return
	}
	userID := value.(int64)
	form := &AddForm{}
	err := ctx.ShouldBind(form)
	if err != nil {
		ctx.JSON(400, nil)
		return
	}
	if form.Type != string(constants.APPHTTP) && form.Type != string(constants.APPSSH) {
		err = zerror.NewByMsg("应用类型检查失败")
		c.log.Error(err)
		ctx.JSON(400, nil)
		return
	}
	proxyType := ""
	if form.Type == string(constants.APPHTTP) {
		proxyType = string(constants.PROXYHTTP)
	}
	if form.Type == string(constants.APPSSH) {
		proxyType = string(constants.PROXYTCPMUX)
	}
	// 插入数据
	serverData := c.data.NewTunnelServerData()
	server, err := serverData.GetByUserID(userID)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	if server == nil {
		server = &data.TunnelServer{
			UserID:   userID,
			Token:    utils.GenerateTokenByUserID(userID),
			Deploy:   constants.SERVERTOBEDEPLOY,
			CreateAt: time.Now().Unix(),
			UpdateAt: time.Now().Unix(),
		}
		err = serverData.Add(server)
		if err != nil {
			c.log.Error(err)
			ctx.JSON(500, nil)
			return
		}
	} else {
		server.Deploy = constants.SERVERTOBEDEPLOY
		server.UpdateAt = time.Now().Unix()
		err = serverData.UpdateDeployStatus(server)
		if err != nil {
			c.log.Error(err)
			ctx.JSON(500, nil)
			return
		}
	}
	appData := c.data.NewTunnelAppData()
	app := &data.TunnelApp{
		UserID:    userID,
		ServerID:  server.ID,
		Name:      form.Name,
		Type:      form.Type,
		ProxyType: proxyType,
		LocalIP:   form.LocalIP,
		LocalPort: form.LocalPort,
		CreateAt:  time.Now().Unix(),
		UpdateAt:  time.Now().Unix(),
	}
	err = appData.Add(app)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	ctx.JSON(200, gin.H{"app_id": app.ID})
	return
}
func (c *App) Update(ctx *gin.Context) {
	value, exists := ctx.Get("User.ID")
	if !exists {
		ctx.JSON(401, nil)
		return
	}
	userID := value.(int64)
	form := &UpdateForm{}
	err := ctx.ShouldBind(form)
	if err != nil {
		ctx.JSON(400, nil)
		return
	}
	if form.Type != string(constants.APPHTTP) && form.Type != string(constants.APPSSH) {
		err = zerror.NewByMsg("应用类型检查失败")
		c.log.Error(err)
		ctx.JSON(400, nil)
		return
	}
	proxyType := ""
	if form.Type == string(constants.APPHTTP) {
		proxyType = string(constants.PROXYHTTP)
	}
	if form.Type == string(constants.APPSSH) {
		proxyType = string(constants.PROXYTCPMUX)
	}
	appData := c.data.NewTunnelAppData()
	app, err := appData.GetByID(form.AppID)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	if app == nil {
		err = zerror.NewByMsg("应用不存在")
		c.log.Error(err)
		ctx.JSON(404, nil)
		return
	}
	if app.UserID != userID {
		err = zerror.NewByMsg("应用不属于当前用户")
		c.log.Error(err)
		ctx.JSON(403, nil)
		return
	}

	serverData := c.data.NewTunnelServerData()
	server := &data.TunnelServer{
		ID:       app.ServerID,
		Deploy:   constants.SERVERTOBEDEPLOY,
		UpdateAt: time.Now().Unix(),
	}
	err = serverData.UpdateDeployStatus(server)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}

	app.Name = form.Name
	app.Type = form.Type
	app.ProxyType = proxyType
	app.LocalIP = form.LocalIP
	app.LocalPort = form.LocalPort
	app.UpdateAt = time.Now().Unix()
	err = appData.Update(app)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	ctx.JSON(200, gin.H{"app_id": app.ID})
	return
}
func (c *App) List(ctx *gin.Context) {
	value, exists := ctx.Get("User.ID")
	if !exists {
		ctx.JSON(401, nil)
		return
	}
	userID := value.(int64)
	appData := c.data.NewTunnelAppData()
	list, err := appData.GetByUserID(userID)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	ctx.JSON(200, list)
	return
}
