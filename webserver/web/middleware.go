package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func (s *HTTPServer) getAccessTokenFromContext(c *gin.Context) (token *JwtToken) {
	val, exists := c.Get("access_token")
	if !exists {
		panic("access token does not exist")
	}

	return val.(*JwtToken)
}

func (s *HTTPServer) dashboardAuthMiddleware(c *gin.Context) {
	accessTokenStr, err := c.Cookie("access_token")
	accessTokenExists := err == nil

	refreshTokenStr, err := c.Cookie("refresh_token")
	refreshTokenExists := err == nil

	if !accessTokenExists && refreshTokenExists {
		// Attempt to refresh session
		refreshToken, err := ParseToken(refreshTokenStr, s.jwtSecretKey)
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			return
		}

		if refreshToken.CustomClaims().Type != JwtTokenTypeRefresh {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		newAccessTokenValidFor := 15 * time.Minute

		claims := refreshToken.CustomClaims()
		newAccessToken, err := MakeToken(JwtCustomFields{
			Type:      JwtTokenTypeAccess,
			Email:     claims.Email,
			FirstName: claims.FirstName,
			LastName:  claims.LastName,
		}, newAccessTokenValidFor)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		newAccessTokenStr, err := newAccessToken.SignedString(s.jwtSecretKey)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.SetCookie(
			"access_token",
			newAccessTokenStr,
			int(newAccessTokenValidFor.Seconds()),
			"", "", false, true,
		)

		accessTokenStr = newAccessTokenStr
	}

	// Parse access token
	accessToken, err := ParseToken(accessTokenStr, s.jwtSecretKey)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if accessToken.CustomClaims().Type != JwtTokenTypeAccess {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Set("access_token", accessToken)

	c.Next()
}

func (s *HTTPServer) createAccountRateLimitMiddleware(c *gin.Context) {
	_, _, _, ok, err := s.createAccountLimiter.Take(c, c.RemoteIP())
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if !ok {
		c.AbortWithStatus(http.StatusTooManyRequests)
		return
	}

	c.Next()
}
