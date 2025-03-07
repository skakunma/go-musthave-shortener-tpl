package handlers

import (
	"GoIncrease1/internal/config"
	jwtAuth "GoIncrease1/internal/jwt"
	"GoIncrease1/internal/shortener"
	"GoIncrease1/internal/storage"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type userUrl struct {
	ShortenURL  string `json:"shorten_url"`
	OriginalURL string `json:"original_url"`
}

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

func GetAddressFromUser(c *gin.Context) {
	claims, exist := c.Get("user")
	if exist != true {
		c.JSON(http.StatusUnauthorized, "You are not autorizate")
	}
	userClaims := claims.(*jwtAuth.Claims)

	ctx := c.Request.Context()

	result, err := config.Cfg.Store.GetLinksByUserId(ctx, userClaims.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, "can't find user with that id")
			return
		}
		c.JSON(http.StatusBadGateway, err)
		return
	}
	response := []userUrl{}
	for key, value := range result {
		response = append(response, userUrl{ShortenURL: key, OriginalURL: value})
	}
	c.JSON(http.StatusOK, response)
}
