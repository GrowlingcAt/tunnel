package constants

type InternalSys string

const (
	// mediahub系统
	SYS_MEDIAHUB InternalSys = "mediahub"
	// aichat系统
	SYS_AICHAT InternalSys = "aichat"
)

const (
	GITLABAUTHURL  = "/oauth/authorize"
	GITLABTOKENURL = "/oauth/token"
	GITLABUSERURL  = "/api/v4/user"
)

type UserState int

const (
	Active   UserState = 1
	InActive UserState = 0
)

type QrScene string

const (
	QRLOGIN QrScene = "qr_login"
)
