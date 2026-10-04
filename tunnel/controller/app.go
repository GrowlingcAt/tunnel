package controller

import (
	"github.com/gin-gonic/gin"
	"net/url"
	"strconv"
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
	data       data.IData //负责操作数据
	dnsFactory domain_resolution.IDNSFactory
	appCache   *app_cache.AppCache //负责操作redis
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

	// 从 Gin 上下文中获取当前登录用户ID
	// 这个值通常由前面的鉴权中间件写入
	value, exists := ctx.Get("User.ID")
	if !exists {
		// 没有用户信息，说明未登录
		ctx.JSON(401, nil)
		return
	}

	// 类型断言，将用户ID转换为int64
	userID := value.(int64)

	// 创建请求参数结构体
	form := &AddForm{}

	// 将HTTP请求中的参数绑定到form结构体
	// 例如 name、type、local_ip、local_port
	err := ctx.ShouldBind(form)
	if err != nil {
		// 参数格式错误
		ctx.JSON(400, nil)
		return
	}
	// 校验应用类型
	// 当前只支持HTTP穿透和SSH穿透
	if form.Type != string(constants.APPHTTP) &&
		form.Type != string(constants.APPSSH) {

		err = zerror.NewByMsg("应用类型检查失败")
		c.log.Error(err)

		ctx.JSON(400, nil)
		return
	}

	// 根据应用类型转换成frp代理类型
	proxyType := ""

	// HTTP应用对应HTTP代理
	if form.Type == string(constants.APPHTTP) {
		proxyType = string(constants.PROXYHTTP)
	}

	// SSH应用对应TCP复用代理
	if form.Type == string(constants.APPSSH) {
		proxyType = string(constants.PROXYTCPMUX)
	}

	// ============================
	// 创建/获取 Tunnel Server
	// ============================

	// 创建TunnelServer数据访问对象
	serverData := c.data.NewTunnelServerData()

	// 根据用户ID查询是否已经存在Tunnel Server
	// 一个用户对应一个Server
	server, err := serverData.GetByUserID(userID)

	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}

	// 如果用户还没有Tunnel Server，则创建
	if server == nil {

		server = &data.TunnelServer{

			// 所属用户
			UserID: userID,

			// 根据用户ID生成认证Token
			// 后续客户端连接Tunnel Server需要使用
			Token: utils.GenerateTokenByUserID(userID),

			// 标记Server需要重新部署
			Deploy: constants.SERVERTOBEDEPLOY,

			CreateAt: time.Now().Unix(),
			UpdateAt: time.Now().Unix(),
		}

		// 保存Tunnel Server到数据库
		err = serverData.Add(server)

		if err != nil {
			c.log.Error(err)
			ctx.JSON(500, nil)
			return
		}

	} else {

		// 如果已有Server
		// 创建新应用后，需要重新生成配置
		server.Deploy = constants.SERVERTOBEDEPLOY
		server.UpdateAt = time.Now().Unix()

		// 更新Server部署状态
		err = serverData.UpdateDeployStatus(server)

		if err != nil {
			c.log.Error(err)
			ctx.JSON(500, nil)
			return
		}
	}

	// ============================
	// 创建 Tunnel App
	// ============================

	// 创建应用数据访问对象
	appData := c.data.NewTunnelAppData()

	// 构造应用数据
	app := &data.TunnelApp{

		// 用户ID
		UserID: userID,

		// 关联刚才创建/查询出的Tunnel Server
		ServerID: server.ID,

		// 应用名称
		Name: form.Name,

		// 应用类型(http/ssh)
		Type: form.Type,

		// frp代理类型
		ProxyType: proxyType,

		// 用户内网地址
		LocalIP: form.LocalIP,

		// 用户内网端口
		LocalPort: form.LocalPort,

		CreateAt: time.Now().Unix(),
		UpdateAt: time.Now().Unix(),
	}

	// 保存应用信息到数据库 tunnel_app 表
	err = appData.Add(app)

	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}

	// 返回创建成功后的应用ID
	ctx.JSON(200, gin.H{
		"app_id": app.ID,
	})

	return
}
func (c *App) Update(ctx *gin.Context) {

	// ① 获取当前登录用户ID
	value, exists := ctx.Get("User.ID")
	if !exists {
		ctx.JSON(401, nil)
		return
	}
	userID := value.(int64)

	// ② 获取并校验请求参数
	// UpdateForm包含：
	// AppID、Name、Type、LocalIP、LocalPort
	form := &UpdateForm{}
	err := ctx.ShouldBind(form)
	if err != nil {
		ctx.JSON(400, nil)
		return
	}

	// ③ 校验应用类型
	if form.Type != string(constants.APPHTTP) &&
		form.Type != string(constants.APPSSH) {
		err = zerror.NewByMsg("应用类型检查失败")
		c.log.Error(err)
		ctx.JSON(400, nil)
		return
	}

	// ④ 根据应用类型确定代理类型
	proxyType := ""
	if form.Type == string(constants.APPHTTP) {
		proxyType = string(constants.PROXYHTTP)
	}
	if form.Type == string(constants.APPSSH) {
		proxyType = string(constants.PROXYTCPMUX)
	}

	// ⑤ 根据AppID查询应用
	appData := c.data.NewTunnelAppData()
	app, err := appData.GetByID(form.AppID)

	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}

	// ⑥ 应用不存在
	if app == nil {
		err = zerror.NewByMsg("应用不存在")
		c.log.Error(err)
		ctx.JSON(404, nil)
		return
	}

	// ⑦ 权限检查
	// 防止用户修改其他用户的应用
	if app.UserID != userID {
		err = zerror.NewByMsg("应用不属于当前用户")
		c.log.Error(err)
		ctx.JSON(403, nil)
		return
	}

	// ⑧ 找到应用所属的 Tunnel Server
	serverData := c.data.NewTunnelServerData()

	server := &data.TunnelServer{
		ID: app.ServerID,

		// 非常关键：
		// 告诉系统这个Server的配置发生变化，需要重新部署
		Deploy: constants.SERVERTOBEDEPLOY,

		UpdateAt: time.Now().Unix(),
	}

	// ⑨ 更新Server的部署状态
	err = serverData.UpdateDeployStatus(server)

	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}

	// ⑩ 修改App的数据
	app.Name = form.Name
	app.Type = form.Type
	app.ProxyType = proxyType
	app.LocalIP = form.LocalIP
	app.LocalPort = form.LocalPort
	app.UpdateAt = time.Now().Unix()

	// ⑪ 保存修改后的App
	err = appData.Update(app)

	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}

	// ⑫ 返回App ID
	ctx.JSON(200, gin.H{"app_id": app.ID})
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

