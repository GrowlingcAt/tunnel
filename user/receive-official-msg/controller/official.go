package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func WxReceiveMessage(ctx *gin.Context) {
	body, _ := io.ReadAll(ctx.Request.Body)
	fmt.Println("AAAAAAAAAAAAAAAAA")
	fmt.Println(ctx.Request.URL.RawQuery)
	fmt.Println(fmt.Sprintf("%s", body))
	fmt.Println("AAAAAAAAAAAAAAAAA")
	ctx.Data(http.StatusOK, "text/plain;charset=utf-8", []byte(""))
	return
}

func WxCheckSignature(ctx *gin.Context) {
	fmt.Println("BBBBBBBBBBBBBBBBB")
	fmt.Println(ctx.Request.URL.RawQuery)
	fmt.Println("BBBBBBBBBBBBBBBBB")
	echoStr := ctx.Query("echostr")
	ctx.Data(http.StatusOK, "text/plain;charset=utf-8", []byte(echoStr))
	return
}
