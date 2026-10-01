package constants

type AppType string

const (
	APPHTTP AppType = "http"
	APPSSH  AppType = "ssh"
)

type ProxyType string

const (
	PROXYHTTP   ProxyType = "http"
	PROXYTCPMUX ProxyType = "tcpmux"
)

type Multiplexer string

const (
	MUXHTTPCONNECT Multiplexer = "httpconnect"
)

type PortType string

const (
	PORTTYPEBIND              PortType = "bind"
	PORTTYPEVHOSTHTTP         PortType = "vhost_http"
	PORTTYPETCPMUXHTTPCONNECT PortType = "tcpmux_http_connect"
)
const (
	SERVERDEPLOYED   = 0
	SERVERTOBEDEPLOY = 1
)
const AUTHMETHOD = "token"
