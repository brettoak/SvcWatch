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
			// Top paths endpoint
			private.GET("/stats/top-paths", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:stats"), ctrl.TopPathsHandler)
			// Overview endpoint
			private.GET("/overview", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:overview"), ctrl.OverviewHandler)
			// Status distribution endpoint
			private.GET("/distribution", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:distribution"), ctrl.StatusDistributionHandler)
			// Log query endpoint
			private.GET("/logs", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:logs"), ctrl.LogsHandler)
		}
	}

	// Register Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
