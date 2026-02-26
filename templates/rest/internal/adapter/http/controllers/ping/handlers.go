package ping

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct {
}

func (c *Controller) Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func NewController() *Controller {
	return &Controller{}
}
