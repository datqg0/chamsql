package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUUID   string
	RefreshUUID  string
	AtExpires    int64
	RtExpires    int64
}

type JWTProvider interface {
	GenerateToken(userID int64, role string, atDuration time.Duration) (*TokenDetails, error)
	ValidateToken(tokenString string) (*jwt.MapClaims, error)
}

type jwtProvider struct {
	secretKey string
}

func NewJWTProvider(secretKey string) JWTProvider {
	return &jwtProvider{
		secretKey: secretKey,
	}
}

func (p *jwtProvider) GenerateToken(userID int64, role string, atDuration time.Duration) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(atDuration).Unix()

	// Access Token (HS512)
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user_id"] = userID
	atClaims["role"] = role
	atClaims["exp"] = td.AtExpires

	at := jwt.NewWithClaims(jwt.SigningMethodHS512, atClaims)
	var err error
	td.AccessToken, err = at.SignedString([]byte(p.secretKey))
	if err != nil {
		return nil, err
	}

	return td, nil
}

func (p *jwtProvider) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(p.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}
