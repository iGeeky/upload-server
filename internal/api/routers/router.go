package routers

import (
	"fmt"
	"strings"

	"github.com/iGeeky/upload-server/configs"
	"github.com/iGeeky/upload-server/internal/api/controller"
	"github.com/iGeeky/upload-server/internal/api/middleware"

	"github.com/iGeeky/open-account/pkg/baselib/ginplus"
	"github.com/iGeeky/open-account/pkg/baselib/log"

	"github.com/gin-gonic/gin"
)

var (
	accessControlAllowOrigin  = "*"
	accessControlAllowHeaders = `Origin, Referer, User-Agent, X-Requested-With, Content-Type, Accept, Range`
)

// APIOption 处理Options请求.
func APIOption(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", accessControlAllowOrigin)
	c.Writer.Header().Set("Access-Control-Allow-Headers", accessControlAllowHeaders)
	c.Writer.Header().Set("Connection", "true")
	log.Infof("OPTIONS")
	c.String(200, "OK")
}

// getAPIRouters 所有router注册点
func getAPIRouters() (routers []RouterInfo) {
	routers = []RouterInfo{
		{HTTP_GET, "/upload/ping", false, TokenNone, controller.APIPing},
		{HTTP_POST, "/upload/simple", true, TokenNone, controller.UploadSimple},
		{HTTP_POST, "/upload/url", true, TokenNone, controller.UploadURLSimple},
		{HTTP_GET, "/upload/check_exist", true, TokenNone, controller.CheckExist},
		{HTTP_POST, "/upload/chunk/init", true, TokenNone, controller.UploadChunkInit},
		{HTTP_POST, "/upload/chunk/upload", true, TokenNone, controller.UploadChunkUpload},
	}
	return
}

func routerOption(g *gin.RouterGroup, optionsRouterURLs map[string]bool, URL string) {
	_, exist := optionsRouterURLs[URL]
	if !exist {
		g.OPTIONS(URL, APIOption)
		optionsRouterURLs[URL] = true
	}
}

func initOneRouter(r *gin.Engine, ver string, routers []RouterInfo) {
	g := r.Group(ver)
	optionsRouterURLs := make(map[string]bool)
	for _, routerInfo := range routers {
		url := ver + routerInfo.URL
		switch routerInfo.Op {
		case HTTP_GET:
			g.GET(routerInfo.URL, routerInfo.Handler)
			routerOption(g, optionsRouterURLs, routerInfo.URL)
		case HTTP_POST:
			g.POST(routerInfo.URL, routerInfo.Handler)
			routerOption(g, optionsRouterURLs, routerInfo.URL)
		case HTTP_PUT:
			g.PUT(routerInfo.URL, routerInfo.Handler)
			routerOption(g, optionsRouterURLs, routerInfo.URL)
		case HTTP_DELETE:
			g.DELETE(routerInfo.URL, routerInfo.Handler)
			routerOption(g, optionsRouterURLs, routerInfo.URL)
		case HTTP_OPTION:
			g.OPTIONS(routerInfo.URL, routerInfo.Handler)
		default:
			log.Errorf("Unknown http method: %d", routerInfo.Op)
		}

		// Check Sign map 添加
		ginplus.SignConfig.SignUrls[url] = routerInfo.CheckSign

	}
}

func getAllowHeaders(c *gin.Context) (allowHeaders string) {
	cfg := &configs.Config.CORS
	if cfg.AllowHeaders == "all" {
		var headerkeys []string
		for k := range c.Request.Header {
			headerkeys = append(headerkeys, k)
		}
		allowHeaders = strings.Join(headerkeys, ", ")
		if allowHeaders != "" {
			allowHeaders = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", allowHeaders)
		} else {
			allowHeaders = "access-control-allow-origin, access-control-allow-headers"
		}
		return
	} else {
		allowHeaders = cfg.AllowHeaders
	}
	return allowHeaders
}

// CorsHeader 自动添加 Cors相关头
func CorsHeader(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	if origin != "" {
		cfg := &configs.Config.CORS
		c.Header("Access-Control-Allow-Origin", cfg.AllowOrigins)
		c.Header("Access-Control-Allow-Methods", cfg.AllowMethods)
		c.Header("Access-Control-Allow-Credentials", fmt.Sprintf("%v", cfg.AllowCredentials))
		c.Header("Access-Control-Allow-Headers", getAllowHeaders(c))
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("content-type", "application/json")
	}
}

// InitRouter 初始化Router
func InitRouter(config *configs.ServerConfig) *gin.Engine {
	r := gin.Default()

	r.Use(gin.Recovery())
	r.Use(CorsHeader)
	r.Use(middleware.PanicHandler)
	r.Use(ginplus.SignCheck)

	initOneRouter(r, "/v1", getAPIRouters())
	return r
}
