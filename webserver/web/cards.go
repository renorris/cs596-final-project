package web

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func (s *HTTPServer) handleCreateCard(c *gin.Context) {
	type RequestBody struct {
		UUID uuid.UUID `json:"uuid"`
	}

	reqBody := RequestBody{}
	if err := c.BindJSON(&reqBody); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if _, err := s.dbPool.CreateCard(c, reqBody.UUID); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusNoContent)
}

func (s *HTTPServer) handleUseCardRequest(c *gin.Context) {
	type RequestBody struct {
		UUID uuid.UUID `json:"uuid"`
	}

	reqBody := RequestBody{}
	if err := c.BindJSON(&reqBody); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := s.dbPool.UseCard(c, reqBody.UUID); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusNoContent)
}
