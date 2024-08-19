package router

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/scch94/micropagosSmtpServer/utils/handler"
	"github.com/scch94/micropagosSmtpServer/utils/middleware"
)

func SetupRouter(ctx context.Context) *gin.Engine {

	//se crea n gin router y se registran los handlers
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	//agregamos middleware globales
	router.Use(gin.Recovery())
	router.Use(middleware.GlobalMiddleware())

	h := handler.Handler{}
	router.GET("/", h.Welcome)
	router.POST("/send", h.SendEmail)

	return router
}
