package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/scch94/ins_log"
)

type Handler struct {
}

func (h *Handler) Welcome(c *gin.Context) {
	ctx := c.Request.Context()

	ctx = ins_log.SetPackageNameInContext(ctx, "handler")

	ins_log.Infof(ctx, "starting handler welcome")

	c.JSON(http.StatusOK, "bienvenidos")
}
