package router

import (
	"insider-assessment/internal/handler"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitRoutes registers all API routes and the Swagger UI
func InitRoutes(r *gin.Engine, h *handler.Handler) {
	// swagger Route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/")
	{
		api.POST("/start", h.StartScheduler)
		api.POST("/stop", h.StopScheduler)
		api.GET("/sent-messages", h.GetSentMessages)
		api.POST("/messages", h.AddMessage) // helper for testing
		api.GET("/health", h.HealthCheck)
		api.GET("/messages/cache", h.GetAllCachedMessages)
	}
}
