package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
	app_cache "tunnel/app-cache"
	"tunnel/constants"
	"tunnel/pkg/zerror"
)

func (c *App) getServiceName(serverID int64) string {
	serviceName := fmt.Sprintf("%s-%d", c.config.TunnelServer.AppName, serverID)
	return serviceName
}
func (c *App) DeployServer(ctx *gin.Context) {
	value, exists := ctx.Get("User.ID")
	if !exists {
		ctx.JSON(401, nil)
		return
	}
	userID := value.(int64)
	serverData := c.data.NewTunnelServerData()
	server, err := serverData.GetByUserID(userID)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	if server == nil {
		err = zerror.NewByMsg("请先创建应用")
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	apps, err := c.data.NewTunnelAppData().GetByServerID(server.ID)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	if len(apps) == 0 {
		err = zerror.NewByMsg("请先创建应用")
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	if server.Deploy == constants.SERVERDEPLOYED {
		ctx.JSON(200, gin.H{
			"service_name":  c.getServiceName(server.ID),
			"client_config": server.ClientConfig,
			"apps":          apps,
		})
		return
	}
	addrData := c.data.NewTunnelAddrData()
	addrs, err := addrData.GetAddrList(server.ID)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	addrs, err = c.assignPorts(apps, addrs)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	// 生成配置
	tunnelServerConf, tunnelClientConf, err := c.generateConfig(server, apps, addrs)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	// 缓存设置
	n, err := c.appCache.SetList(apps)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	if n > 0 {
		err = c.appCache.SetDeployStatus(app_cache.TodoDeploy)
		if err != nil {
			c.log.Error(err)
			ctx.JSON(500, nil)
			return
		}
	}
	// 域名解析
	dns := c.dnsFactory.NewDns()
	for _, app := range apps {
		if len(app.CustomDomains) == 0 {
			err = zerror.NewByMsg("请先设置域名")
			c.log.Error(err)
			ctx.JSON(500, nil)
			return
		}
		subdomain := fmt.Sprintf("%s.%s", app.CustomDomains[0], c.config.TunnelClient.RootDomain)
		exists, err := dns.CheckSubDomainExists(subdomain)
		if err != nil {
			c.log.Error(err)
			ctx.JSON(500, nil)
			return
		}
		if exists {
			continue
		}
		rr := strings.Replace(subdomain, fmt.Sprintf(".%s", c.config.AliYunDomain.RootDomain), "", 1)
		typ := "A"
		value := c.config.TunnelServer.IP
		recordID, err := dns.AddSubDomainRecord(rr, typ, value)
		if err != nil {
			c.log.Error(err)
			ctx.JSON(500, nil)
			return
		}
		dns.UpdateSubDomainRemark(recordID, "tunnel1 课程生成域名")
	}

	// 服务部署
	ports := make(map[uint32]string)
	for _, addr := range addrs {
		ports[uint32(addr.Port)] = "tcp"
	}
	serviceName, err := c.deployK8s(server.ID, tunnelServerConf, ports)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	server.Deploy = constants.SERVERDEPLOYED
	server.ClientConfig = string(tunnelClientConf)
	server.ServerConfig = string(tunnelServerConf)
	server.UpdateAt = time.Now().Unix()
	err = serverData.UpdateConfigs(server)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(500, nil)
		return
	}
	ctx.JSON(200, gin.H{
		"service_name":  serviceName,
		"client_config": server.ClientConfig,
		"apps":          apps,
	})
	return
}
