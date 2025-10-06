package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	glog "github.com/labstack/gommon/log"
	"golang.org/x/net/http2"
)

//go:embed public
var assets embed.FS

func startWeb() {
	e := echo.New()
	h2s := &http2.Server{}

	if enableDebug {
		e.Use(middleware.Logger())
		e.Logger.SetLevel(glog.DEBUG)
	}

	e.Use(middleware.Recover())

	fsys, err := fs.Sub(assets, "public")
	if err != nil {
		log.Fatal(err)
	}

	// 用rice包装一个内存里的文件服务，会把指定的目录按正常的文件系统打包到一起
	assetHandler := http.FileServer(http.FS(fsys))

	// 默认寻找静态资源里面的index.html文件
	e.HEAD("/", echo.WrapHandler(assetHandler))
	e.GET("/", func(c echo.Context) error {
		q := c.QueryParam("q")

		if q != "" {
			if _, err := url.Parse(q); err == nil {
				return c.Redirect(http.StatusPermanentRedirect, q)
			}
		}

		return echo.WrapHandler(assetHandler)(c)
	})

	e.HEAD("/favicon.ico", echo.WrapHandler(assetHandler))
	e.GET("/favicon.ico", echo.WrapHandler(assetHandler))

	e.GET("/time", timeHandler)

	g := e.Group("/gh")
	g.Any("*", GithubProxyHandler)

	e.Logger.Fatal(e.StartH2CServer(bindAddr, h2s))
}

func timeHandler(c echo.Context) error {
	return c.String(http.StatusOK, time.Now().String())
}
