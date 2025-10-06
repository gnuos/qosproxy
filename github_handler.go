package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

// 允许的文件大小，默认999GB，相当于无限制
const SIZE_LIMIT = 1024 * 1024 * 1024 * 999

var blobToRaw = regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:blob|raw)/.*$`)

// 全局变量：被阻止的内容类型
var blockedContentTypes = map[string]bool{
	"text/html":             true,
	"application/xhtml+xml": true,
	"text/xml":              true,
	"application/xml":       true,
}

// GitHubProxyHandler GitHub代理处理器
func GithubProxyHandler(c echo.Context) error {
	rawPath := processPrefix(c.Param("*"))

	if !isValidUrl(cfg.Rules, rawPath) {
		return c.String(http.StatusForbidden, "无效请求")
	}

	// 将blob链接转换为raw链接
	if blobToRaw.MatchString(rawPath) {
		rawPath = strings.Replace(rawPath, "/blob/", "/raw/", 1)
	}

	return ProxyGitHubRequest(c, rawPath)
}

// ProxyGitHubRequest 代理GitHub请求
func ProxyGitHubRequest(c echo.Context, u string) error {
	return proxyGitHubWithRedirect(c, u, 0)
}

// proxyGitHubWithRedirect 带重定向的GitHub代理请求
func proxyGitHubWithRedirect(c echo.Context, u string, redirectCount int) error {
	var err error
	const maxRedirects = 20
	if redirectCount > maxRedirects {
		return c.String(http.StatusLoopDetected, "重定向次数过多，可能存在循环重定向")
	}

	resp := sendRequest(c, u)
	if resp.Err != nil && resp.Err != io.EOF {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("server error %v", resp.Err))
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.Logger().Errorf("关闭代理响应体失败: %v\n", err.Error())
		}
	}()

	// 检查文件大小限制
	if resp.ContentLength > SIZE_LIMIT {
		return c.String(http.StatusRequestEntityTooLarge, fmt.Sprintf("文件过大，限制大小: %d MB", SIZE_LIMIT/(1024*1024)))
	}

	// 清理安全相关的头
	resp.Header.Del("Content-Security-Policy")
	resp.Header.Del("Referrer-Policy")
	resp.Header.Del("Strict-Transport-Security")

	for key := range resp.Header {
		c.Response().Header().Set(key, resp.GetHeader(key))
	}

	// 处理重定向
	if location, err := resp.Location(); err == nil {
		if isValidUrl(cfg.Rules, u) {
			return c.Redirect(resp.GetStatusCode(), "/"+location.String())
		} else {
			// 递归重定向，最大不超过20次重定向
			if err := proxyGitHubWithRedirect(c, location.String(), redirectCount+1); err != nil {
				return err
			}
			return nil
		}
	}

	if c.Request().Method == "GET" {
		// 检查并处理被阻止的内容类型
		if contentType := resp.GetContentType(); blockedContentTypes[strings.ToLower(strings.Split(contentType, ";")[0])] {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error":   "Content type not allowed",
				"message": "检测到网页类型，本服务不支持加速网页，请检查您的链接是否正确。",
			})
		}
	}

	// 获取真实域名
	realHost := c.Request().Header.Get("X-Forwarded-Host")
	if realHost == "" {
		realHost = c.Request().Host
	}
	if !strings.HasPrefix(realHost, "http://") && !strings.HasPrefix(realHost, "https://") {
		realHost = "https://" + realHost
	}

	var processedBody = resp.Body
	var processedSize int64 = 0

	// 智能处理.sh .ps1 .py文件
	if strings.HasSuffix(strings.ToLower(u), ".sh") || strings.HasSuffix(strings.ToLower(u), ".ps1") || strings.HasSuffix(strings.ToLower(u), ".py") {
		isGzipCompressed := resp.GetHeader("") == "gzip"

		processedBody, processedSize, err = ProcessSmart(resp.Body, isGzipCompressed, realHost)
		if err != nil {
			fmt.Printf("智能处理失败，回退到直接代理: %v\n", err)
			processedBody = resp.Body
			processedSize = 0
		}

		// 智能设置响应头
		if processedSize > 0 {
			c.Response().Header().Del("Content-Length")
			c.Response().Header().Del("Content-Encoding")
			c.Response().Header().Set("Transfer-Encoding", "chunked")
		}
	}

	c.Response().WriteHeader(resp.GetStatusCode())

	// 输出处理后的内容
	_, err = io.Copy(c.Response().Writer, processedBody)

	return err
}
