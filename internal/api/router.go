package api

import (
	"SvcWatch/internal/config"
	"SvcWatch/internal/middleware"
	"SvcWatch/internal/monitor"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// APIController holds dependencies for API handlers.
type APIController struct {
	monitors []*monitor.Monitor
	cfg      *config.Config
}

// PingHandler Health Check
// @Summary Health Check
// @Description returns a "pong" string to verify the service is running
// @Tags System
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sev/ping [get]
func (ctrl *APIController) PingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// StatsHandler Get aggregated logs statistics
// @Summary Get aggregated logs statistics
// @Description Retrieves the total log count for each configured monitored target.
// @Tags Statistics
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sev/stats [get]
func (ctrl *APIController) StatsHandler(c *gin.Context) {
	stats := make(map[string]interface{})
	for i, monInst := range ctrl.monitors {
		tableName := ctrl.cfg.Targets[i].Table
		stats[tableName] = monInst.GetStats()["total_logs"]
	}
	c.JSON(200, stats)
}

// SetupRouter initializes and configures the Gin API router.
func SetupRouter(monitors []*monitor.Monitor, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Enable CORS for all origins (fixes Swagger UI doc.json fetch issues)
	router.Use(cors.Default())

	ctrl := &APIController{monitors: monitors, cfg: cfg}

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
		}
	}

	// Register Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
