package web

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/resend/resend-go/v2"
	"golang.org/x/crypto/bcrypt"
	"lockbox-webserver/db"
	"net/http"
	"time"
)

type NewAccountContext struct {
	Email    string
	Password string
}

func (s *HTTPServer) handleGetLoginPage(c *gin.Context) {
	mainTemplateSet.WriteTemplate(c, http.StatusOK, "login", nil)
}

func (s *HTTPServer) handleLoginSubmit(c *gin.Context) {
	type RequestParams struct {
		Email      string `form:"email" binding:"required,email"`
		Password   string `form:"password" binding:"required"`
		RememberMe string `form:"remember_me" binding:"max=2"`
	}

	reqParams := RequestParams{}

	if val, exists := c.Get("new_account"); exists {
		newAccContext := val.(*NewAccountContext)
		reqParams.Email = newAccContext.Email
		reqParams.Password = newAccContext.Password
	} else {
		if err := c.ShouldBind(&reqParams); err != nil {
			// Check if we have some context for a newly created account
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	rememberMe := reqParams.RememberMe == "on"

	// Load the user record
	user, err := s.dbPool.SelectUserByEmail(c, reqParams.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			mainTemplateSet.WriteTemplate(c,
				http.StatusUnauthorized,
				"login",
				NewAlertMsg("Invalid email and/or password"))
			return
		}

		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(reqParams.Password)) != nil {
		mainTemplateSet.WriteTemplate(c,
			http.StatusUnauthorized,
			"login",
			NewAlertMsg("Invalid email and/or password"))
		return
	}

	// Build token cookies
	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 24 * time.Hour
	if rememberMe {
		refreshTokenDuration *= 30 // ~ 30 days
	}

	accessToken, err := MakeToken(JwtCustomFields{
		Type:      JwtTokenTypeAccess,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, accessTokenDuration)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	accessTokenStr, err := accessToken.SignedString(s.jwtSecretKey)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	refreshToken, err := MakeToken(JwtCustomFields{
		Type:      JwtTokenTypeRefresh,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, refreshTokenDuration)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	refreshTokenStr, err := refreshToken.SignedString(s.jwtSecretKey)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.SetCookie(
		"access_token",
		accessTokenStr,
		int(accessTokenDuration.Seconds()),
		"", "", false, true,
	)

	c.SetCookie(
		"refresh_token",
		refreshTokenStr,
		int(refreshTokenDuration.Seconds()),
		"", "", false, true,
	)

	c.Redirect(http.StatusFound, "/app/dashboard")
}

func (s *HTTPServer) handleGetLogout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "", "", false, true)
	c.SetCookie("refresh_token", "", -1, "", "", false, true)

	c.Redirect(http.StatusFound, "/app/login")
}

func (s *HTTPServer) handleGetCreateAccountPage(c *gin.Context) {
	mainTemplateSet.WriteTemplate(c, http.StatusOK, "create_account", nil)
}

func (s *HTTPServer) handleCreateAccountSubmit(c *gin.Context) {
	// Check form parameters
	email := c.PostForm("email")
	password := c.PostForm("password")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")

	if email == "" || password == "" || firstName == "" || lastName == "" {
		mainTemplateSet.WriteTemplate(c,
			http.StatusUnauthorized,
			"create_account",
			NewAlertMsg("One or more fields are empty"))
		return
	}

	_, err := s.dbPool.SelectUserByEmail(c, email)
	if err == nil {
		mainTemplateSet.WriteTemplate(c,
			http.StatusUnauthorized,
			"create_account",
			NewAlertMsg("Email already in use"))
		return
	}

	registrationToken, err := MakeToken(JwtCustomFields{
		Type:      JwtTokenTypeRegistration,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  password,
	}, 15*time.Minute)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	registrationTokenStr, err := registrationToken.SignedString(s.jwtSecretKey)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	type EmailBodyData struct {
		AlertMsg        string
		ConfirmationURL string
	}

	emailBodyData := EmailBodyData{
		ConfirmationURL: s.hostname + "/app/confirmemail/" + registrationTokenStr,
	}
	emailBody, err := mainTemplateSet.FormatTemplate("email", &emailBodyData)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	emailParams := &resend.SendEmailRequest{
		From:    "No Reply <noreply@resend.reesenorr.is>",
		To:      []string{email},
		Html:    emailBody.String(),
		Subject: "Confirm Lockbox Email",
	}

	if _, err = s.resendClient.Emails.Send(emailParams); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	mainTemplateSet.WriteTemplate(c, http.StatusOK, "check_email", nil)
}

func (s *HTTPServer) handleConfirmEmailPage(c *gin.Context) {
	token, exists := c.Params.Get("token")
	if !exists {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	registrationToken, err := ParseToken(token, s.jwtSecretKey)
	if err != nil {
		mainTemplateSet.WriteTemplate(c,
			http.StatusUnauthorized,
			"create_account",
			NewAlertMsg("Invalid registration details. Please try again later"))
		return
	}

	claims := registrationToken.CustomClaims()

	if claims.Type != JwtTokenTypeRegistration {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if _, err = s.dbPool.InsertUser(c, &db.InsertUserContext{
		Email:             claims.Email,
		PlaintextPassword: claims.Password,
		FirstName:         claims.FirstName,
		LastName:          claims.LastName,
	}); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Set("new_account", &NewAccountContext{
		Email:    claims.Email,
		Password: claims.Password,
	})

	s.handleLoginSubmit(c)
}