func (c *App) Delete(ctx *gin.Context) {

	// ① 获取当前登录用户ID
	value, exists := ctx.Get("User.ID")
	if !exists {
		ctx.JSON(401, nil)
		return
	}
	userID := value.(int64)

	// ② 获取请求参数

	appID, err := strconv.ParseInt(ctx.Query("app_id"), 10, 64)
	if err != nil {
		ctx.JSON(400, nil)
		return
	}

	//查询应用
	appData := c.data.NewTunnelAppData()
	app, err := appData.GetByID(appID)

	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}

	// ④ 应用不存在
	if app == nil {
		err = zerror.NewByMsg("应用不存在")
		c.log.Error(err)
		ctx.JSON(404, nil)
		return
	}

	// ⑤ 权限检查
	if app.UserID != userID {
		err = zerror.NewByMsg("应用不属于当前用户")
		c.log.Error(err)
		ctx.JSON(403, nil)
		return
	}

	err = appData.Delete(app.ID)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	// 删除 Redis 中对应的应用缓存
	err = c.appCache.Delete(app.EntryDomain)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	// 标记 Gateway 需要重新部署
	err = c.appCache.SetDeployStatus(app_cache.TodoDeploy)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	//删除dns解析
	dns := c.dnsFactory.NewDns()

	parseURL, err := url.Parse(app.EntryDomain)
	if err != nil {
		ctx.JSON(500, nil)
		return
	}

	err = dns.DeleteSubDomainRecord(parseURL.Host)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	// 标记 Tunnel Server 需要重新部署
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

	// ⑨ 返回 App ID
	ctx.JSON(200, gin.H{
		"app_id": app.ID,
	})
}
