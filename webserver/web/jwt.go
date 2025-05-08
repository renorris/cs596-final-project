package web

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

const issuer = "lockbox"

type JwtToken struct {
	*jwt.Token
}

type JwtCustomClaims struct {
	jwt.RegisteredClaims
	JwtCustomFields
}

type JwtCustomFields struct {
	Type      JwtTokenType `json:"token_type"`
	Email     string       `json:"email"`
	FirstName string       `json:"first_name"`
	LastName  string       `json:"last_name"`
	Password  string       `json:"password,omitempty"`
}

type JwtTokenType string

const (
	JwtTokenTypeAccess       JwtTokenType = "access"
	JwtTokenTypeRefresh      JwtTokenType = "refresh"
	JwtTokenTypeRegistration JwtTokenType = "registration"
)

func ParseToken(tokenStr string, secretKey []byte) (token *JwtToken, err error) {
	jwtToken, err := jwt.ParseWithClaims(tokenStr, &JwtCustomClaims{}, func(t *jwt.Token) (any, error) {
		return secretKey, nil
	},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(issuer),
	)
	if err != nil {
		return
	}

	token = &JwtToken{jwtToken}

	return
}

func MakeToken(fields JwtCustomFields, validFor time.Duration) (token *JwtToken, err error) {
	now := time.Now()

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &JwtCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: &jwt.NumericDate{Time: now.Add(validFor)},
			NotBefore: &jwt.NumericDate{Time: now.Add(-30 * time.Second)},
			IssuedAt:  &jwt.NumericDate{Time: now},
			ID:        uuid.New().String(),
		},
		JwtCustomFields: fields,
	})

	token = &JwtToken{jwtToken}

	return
}

func (t *JwtToken) CustomClaims() *JwtCustomClaims {
	return (t.Claims).(*JwtCustomClaims)
}
