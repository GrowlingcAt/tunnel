package login

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/o1egl/govatar"
	redis2 "github.com/redis/go-redis/v9"
	"image/jpeg"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"user/data"
	"user/pkg/constants"
	"user/pkg/db/redis"
	"user/pkg/utils"
	"user/pkg/zerror"
	"user/pkg/zjwt"
	"user/services"
	"user/services/crontab"
	"user/services/crontab/proto"
)

func (c *LoginController) WxCallBackByTicket(ctx *gin.Context) {
	sys := ctx.DefaultQuery("sys", "")
	if sys == "" {
		err := zerror.NewByMsg("请指定目标系统")
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	ticket := ctx.Query("ticket")
	if sys == "" {
		err := zerror.NewByMsg("票据无效")
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	redisClient := c.redisPool.Get()
	defer c.redisPool.Put(redisClient)
	key := redis.GetKey(ticket)
	openID, err := redisClient.Get(context.Background(), key).Result()
	if err != nil && err == redis2.Nil {
		ctx.JSON(http.StatusOK, nil)
		return
	}
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	officialUserData := c.data.NewOfficialUserData()
	officialEntity, err := officialUserData.GetByOpenID(openID)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if officialEntity == nil {
		officialEntity = &data.OfficialUser{
			OpenID:   openID,
			CreateAt: time.Now().Unix(),
		}
		err = officialUserData.AddUser(officialEntity)
		if err != nil {
			c.log.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}

	userData := c.data.NewUserData()
	userEntity, err := userData.GetByID(officialEntity.UserID)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	if userEntity.Name == "" {
		nickname := utils.RandNickName(userEntity.ID)
		avatar, err := c.generateAvatar(nickname)
		if err != nil {
			c.log.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		userEntity.Name = nickname
		userEntity.AvatarUrl = avatar
		err = userData.UpdateUserInfo(userEntity)
		if err != nil {
			c.log.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}

	hs := zjwt.NewHs(c.config.Jwt.HashKey, zjwt.HS256)
	now := time.Now()
	userClaims := &zjwt.UserClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Second * time.Duration(zjwt.EXPIRES_IN))),
		},
		UserID: userEntity.ID,
		Name:   userEntity.Name,
		Avatar: userEntity.AvatarUrl,
	}
	accessToken, err := hs.Sign(userClaims)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	//构造cookie，并写入响应头
	cookies := c.getAcrossSubdomainCookie(accessToken)
	for _, cookie := range cookies {
		http.SetCookie(ctx.Writer, cookie)
	}
	redirectUrl := c.config.InternalSystemEntry[sys]
	ctx.JSON(http.StatusOK, gin.H{
		"redirect_url": redirectUrl,
	})
}

func (c *LoginController) generateAvatar(name string) (string, error) {
	img, err := govatar.GenerateForUsername(0, name)
	if err != nil {
		c.log.Error(err)
		return "", err
	}
	buffer := new(bytes.Buffer)
	err = jpeg.Encode(buffer, img, nil)
	if err != nil {
		c.log.Error(err)
		return "", err
	}
	io.NopCloser(bytes.NewReader(buffer.Bytes()))
	s := c.sf.CreateStorage()
	return s.Upload(io.NopCloser(bytes.NewReader(buffer.Bytes())), nil, fmt.Sprintf("/avatar/%s.jpg", name))
}

// 消息接收
func (c *LoginController) WxReceiveMessage(ctx *gin.Context) {
	cnf := c.config
	signature := ctx.Query("signature")
	timestamp := ctx.Query("timestamp")
	nonce := ctx.Query("nonce")
	signature1 := MakeSignature(cnf.WxOfficial.MsgToken, timestamp, nonce)
	if signature1 != signature {
		err := zerror.NewByMsg("签名验证失败")
		c.log.Error(err)
		ctx.Data(http.StatusInternalServerError, "text/plain;charset=utf-8", []byte(""))
		return
	}
	msg := &Message{}
	err := ctx.BindXML(msg)
	if err != nil {
		c.log.Error(err)
		ctx.Data(http.StatusInternalServerError, "text/plain;charset=utf-8", []byte(""))
		return
	}
	err = c.wxReceiveMessage(msg)
	if err != nil {
		c.log.Error(err)
		ctx.Data(http.StatusInternalServerError, "text/plain;charset=utf-8", []byte(""))
		return
	}
	fmt.Println(msg)

	ctx.Data(http.StatusOK, "text/plain;charset=utf-8", []byte(""))
	return
}

// 验签
func (c *LoginController) WxCheckSignature(ctx *gin.Context) {
	cnf := c.config
	signature := ctx.Query("signature")
	echostr := ctx.Query("echostr")
	timestamp := ctx.Query("timestamp")
	nonce := ctx.Query("nonce")
	signature1 := MakeSignature(cnf.WxOfficial.MsgToken, timestamp, nonce)
	if signature1 == signature {
		ctx.Data(http.StatusOK, "text/plain;charset=utf-8", []byte(echostr))
	}
	return
}

func MakeSignature(token, timesamp, nonce string) string {
	sortArr := []string{
		token, timesamp, nonce,
	}
	sort.Strings(sortArr)
	var buffer bytes.Buffer
	for _, value := range sortArr {
		buffer.WriteString(value)
	}
	sha := sha1.New()
	sha.Write(buffer.Bytes())
	return fmt.Sprintf("%x", sha.Sum(nil))
}

type wxErr struct {
	ErrCode int    `json:"errcode,omitempty"`
	ErrMsg  string `json:"errmsg,omitempty"`
}
type qrCode struct {
	wxErr
	ExpireSeconds int    `json:"expire_seconds"`
	Ticket        string `json:"ticket"`
	Url           string `json:"url"`
	QrCodeUrl     string `json:"qr_code_url"`
}

func (c *LoginController) getWxMpQrCode() (*qrCode, error) {
	accessToken, err := c.getWxAccessToken()
	if err != nil {
		c.log.Error(err)
		return nil, err
	}
	addr := "https://api.weixin.qq.com/cgi-bin/qrcode/create"
	u, _ := url.Parse(addr)
	params := url.Values{}
	params.Add("access_token", accessToken)
	u.RawQuery = params.Encode()

	payload := strings.NewReader(fmt.Sprintf(`{"expire_seconds": 180, "action_name": "QR_STR_SCENE", "action_info": {"scene": {"scene_str": "%s"}}}`, constants.QRLOGIN))
	httpclient := &http.Client{}
	req, err := http.NewRequest("POST", u.String(), payload)
	if err != nil {
		c.log.Error(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	httpRes, err := httpclient.Do(req)
	if err != nil {
		c.log.Error(err)
		return nil, err
	}
	defer httpRes.Body.Close()
	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		c.log.Error(err)
		return nil, err
	}
	qr := &qrCode{}
	err = json.Unmarshal(body, qr)
	if err != nil {
		c.log.Error(err)
		return nil, err
	}
	if qr.ErrCode != 0 {
		c.log.Error(qr.ErrMsg)
		err = zerror.NewByMsg(qr.ErrMsg)
		return nil, err
	}
	qr.Ticket = url.QueryEscape(qr.Ticket)
	qr.QrCodeUrl = fmt.Sprintf("https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=%s", qr.Ticket)
	return qr, nil

}
func (c *LoginController) getWxAccessToken() (string, error) {
	pool := crontab.NewCrontabClientPool()
	conn := pool.Get()
	defer pool.Put(conn)

	client := proto.NewTokenClient(conn)
	ctx := services.AppendBearerTokenToContext(context.Background(), c.config.DependOn.Crontab.AccessToken)
	in := &proto.TokenRequest{Id: c.config.WxOfficial.AppID}
	res, err := client.GetToken(ctx, in)
	if err != nil {
		c.log.Error(err)
		return "", zerror.NewByErr(err)
	}
	return res.AccessToken, nil

}

type CDATA struct {
	Value string `xml:",cdata"`
}
type Message struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   CDATA
	FromUserName CDATA
	CreateTime   int64
	MsgType      CDATA
	Event        CDATA
	EventKey     CDATA
	Ticket       CDATA
}

func (c *LoginController) wxReceiveMessage(msg *Message) error {
	var err error
	switch msg.MsgType.Value {
	case "event":
		err = c.wxEventHandler(msg)
		if err != nil {
			c.log.Error(err)
			return err
		}
	}
	return nil
}
func (c *LoginController) wxEventHandler(msg *Message) error {
	switch strings.ToLower(msg.Event.Value) {
	case "scan":
		if msg.EventKey.Value == string(constants.QRLOGIN) || msg.EventKey.Value == fmt.Sprintf("qrsence_%s", constants.QRLOGIN) {
			if msg.Ticket.Value == "" {
				err := zerror.NewByMsg("未能识别到有效二维码")
				c.log.Error(err)
				return err
			}
			redisClient := c.redisPool.Get()
			defer c.redisPool.Put(redisClient)

			key := redis.GetKey(msg.Ticket.Value)
			err := redisClient.SetEx(context.Background(), key, msg.FromUserName.Value, time.Minute*3).Err()
			if err != nil {
				c.log.Error(err)
				return err
			}
		}
	}
	return nil
}
