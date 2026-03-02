package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/EduGoGroup/edugo-shared/common/types/enum"
	sharedmw "github.com/EduGoGroup/edugo-shared/middleware/gin"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/container"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
)

// Setup creates and configures the Gin router with all routes.
func Setup(c *container.Container) *gin.Engine {
	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(middleware.Metrics())
	r.Use(middleware.ErrorHandler(c.Logger))
	r.Use(middleware.RequestLogging(c.Logger))

	// Public endpoints
	r.GET("/health", c.Handlers.Health.Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1
	v1 := r.Group("/api/v1")
	v1.GET("/health", c.Handlers.Health.Health) // also under basePath for Swagger
	v1.Use(middleware.RemoteAuthMiddleware(c.AuthClient))

	// Materials
	materials := v1.Group("/materials")
	{
		materials.GET("", sharedmw.RequirePermission(enum.PermissionMaterialsRead), c.Handlers.Material.List)
		materials.GET("/:id", sharedmw.RequirePermission(enum.PermissionMaterialsRead), c.Handlers.Material.GetByID)
		materials.GET("/:id/versions", sharedmw.RequirePermission(enum.PermissionMaterialsRead), c.Handlers.Material.GetWithVersions)
		materials.GET("/:id/download-url", sharedmw.RequirePermission(enum.PermissionMaterialsDownload), c.Handlers.Material.GenerateDownloadURL)
		materials.GET("/:id/summary", sharedmw.RequirePermission(enum.PermissionMaterialsRead), c.Handlers.Summary.GetSummary)
		materials.GET("/:id/assessment", sharedmw.RequirePermission(enum.PermissionAssessmentsRead), c.Handlers.Assessment.GetAssessment)
		materials.GET("/:id/stats", sharedmw.RequirePermission(enum.PermissionStatsUnit), c.Handlers.Stats.GetMaterialStats)

		materials.POST("", sharedmw.RequirePermission(enum.PermissionMaterialsCreate), c.Handlers.Material.Create)
		materials.POST("/:id/upload-url", sharedmw.RequirePermission(enum.PermissionMaterialsCreate), c.Handlers.Material.GenerateUploadURL)
		materials.POST("/:id/upload-complete", sharedmw.RequirePermission(enum.PermissionMaterialsCreate), c.Handlers.Material.UploadComplete)

		materials.PUT("/:id", sharedmw.RequirePermission(enum.PermissionMaterialsUpdate), c.Handlers.Material.Update)

		// Assessment attempts nested under material
		materials.POST("/:id/assessment/attempts", sharedmw.RequirePermission(enum.PermissionAssessmentsAttempt), c.Handlers.Assessment.CreateAttempt)
	}

	// Assessment Management (teacher/admin)
	assessments := v1.Group("/assessments")
	{
		assessments.GET("", sharedmw.RequirePermission(enum.PermissionAssessmentsRead), c.Handlers.AssessmentMgmt.List)
		assessments.GET("/:id", sharedmw.RequirePermission(enum.PermissionAssessmentsRead), c.Handlers.AssessmentMgmt.GetByID)
		assessments.POST("", sharedmw.RequirePermission(enum.PermissionAssessmentsCreate), c.Handlers.AssessmentMgmt.Create)
		assessments.PUT("/:id", sharedmw.RequirePermission(enum.PermissionAssessmentsUpdate), c.Handlers.AssessmentMgmt.Update)
		assessments.PATCH("/:id/publish", sharedmw.RequirePermission(enum.PermissionAssessmentsPublish), c.Handlers.AssessmentMgmt.Publish)
		assessments.PATCH("/:id/archive", sharedmw.RequirePermission(enum.PermissionAssessmentsUpdate), c.Handlers.AssessmentMgmt.Archive)
		assessments.DELETE("/:id", sharedmw.RequirePermission(enum.PermissionAssessmentsDelete), c.Handlers.AssessmentMgmt.Delete)

		// Question management (nested under assessment)
		assessments.GET("/:id/questions", sharedmw.RequirePermission(enum.PermissionAssessmentsRead), c.Handlers.AssessmentMgmt.GetQuestions)
		assessments.POST("/:id/questions", sharedmw.RequirePermission(enum.PermissionAssessmentsCreate), c.Handlers.AssessmentMgmt.AddQuestion)
		assessments.PUT("/:id/questions/:idx", sharedmw.RequirePermission(enum.PermissionAssessmentsUpdate), c.Handlers.AssessmentMgmt.UpdateQuestion)
		assessments.DELETE("/:id/questions/:idx", sharedmw.RequirePermission(enum.PermissionAssessmentsDelete), c.Handlers.AssessmentMgmt.DeleteQuestion)
	}

	// Attempts (top-level)
	attempts := v1.Group("/attempts")
	{
		attempts.GET("/:id/results", sharedmw.RequirePermission(enum.PermissionAssessmentsViewResults), c.Handlers.Assessment.GetAttemptResult)
	}

	// Users
	users := v1.Group("/users")
	{
		users.GET("/me/attempts", sharedmw.RequirePermission(enum.PermissionAssessmentsViewResults), c.Handlers.Assessment.ListUserAttempts)
	}

	// Progress
	v1.PUT("/progress", sharedmw.RequirePermission(enum.PermissionProgressUpdate), c.Handlers.Progress.Upsert)

	// Screens
	screens := v1.Group("/screens")
	{
		screens.GET("/navigation", sharedmw.RequirePermission(enum.PermissionScreensRead), c.Handlers.Screen.GetNavigation)
		screens.GET("/resource/:resourceKey", sharedmw.RequirePermission(enum.PermissionScreensRead), c.Handlers.Screen.GetScreensByResource)
		screens.GET("/:screenKey", sharedmw.RequirePermission(enum.PermissionScreensRead), c.Handlers.Screen.GetScreen)
		screens.PUT("/:screenKey/preferences", sharedmw.RequirePermission(enum.PermissionScreenInstancesUpdate), c.Handlers.Screen.SavePreferences)
	}

	// Stats
	stats := v1.Group("/stats")
	{
		stats.GET("/global", sharedmw.RequirePermission(enum.PermissionStatsGlobal), c.Handlers.Stats.GetGlobalStats)
	}

	// Guardian Relations (self-registration)
	guardianRelations := v1.Group("/guardian-relations")
	{
		guardianRelations.POST("/request", sharedmw.RequirePermission(enum.PermissionGuardianRelationsRequest), c.Handlers.Guardian.RequestRelation)
		guardianRelations.GET("/pending", sharedmw.RequirePermission(enum.PermissionGuardianRelationsRead), c.Handlers.Guardian.ListPendingRequests)
		guardianRelations.POST("/:id/approve", sharedmw.RequirePermission(enum.PermissionGuardianRelationsApprove), c.Handlers.Guardian.ApproveRequest)
		guardianRelations.POST("/:id/reject", sharedmw.RequirePermission(enum.PermissionGuardianRelationsApprove), c.Handlers.Guardian.RejectRequest)
	}

	// Guardians (my children + stats)
	guardians := v1.Group("/guardians")
	{
		guardians.GET("/me/children", sharedmw.RequirePermission(enum.PermissionGuardianRelationsRequest), c.Handlers.Guardian.ListMyChildren)
		guardians.GET("/me/children/:childId/progress", sharedmw.RequirePermission(enum.PermissionGuardianRelationsRequest), c.Handlers.Guardian.GetChildProgress)
		guardians.GET("/me/stats", sharedmw.RequirePermission(enum.PermissionGuardianRelationsRequest), c.Handlers.Guardian.GetMyStats)
	}

	return r
}
