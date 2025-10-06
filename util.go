package main

import (
	"log/slog"
	"regexp"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/labstack/echo/v4"
)

func sendRequest(c echo.Context, u string) *req.Response {
	r := globalClient.R()

	r.SetURL(u)
	r.Method = c.Request().Method
	r.SetBody(c.Request().Body)

	// 重定向之后的Location里面链接可能带有查询参数，Go的框架一般用QueryString表示URL里面问号之后的键值对
	// 这些查询参数需要传给客户端用于发到目标链接，内容可能会涉及Token之类的
	if queryStr := c.QueryString(); len(queryStr) > 0 {
		r.SetQueryString(queryStr)
	}

	// 复制请求头
	for key := range c.Request().Header {
		r.SetHeader(key, c.Request().Header.Get(key))
	}

	r.Headers.Del("Host")

	return r.Do()
}

func processPrefix(p string) string {
	rawPath := p

	for strings.HasPrefix(rawPath, "/") {
		rawPath = strings.TrimPrefix(rawPath, "/")
	}

	// 自动补全协议头
	if !strings.HasPrefix(rawPath, "https://") {
		if strings.HasPrefix(rawPath, "http:/") || strings.HasPrefix(rawPath, "https:/") {
			rawPath = strings.Replace(rawPath, "http:/", "", 1)
			rawPath = strings.Replace(rawPath, "https:/", "", 1)
		}
		rawPath, _ = strings.CutPrefix(rawPath, "http://")
		rawPath = "https://" + rawPath
	}

	return rawPath
}

func isValidUrl(rules []string, uri string) bool {
	for _, rule := range rules {
		m, err := regexp.MatchString(rule, uri)
		if err != nil {
			slog.Default().Error(err.Error())
			continue
		}

		if m {
			return m
		}
	}

	return false
}

// transformURL URL转换函数
func transformURL(url, host string) string {
	if strings.Contains(url, host) {
		return url
	}

	if strings.HasPrefix(url, "http://") {
		url = "https" + url[4:]
	} else if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "//") {
		url = "https://" + url
	}

	// 确保 host 有协议头
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	host = strings.TrimSuffix(host, "/")

	return host + "/" + url
}
