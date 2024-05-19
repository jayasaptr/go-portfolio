package handler

import (
	"database/sql"
	"log"
	"net/http"
	"path/filepath"
	"portfolio/model"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	formatter "github.com/ivanauliaa/response-formatter"
)

func AddPortfolioWithSkills(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate JWT token
		authorizationHeader := c.GetHeader("Authorization")
		if authorizationHeader == "" {
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Authorization header not provided"))
			return
		}

		tokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Token not provided"))
			return
		}

		_, err := ValidateToken(tokenString, jwtKey, db)
		if err != nil {
			log.Printf("Error validating token: %v", err)
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Invalid token"))
			return
		}

		// Parse form data
		title := c.PostForm("title")
		content := c.PostForm("content")
		file, err := c.FormFile("image")
		if err != nil {
			log.Printf("Error retrieving form file: %v", err)
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Image file is required"))
			return
		}

		// Generate unique ID for the portfolio
		portfolioID := uuid.New().String()

		// Save the image in the uploads/portfolio directory
		newFilename := uuid.New().String() + filepath.Ext(file.Filename)
		if err := c.SaveUploadedFile(file, "uploads/portfolio/"+newFilename); err != nil {
			log.Printf("Error saving uploaded file: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to save image"))
			return
		}

		// Create portfolio instance
		portfolio := model.Portfolio{
			ID:      portfolioID,
			Title:   title,
			Image:   newFilename,
			Content: content,
		}

		// Insert portfolio into database
		if err := model.InsertPortfolio(db, &portfolio); err != nil {
			log.Printf("Error inserting portfolio into database: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to insert portfolio"))
			return
		}

		// Retrieve skill IDs from form data
		skillIDs := c.PostFormArray("skill_ids")
		if len(skillIDs) == 0 {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("At least one skill ID is required"))
			return
		}

		// Add skills to the portfolio
		if err := portfolio.AddSkills(db, skillIDs); err != nil {
			log.Printf("Error adding skills to portfolio: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to add skills to portfolio"))
			return
		}

		// Return success response
		c.JSON(http.StatusOK, formatter.SuccessResponse(map[string]interface{}{
			"id":      portfolio.ID,
			"title":   portfolio.Title,
			"content": portfolio.Content,
			"image":   newFilename,
			"skills":  skillIDs,
		}))
	}
}

func GetPortfolioAndSkillsPaginated(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the user is logged in
		authorizationHeader := c.GetHeader("Authorization")
		if authorizationHeader == "" {
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Authorization header not provided"))
			return
		}

		tokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Token not provided"))
			return
		}

		_, err := ValidateToken(tokenString, jwtKey, db)
		if err != nil {
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Invalid token"))
			return
		}

		// Pagination parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		offset := (page - 1) * limit

		// Retrieve portfolios with pagination
		portfolios, err := model.GetPortfoliosPaginated(db, offset, limit)
		if err != nil {
			log.Printf("Error retrieving portfolios: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve portfolios"))
			return
		}

		// Retrieve skills for each portfolio
		for i, portfolio := range portfolios {
			skills, err := model.GetSkillsByPortfolioID(db, portfolio.ID)
			if err != nil {
				log.Printf("Error retrieving skills for portfolio %s: %v", portfolio.ID, err)
				c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve skills for portfolio"))
				return
			}
			portfolios[i].Skills = skills
		}

		// Return success response with portfolios and their skills
		c.JSON(http.StatusOK, formatter.SuccessResponse(map[string]interface{}{
			"portfolios": portfolios,
		}))
	}
}
