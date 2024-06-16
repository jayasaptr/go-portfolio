package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"portfolio/handler"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	// Muat file .env
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading env file %v\n", err)
		os.Exit(1)
	}

	connStr, err := loadPostgresConfig()
	if err != nil {
		fmt.Printf("Gagal membuat koneksi database %v\n", err)
		os.Exit(1)
	}
	// db, err := sql.Open("pgx", os.Getenv("DB_URI"))
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		fmt.Printf("Gagal membuat koneksi database %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		fmt.Printf("Gagal memverifikasi koneksi database : %v\n", err)
		os.Exit(1)
	}

	if _, err = migrate(db); err != nil {
		fmt.Printf("Gagal melakukan migrasi databse : %v\n", err)
		os.Exit(1)
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	r.POST("/api/v1/auth/register", handler.RegisterAuth(db))
	r.POST("/api/v1/auth/login", handler.LoginAuth(db, os.Getenv("JWT_SECRET")))
	r.GET("/api/v1/user", handler.GetUserWithJWT(db, os.Getenv("JWT_SECRET")))
	r.DELETE("/api/v1/user", handler.DeleteUser(db, os.Getenv("JWT_SECRET")))

	//skills
	r.POST("/api/v1/skills", handler.AddSkills(db, os.Getenv("JWT_SECRET")))
	r.GET("/api/v1/skills", handler.GetSkill(db))
	r.GET("/api/v1/skills/:id", handler.GetSkillByID(db))
	r.PUT("/api/v1/skills/:id", handler.UpdateSkill(db, os.Getenv("JWT_SECRET")))
	r.DELETE("/api/v1/skills/:id", handler.DeleteSkill(db, os.Getenv("JWT_SECRET")))

	//portfolio
	r.POST("/api/v1/portfolio", handler.AddPortfolioWithSkills(db, os.Getenv("JWT_SECRET")))
	r.GET("/api/v1/portfolio", handler.GetPortfolioAndSkillsPaginated(db))
	r.GET("/api/v1/portfolio/:id", handler.GetPortfolioAndSkillsByID(db))
	r.DELETE("/api/v1/portfolio/:id", handler.DeletePortfolioHandler(db, os.Getenv("JWT_SECRET")))
	r.PUT("/api/v1/portfolio/:id", handler.UpdatePortfolioHandler(db, os.Getenv("JWT_SECRET")))

	//experience
	r.POST("/api/v1/experience", handler.AddExperiance(db, os.Getenv("JWT_SECRET")))
	r.GET("/api/v1/experience", handler.GetExperience(db))
	r.GET("/api/v1/experience/:id", handler.GetExperienceByID(db))
	r.PUT("/api/v1/experience/:id", handler.UpdateExperience(db, os.Getenv("JWT_SECRET")))
	r.DELETE("/api/v1/experience/:id", handler.DeleteExperience(db, os.Getenv("JWT_SECRET")))

	r.PUT("/api/v1/portfolio-skill/:id", handler.AddSkillsToPortfolio(db, os.Getenv("JWT_SECRET")))
	r.PUT("/api/v1/experience-skill/:id", handler.AddSkillsToExperience(db, os.Getenv("JWT_SECRET")))

	r.PUT("/api/v1/portfolio-skill/:id", handler.DeleteSkillWithRelationsHandler(db, os.Getenv("JWT_SECRET")))
	r.PUT("/api/v1/experience-skill/:id", handler.DeleteSkillExperienceWithRelationsHandler(db, os.Getenv("JWT_SECRET")))

	// Serve static files for images
	r.Static("/uploads", "./uploads")

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	if err = server.ListenAndServe(); err != nil {
		fmt.Printf("Gagal menjalankan server %v\n", err)
		os.Exit(1)
	}
}

func loadPostgresConfig() (string, error) {
	if os.Getenv("DB_HOST") == "" {
		return "", fmt.Errorf("environment variable DB_HOST must be set")
	}
	if os.Getenv("DB_PORT") == "" {
		return "", fmt.Errorf("environment variable DB_PORT must be set")
	}
	if os.Getenv("DB_USER") == "" {
		return "", fmt.Errorf("environment variable DB_USER must be set")
	}
	if os.Getenv("DB_DATABASE") == "" {
		return "", fmt.Errorf("environment variable DB_DATABASE must be set")
	}
	if os.Getenv("DB_PASSWORD") == "" {
		return "", fmt.Errorf("environment variable DB_PASSWORD must be set")
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_DATABASE"),
	)
	return connStr, nil
}
