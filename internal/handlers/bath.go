package handlers

import (
	"GoIncrease1/internal/shortener"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type infoAboutURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type infoAboutURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortLink     string `json:"short_url"`
}

func Batch(c *gin.Context) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		c.JSON(http.StatusBadRequest, "Content-Type must be application/json")
		return
	}

	var (
		buf      bytes.Buffer
		links    []infoAboutURL
		response []infoAboutURLResponse
	)

	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Have info in body?")
		return
	}
	err = json.Unmarshal(buf.Bytes(), &links)
	if err != nil {
		c.JSON(http.StatusBadRequest, "JSON is not correctly")
		return
	}
	ctx := c.Request.Context()

	for _, link := range links {
		shorten, err := shortener.AddLink(ctx, link.OriginalURL, link.CorrelationID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Problem service")
			return
		}
		response = append(response, infoAboutURLResponse{
			CorrelationID: link.CorrelationID,
			ShortLink:     shorten,
		})
	}

	c.JSON(http.StatusCreated, response)
}
