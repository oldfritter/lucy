package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/newrelic/go-agent/v3/integrations/nrecho-v4"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/oldfritter/validator/v10"

	"github.com/oldfritter/lucy/base"
	"github.com/oldfritter/lucy/initialize"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/lib/helper"
	"github.com/oldfritter/lucy/lib/kv"
	_ "github.com/oldfritter/lucy/lib/storage/oss"
	"github.com/oldfritter/lucy/service/api/route"
	"github.com/oldfritter/lucy/util"
)

func main() {
	defer kv.CloseRedisPools()

	initialize.MigrateDB()
	defer db.CloseDB()

	e := echo.New()
	if licenseKey := base.GetConfig().Get("newrelic.license_key", ""); licenseKey != "" {
		if app, err := newrelic.NewApplication(
			newrelic.ConfigAppName("lucy-api"),
			newrelic.ConfigLicense(licenseKey),
		); err != nil {
			os.Exit(1)
		} else {
			e.Use(nrecho.Middleware(app))
		}
	}
	e.Use(
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
			AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		}),
		middleware.Secure(),
		middleware.Recover(),
		middleware.Logger(),
	)
	e.Validator = &helper.CustomValidator{Validator: validator.New()}
	e.HTTPErrorHandler = helper.CustomHTTPErrorHandler
	e.Logger.SetOutput(util.GetLogFile())
	e.DisableHTTP2 = true
	e.HideBanner = true

	route.SetV1Interface(e)

	// 静态文件服务
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/app")
	})
	e.File("/app", "public/index.html")
	app := e.Group("/app")
	app.Static("/assets", "public/assets")
	app.GET("/*", func(c echo.Context) error {
		return c.File("public/index.html")
	})

	var err error
	go func() {
		if err = e.Start(":4000"); err != nil {
			log.Fatal("shutting down the server")
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = e.Shutdown(ctx); err != nil {
		log.Println(err)
	}
}
