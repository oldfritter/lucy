package middleware

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo/v4"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/lib/kv"
	"github.com/oldfritter/lucy/util"
)

var (
	LoginTokenKey = fmt.Sprintf("%s:login:token:web", base.RedisNamespace)
)

func Auth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return treatToken(c, func(token string, fn func(string, any, *jwt.Token, echo.HandlerFunc) error) error {
				jwttoken, err := base.JWT.ValidateToken(token)
				if err != nil || jwttoken == nil {
					return util.BuildError("1004")
				}
				jtc := jwttoken.Claims.(jwt.MapClaims)
				return fn(
					func() string {
						return LoginTokenKey
					}(),
					jtc["UserId"],
					jwttoken,
					next,
				)
			})
		}
	}
}

func treatToken(c echo.Context, fn func(string, func(string, any, *jwt.Token, echo.HandlerFunc) error) error) error {
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		token = c.QueryParam("authorization")
		if token == "" {
			token = c.FormValue("authorization")
		}
	}
	if len(token) < 1 {
		return util.BuildError("1004")
	}
	return fn(token, func(tokenKey string, userId any, jwttoken *jwt.Token, next echo.HandlerFunc) error {
		key := fmt.Sprintf("%s:%v", tokenKey, userId)
		cacheRedis := kv.GetRedisConn("cache")
		defer cacheRedis.Close()
		if t, _ := redis.String(cacheRedis.Do("GET", key)); t == "" || t == token {
			c.Set("Claim", jwttoken.Claims)
			return next(c)
		} else {
			return util.BuildError("1004")
		}
	})
}
