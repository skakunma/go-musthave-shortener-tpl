package handlers

import (
	"GoIncrease1/internal/shortener"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAddress(c *gin.Context) {
	path := c.Param("key")
	link, found := shortener.GetLink(path)
	if found {
		c.Redirect(http.StatusTemporaryRedirect, link)
	} else {
		c.JSON(http.StatusNotFound, nil)
	}
}
