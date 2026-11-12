package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lesmana/sekolah-karir-applicant-api/config"
	_ "github.com/lesmana/sekolah-karir-applicant-api/docs"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/handler"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/repository"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/seeder"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Applicant Records API - Sekolah Karir
// @version         1.0
// @description     RESTful API for applicant records management (create, list, stage-update, detail, health check).
// @contact.name    Lesmana Adh Kusuma (SKR26-0005)
// @contact.email   kandidat@sekolahkarir.com
// @license.name    MIT
// @host            localhost:8080
// @BasePath        /
func main() {
	// Database Path configuration
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "applicants.db"
	}

	// CLI Flags
	seedFlag := flag.Bool("seed", false, "Seed database with mock applicant data")
	flag.Parse()

	// Initialize Database with GORM and SQLite
	db, err := config.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Auto-seed if flag --seed or env SEED_DATA=true is set
	if *seedFlag || os.Getenv("SEED_DATA") == "true" {
		if err := seeder.SeedApplicants(db); err != nil {
			log.Printf("[Seeder] Warning: failed to seed database: %v\n", err)
		}
	}

	// Setup Layered Architecture Dependencies
	applicantRepo := repository.NewApplicantRepository(db)
	applicantService := service.NewApplicantService(applicantRepo)
	applicantHandler := handler.NewApplicantHandler(applicantService)

	// Setup Gin Engine
	router := gin.Default()

	// CORS Middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// System Routes
	router.GET("/health", applicantHandler.HealthCheck)

	// Swagger Interactive Documentation Route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 Routes
	v1 := router.Group("/api/v1")
	{
		applicants := v1.Group("/applicants")
		{
			applicants.POST("", applicantHandler.CreateApplicant)
			applicants.GET("", applicantHandler.ListApplicants)
			applicants.GET("/:id", applicantHandler.GetApplicantByID)
			applicants.PATCH("/:id/stage", applicantHandler.UpdateApplicantStage)
		}
	}

	// Port configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("==================================================")
	log.Printf(" Applicant Records API is running on port %s", port)
	log.Printf(" Swagger UI: http://localhost:%s/swagger/index.html", port)
	log.Printf(" Health Check: http://localhost:%s/health", port)
	log.Printf("==================================================")

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

