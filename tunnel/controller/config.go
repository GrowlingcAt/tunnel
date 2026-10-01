package controller

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"time"
	"tunnel/constants"
	"tunnel/data"
	"tunnel/pkg/utils"
)

type serverConfig struct {
	BindPort              int `yaml:"bindPort,omitempty"`
	TcpmuxHTTPConnectPort int `yaml:"tcpmuxHTTPConnectPort,omitempty"`
	VhostHTTPPort         int `yaml:"vhostHTTPPort,omitempty"`
	Auth                  struct {
		Method string `yaml:"method,omitempty"`
		Token  string `yaml:"token,omitempty"`
	}
}
type clientConfig struct {
	ServerAddr string `yaml:"serverAddr,omitempty"`
	ServerPort int    `yaml:"serverPort,omitempty"`
	Auth       struct {
		Method string `yaml:"method,omitempty"`
		Token  string `yaml:"token,omitempty"`
	}
	Proxies []*proxy `yaml:"proxies,omitempty"`
}
type proxy struct {
	Name          string                `yaml:"name,omitempty"`
	Type          constants.ProxyType   `yaml:"type,omitempty"`
	Multiplexer   constants.Multiplexer `yaml:"multiplexer,omitempty"`
	LocalIP       string                `yaml:"localIP,omitempty"`
	LocalPort     int                   `yaml:"localPort,omitempty"`
	CustomDomains []string              `yaml:"customDomains,omitempty"`
}

func (c *App) assignPorts(apps []*data.TunnelApp, addrs []*data.TunnelAddr) ([]*data.TunnelAddr, error) {
	appTypes := make(map[string]struct{})
	for _, item := range apps {
		appTypes[item.Type] = struct{}{}
	}
	addrPortTypes := make(map[string]int)
	for _, item := range addrs {
		addrPortTypes[item.PortType] = item.Port
	}
	toAssignPorts := make([]*data.TunnelAddr, 0)
	if _, ok := addrPortTypes[string(constants.PORTTYPEBIND)]; !ok {
		p := &data.TunnelAddr{
			ServerID: apps[0].ServerID,
			IP:       c.config.TunnelServer.IP,
			PortType: string(constants.PORTTYPEBIND),
			CreateAt: time.Now().Unix(),
			UpdateAt: time.Now().Unix(),
		}
		toAssignPorts = append(toAssignPorts, p)
	}
	if _, ok := appTypes[string(constants.APPHTTP)]; ok {
		if _, ok := addrPortTypes[string(constants.PORTTYPEVHOSTHTTP)]; !ok {
			p := &data.TunnelAddr{
				ServerID: apps[0].ServerID,
				IP:       c.config.TunnelServer.IP,
				PortType: string(constants.PORTTYPEVHOSTHTTP),
				CreateAt: time.Now().Unix(),
				UpdateAt: time.Now().Unix(),
			}
			toAssignPorts = append(toAssignPorts, p)
		}
	}
	if _, ok := appTypes[string(constants.APPSSH)]; ok {
		if _, ok := addrPortTypes[string(constants.PORTTYPETCPMUXHTTPCONNECT)]; !ok {
			p := &data.TunnelAddr{
				ServerID: apps[0].ServerID,
				IP:       c.config.TunnelServer.IP,
				PortType: string(constants.PORTTYPETCPMUXHTTPCONNECT),
				CreateAt: time.Now().Unix(),
				UpdateAt: time.Now().Unix(),
			}
			toAssignPorts = append(toAssignPorts, p)
		}
	}
	if len(toAssignPorts) > 0 {
		addrData := c.data.NewTunnelAddrData()
		err := addrData.AssignPorts(toAssignPorts, c.config.TunnelServer.IP, c.config.TunnelServer.MinPort, c.config.TunnelServer.MaxPort)
		if err != nil {
			c.log.Error(err)
			return nil, err
		}
	}
	addrs = append(addrs, toAssignPorts...)
	return addrs, nil
}

func (c *App) generateConfig(server *data.TunnelServer, apps []*data.TunnelApp, addrs []*data.TunnelAddr) (tunnelServerConf, tunnelClientConf []byte, err error) {
	sConfig := &serverConfig{
		Auth: struct {
			Method string `yaml:"method,omitempty"`
			Token  string `yaml:"token,omitempty"`
		}{Method: constants.AUTHMETHOD, Token: server.Token},
	}
	addrMp := make(map[string]*data.TunnelAddr)
	for _, addr := range addrs {
		addrMp[addr.PortType] = addr
		switch constants.PortType(addr.PortType) {
		case constants.PORTTYPEBIND:
			sConfig.BindPort = addr.Port
		case constants.PORTTYPETCPMUXHTTPCONNECT:
			sConfig.TcpmuxHTTPConnectPort = addr.Port
		case constants.PORTTYPEVHOSTHTTP:
			sConfig.VhostHTTPPort = addr.Port
		}
	}
	cConfig := &clientConfig{
		ServerAddr: c.config.TunnelServer.IP,
		ServerPort: sConfig.BindPort,
		Auth: struct {
			Method string `yaml:"method,omitempty"`
			Token  string `yaml:"token,omitempty"`
		}{
			Method: constants.AUTHMETHOD,
			Token:  server.Token,
		},
		Proxies: make([]*proxy, 0),
	}
	rootDomain := c.config.TunnelClient.RootDomain
	appData := c.data.NewTunnelAppData()
	for _, app := range apps {
		if len(app.CustomDomains) == 0 {
			domain := fmt.Sprintf("%s-%s", utils.ToBase36(server.ID), utils.ToBase36(app.ID))
			app.CustomDomains = []string{domain}
		}
		switch constants.AppType(app.Type) {
		case constants.APPHTTP:
			app.EntryDomain = fmt.Sprintf("https://%s.%s", app.CustomDomains[0], rootDomain)
			app.EntryPort = addrMp[string(constants.PORTTYPEVHOSTHTTP)].Port
		case constants.APPSSH:
			app.EntryDomain = fmt.Sprintf("%s.%s", app.CustomDomains[0], rootDomain)
			app.EntryPort = 22
			app.Proxy = addrMp[string(constants.PORTTYPEBIND)].IP
			app.ProxyPort = addrMp[string(constants.PORTTYPETCPMUXHTTPCONNECT)].Port
		}
		app.UpdateAt = time.Now().Unix()
		err = appData.Update(app)
		if err != nil {
			c.log.Error(err)
			return nil, nil, err
		}
		p := &proxy{
			Name:          app.Name,
			Type:          constants.ProxyType(app.Type),
			CustomDomains: []string{fmt.Sprintf("%s.%s", app.CustomDomains[0], rootDomain)},
			LocalIP:       app.LocalIP,
			LocalPort:     app.LocalPort,
		}
		if app.Type == string(constants.APPSSH) {
			p.Multiplexer = constants.MUXHTTPCONNECT
		}
		cConfig.Proxies = append(cConfig.Proxies, p)
	}
	frpsBytes, err := yaml.Marshal(sConfig)
	if err != nil {
		c.log.Error(err)
		return nil, nil, err
	}
	frpcBytes, err := yaml.Marshal(cConfig)
	if err != nil {
		c.log.Error(err)
		return nil, nil, err
	}
	return frpsBytes, frpcBytes, nil
}
