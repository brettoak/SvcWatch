package api

import (
	"SvcWatch/internal/config"
	"SvcWatch/internal/controller"
	"SvcWatch/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter initializes and configures the Gin API router.
func SetupRouter(ctrl *controller.MonitorController, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Enable CORS for all origins (fixes Swagger UI doc.json fetch issues)
	router.Use(cors.Default())

	v1 := router.Group("/api/v1/sev")
	{
		// Public routes
		v1.GET("/ping", ctrl.PingHandler)

		// Protected routes require token authentication
		private := v1.Group("")
		private.Use(middleware.TokenAuthMiddleware(cfg.Auth.PassportURL))
		{
			// Example permission required to view stats
			private.GET("/stats", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:stats"), ctrl.StatsHandler)
			// Time series stats endpoint
			private.GET("/stats/timeseries", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:stats"), ctrl.TimeSeriesHandler)
			// Top Paths endpoint
			private.GET("/stats/top-paths", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:stats"), ctrl.TopPathsHandler)
			// Top IPs endpoint
			private.GET("/stats/top-ips", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:stats"), ctrl.TopIPsHandler)
			// Top User-Agents endpoint
			private.GET("/stats/top-user-agents", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:stats"), ctrl.TopUserAgentsHandler)
			// Geo distribution endpoint
			private.GET("/stats/geo-distribution", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:stats"), ctrl.GeoDistributionHandler)
			// Overview endpoint
			private.GET("/overview", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:overview"), ctrl.OverviewHandler)
			// Status distribution endpoint
			private.GET("/distribution", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:distribution"), ctrl.StatusDistributionHandler)
			// Log query endpoint
			private.GET("/logs", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:logs"), ctrl.LogsHandler)
			// Real-time logs websocket endpoint
			private.GET("/logs/ws", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:logs"), ctrl.LogsWebSocketHandler)
			// Real-time stats websocket endpoint
			private.GET("/stats/ws", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:stats"), ctrl.StatsWebSocketHandler)
		}
	}

	// Intercept /swagger/swagger-initializer.js requests using middleware to inject our custom operationsSorter configuration.
	// This avoids Gin's panic where wildcard routes (/swagger/*any) conflict with static routes (/swagger/swagger-initializer.js).
	router.Use(func(c *gin.Context) {
		if c.Request.Method == "GET" && c.Request.URL.Path == "/swagger/swagger-initializer.js" {
			c.Header("Content-Type", "application/javascript")
			c.String(200, `window.onload = function() {
  const ui = SwaggerUIBundle({
    url: "doc.json",
    dom_id: '#swagger-ui',
    validatorUrl: null,
    oauth2RedirectUrl: window.location.protocol + "//" + window.location.host + window.location.pathname.split('/').slice(0, window.location.pathname.split('/').length - 1).join('/') + "/oauth2-redirect.html",
    persistAuthorization: false,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout",
    docExpansion: "list",
    deepLinking: true,
    defaultModelsExpandDepth: 1,
    operationsSorter: function(a, b) {
      const pathOrder = {
        "/api/v1/sev/ping": 1,
        "/api/v1/sev/overview": 2,
        "/api/v1/sev/stats/ws": 3,
        "/api/v1/sev/distribution": 4,
        "/api/v1/sev/logs/ws": 5,
        "/api/v1/sev/stats/top-paths": 6,
        "/api/v1/sev/stats/top-ips": 7,
        "/api/v1/sev/stats/top-user-agents": 8,
        "/api/v1/sev/stats/timeseries": 9,
        "/api/v1/sev/logs": 10,
        "/api/v1/sev/stats/geo-distribution": 11,
        "/api/v1/sev/stats": 12
      };
      const pathA = a.get("path");
      const pathB = b.get("path");
      const rankA = pathOrder[pathA] || 999;
      const rankB = pathOrder[pathB] || 999;
      if (rankA !== rankB) {
        return rankA - rankB;
      }
      return pathA.localeCompare(pathB);
    }
  })

  window.ui = ui
}`)
			c.Abort()
			return
		}
		c.Next()
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
