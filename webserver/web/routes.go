package web

import (
	"github.com/gin-gonic/gin"
	"os"
)

func (s *HTTPServer) setupRoutes() (e *gin.Engine, err error) {
	e = gin.New()
	e.Use(gin.Recovery())
	e.Use(gin.Logger())

	apiGroup := e.Group("/api")

	accounts := make(gin.Accounts)
	accounts[os.Getenv("ESP32_USERNAME")] = os.Getenv("ESP32_PASSWORD")
	apiGroup.Use(gin.BasicAuth(accounts))

	cardsGroup := apiGroup.Group("/cards")
	cardsGroup.POST("/new", s.handleCreateCard)
	cardsGroup.POST("/use", s.handleUseCardRequest)

	appGroup := e.Group("/app")

	appGroup.GET("/login", s.handleGetLoginPage)
	appGroup.POST("/login", s.handleLoginSubmit)
	appGroup.GET("/logout", s.handleGetLogout)
	appGroup.GET("/confirmemail/:token", s.handleConfirmEmailPage)

	createAccountGroup := appGroup.Group("/createaccount")
	createAccountGroup.GET("", s.handleGetCreateAccountPage)
	createAccountGroup.POST("", s.handleCreateAccountSubmit).Use(s.createAccountRateLimitMiddleware)

	dashboardGroup := appGroup.Group("/dashboard")

	dashboardGroup.Use(s.dashboardAuthMiddleware)
	dashboardGroup.GET("", s.handleGetDashboardPage)
	dashboardGroup.POST("/incrementopens/:cardUUID", s.handleDashboardIncrementOpens)
	dashboardGroup.POST("/decrementopens/:cardUUID", s.handleDashboardDecrementOpens)
	dashboardGroup.POST("/setopens/:cardUUID", s.handleDashboardSetOpens)
	dashboardGroup.POST("/updatefriendlyname/:cardUUID", s.handleUpdateCardFriendyName)

	return
}
