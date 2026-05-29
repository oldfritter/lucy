package base

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var (
	JWT = JwtServices{
		SecretKey: "LIfyf5vFTYT;sad*^sdsfs886gKH",
		Issure:    "Happycoding",
		Tag:       "web",
	}
)

type AuthCustomClaims struct {
	UserId int
	Tag    string
	jwt.RegisteredClaims
}

type JwtServices struct {
	SecretKey string
	Issure    string
	Tag       string
}

type GenerateTokenParam interface {
	GetUserId() int
	GetExpiredAt() time.Time
}

func (js *JwtServices) GenerateToken(user GenerateTokenParam) (string, error) {
	claims := &AuthCustomClaims{
		UserId: user.GetUserId(),
		Tag:    js.Tag,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(user.GetExpiredAt()),
			Issuer:    js.Issure,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(js.SecretKey))
	if err != nil {
		return "", err
	}
	return t, nil
}

func (js *JwtServices) GenerateClaim(user GenerateTokenParam) jwt.MapClaims {
	return jwt.MapClaims{
		"UserId": float64(user.GetUserId()),
		"Tag":    js.Tag,
		"exp":    user.GetExpiredAt().Unix(),
		"iat":    time.Now().Unix(),
		"iss":    js.Issure,
	}
}

func (js *JwtServices) ValidateToken(encodedToken string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(token *jwt.Token) (any, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("invalid token : %v", token.Header["alg"])
		}
		return []byte(js.SecretKey), nil
	})
}

func GetClaim(c echo.Context) (acc AuthCustomClaims, err error) {
	if c.Get("Claim") == nil {
		acc = AuthCustomClaims{
			UserId: 0,
			Tag:    "",
		}
		return
	}
	Claims := c.Get("Claim").(jwt.MapClaims)
	acc = AuthCustomClaims{
		UserId: int(Claims["UserId"].(float64)),
		Tag:    Claims["Tag"].(string),
	}
	return
}
