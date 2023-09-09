package routes

import (
	"app/controllers"
	"app/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Load(router *gin.Engine) {
	//router.Use(gin.Recovery()) // error handle

	router.Use(middleware.RecoverMiddleware()) // error handle

	router.GET("hello", func(ctx *gin.Context) {
		controller := controllers.HelloController{}
		controller.Index(ctx)
	})

	router.GET("ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

}
