package handlers

import (
	"GoIncrease1/internal/shortener"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAddress(c *gin.Context) {
	path := c.Param("key")
	ctx := c.Request.Context()
	link, found := shortener.GetLink(ctx, path)
	if found {
		c.Redirect(http.StatusTemporaryRedirect, link)
	} else {
		c.JSON(http.StatusNotFound, nil)
	}
}
