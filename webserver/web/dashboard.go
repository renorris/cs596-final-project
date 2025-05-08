package web

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"lockbox-webserver/db"
	"net/http"
	"strconv"
)

func (s *HTTPServer) handleGetDashboardPage(c *gin.Context) {

	token := s.getAccessTokenFromContext(c)

	user, err := s.dbPool.SelectUserByEmail(c, token.CustomClaims().Email)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	cards, err := s.dbPool.ListCards(c)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	type DashboardPageData struct {
		AlertMsg string
		User     *db.User
		Cards    []*db.Card
	}

	pageData := DashboardPageData{
		User:  user,
		Cards: cards,
	}

	mainTemplateSet.WriteTemplate(c, http.StatusOK, "dashboard", &pageData)
}

func (s *HTTPServer) handleDashboardIncrementOpens(c *gin.Context) {
	cardUUIDStr, exists := c.Params.Get("cardUUID")
	if !exists {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	cardUUID, err := uuid.Parse(cardUUIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err = s.dbPool.AddCardOpens(c, cardUUID, 1); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusFound, "/app/dashboard")

	return
}

func (s *HTTPServer) handleDashboardDecrementOpens(c *gin.Context) {
	cardUUIDStr, exists := c.Params.Get("cardUUID")
	if !exists {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	cardUUID, err := uuid.Parse(cardUUIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err = s.dbPool.AddCardOpens(c, cardUUID, -1); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusFound, "/app/dashboard")

	return
}

func (s *HTTPServer) handleDashboardSetOpens(c *gin.Context) {
	cardUUIDStr, exists := c.Params.Get("cardUUID")
	if !exists {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	cardUUID, err := uuid.Parse(cardUUIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	numOpensStr := c.PostForm("num")
	if numOpensStr == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	numOpens, err := strconv.Atoi(numOpensStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err = s.dbPool.SetCardOpens(c, cardUUID, numOpens); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusFound, "/app/dashboard")

	return
}

func (s *HTTPServer) handleUpdateCardFriendyName(c *gin.Context) {
	cardUUIDStr, exists := c.Params.Get("cardUUID")
	if !exists {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	cardUUID, err := uuid.Parse(cardUUIDStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	newName := c.PostForm("name")
	if newName == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err = s.dbPool.UpdateCardFriendlyName(c, cardUUID, newName); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusFound, "/app/dashboard")

	return
}
