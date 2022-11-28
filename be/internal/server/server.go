package server

import (
	"neko-question-box-be/internal/api"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitServer() *gin.Engine {
	r := gin.Default()

	r.Use(func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", ctx.Request.Header.Get("Origin"))
		ctx.Header("Access-Control-Allow-Headers", "content-type, authorization")
		if ctx.Request.Method == http.MethodOptions {
			ctx.Status(http.StatusOK)
			return
		}
		ctx.Next()
	})
	api.OtherHandlers().Install(r.Group(""))
	qGroup := r.Group("question")
	qGroup.Use()
	api.QuestionHandlers().Install(qGroup)
	return r
}
