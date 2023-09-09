package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type HelloController struct {
}

func (t *HelloController) Index(c *gin.Context) {
	//time.Sleep(10 * time.Second)
	c.String(http.StatusOK, "hello world"+c.Query("name"))
}
