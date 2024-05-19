package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"portfolio/handler"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	db, err := sql.Open("pgx", os.Getenv("DB_URI"))
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

	r.POST("/api/v1/auth/register", handler.RegisterAuth(db))
	r.POST("/api/v1/auth/login", handler.LoginAuth(db, os.Getenv("JWT_SECRET")))
	r.GET("/api/v1/user", handler.GetUserWithJWT(db, os.Getenv("JWT_SECRET")))

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
