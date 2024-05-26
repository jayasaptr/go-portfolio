package handler

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"portfolio/model"
	"strconv"
	"strings"
	"time"

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

		status := c.PostForm("status")
		dateProjectStr := c.PostForm("date_project")

		//parse dateProject
		dateProject, err := time.Parse("2006-01-02", dateProjectStr)

		if err != nil {
			log.Printf("Error parsing start date: %v\n", err)
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Invalid start date format, expected yyyy-mm-dd"))
			return
		}

		// Create portfolio instance
		portfolio := model.Portfolio{
			ID:          portfolioID,
			Title:       title,
			Image:       newFilename,
			Content:     content,
			Status:      status,
			DateProject: dateProject,
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

func GetPortfolioAndSkillsPaginated(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		// Include server URL in the image link
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}

		// Retrieve skills for each portfolio and include image paths
		for i, portfolio := range portfolios {
			skills, err := model.GetSkillsByPortfolioID(db, portfolio.ID)
			if err != nil {
				log.Printf("Error retrieving skills for portfolio %s: %v", portfolio.ID, err)
				c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve skills for portfolio"))
				return
			}

			// Append the server path to each skill's image
			for j := range skills {
				skills[j].Image = scheme + "://" + c.Request.Host + "/uploads/skills/" + skills[j].Image
			}

			portfolios[i].Skills = skills
		}

		// Append the server path to each portfolio's image
		for i := range portfolios {
			portfolios[i].Image = scheme + "://" + c.Request.Host + "/uploads/portfolio/" + portfolios[i].Image
		}

		// Return success response with portfolios and their skills
		c.JSON(http.StatusOK, formatter.SuccessResponse(map[string]interface{}{
			"portfolios": portfolios,
		}))
	}
}

func DeleteSkillWithRelationsHandler(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkUserLogin(c, jwtKey, db) {
			return
		}

		portfolioID := c.Param("id")
		if portfolioID == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Portfolio ID is required"))
			return
		}

		skillID := c.PostForm("skill_id")
		if skillID == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Skill ID is required"))
			return
		}

		err := model.DeleteSkillAndPortfolioRelations(db, skillID, portfolioID)
		if err != nil {
			log.Printf("Error deleting skill with relations: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to delete skill from portfolio"))
			return
		}

		c.JSON(http.StatusOK, formatter.SuccessResponse("Skill successfully deleted from portfolio"))
	}
}

func AddSkillsToPortfolio(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkUserLogin(c, jwtKey, db) {
			return
		}

		portfolioID := c.Param("id")
		if portfolioID == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Portfolio ID is required"))
			return
		}

		// Retrieve skill IDs from form input
		skillIDs := c.PostFormArray("skill_ids")
		if len(skillIDs) == 0 {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Skill IDs are required"))
			return
		}

		portfolio, err := model.GetPortfolioByID(db, portfolioID)
		if err != nil {
			log.Printf("Error retrieving portfolio: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve portfolio"))
			return
		}

		if err := portfolio.AddSkills(db, skillIDs); err != nil {
			log.Printf("Error adding skills to portfolio: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to add skills to portfolio"))
			return
		}

		c.JSON(http.StatusOK, formatter.SuccessResponse("Skills successfully added to portfolio"))
	}
}

func UpdatePortfolioHandler(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkUserLogin(c, jwtKey, db) {
			return
		}

		portfolioID := c.Param("id")
		if portfolioID == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Portfolio ID is required"))
			return
		}

		if err := c.Request.ParseForm(); err != nil {
			log.Printf("Error parsing form data: %v", err)
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Error parsing form data"))
			return
		}

		// Retrieve existing portfolio to update
		existingPortfolio, err := model.GetPortfolioByID(db, portfolioID)
		if err != nil {
			log.Printf("Error retrieving existing portfolio: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve portfolio"))
			return
		}

		// Update portfolio instance with form data if provided
		title := c.PostForm("title")
		if title != "" && title != existingPortfolio.Title {
			existingPortfolio.Title = title
		}

		subtitle := c.PostForm("subtitle")
		if subtitle != "" && subtitle != existingPortfolio.Subtitle {
			existingPortfolio.Subtitle = subtitle
		}

		content := c.PostForm("content")
		if content != "" && content != existingPortfolio.Content {
			existingPortfolio.Content = content
		}

		// Handle file upload for image if provided
		file, header, err := c.Request.FormFile("image")
		if err == nil {
			// Generate new filename and save the file
			newFilename := uuid.New().String() + filepath.Ext(header.Filename)
			if err := c.SaveUploadedFile(header, "./uploads/portfolio/"+newFilename); err != nil {
				log.Printf("Error saving uploaded file: %v", err)
				c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to save image"))
				return
			}

			// Delete old image if new one is uploaded
			if existingPortfolio.Image != "" {
				oldImagePath := "./uploads/portfolio/" + existingPortfolio.Image
				if err := os.Remove(oldImagePath); err != nil {
					if !os.IsNotExist(err) {
						log.Printf("Error deleting old image file: %v", err)
						c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to delete old image"))
						return
					}
				}
			}

			// Update portfolio image to new filename
			existingPortfolio.Image = newFilename
		} else if err != http.ErrMissingFile {
			log.Printf("Error retrieving image file: %v", err)
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Failed to retrieve image file"))
			return
		}

		status := c.PostForm("status")
		if status != "" && status != existingPortfolio.Status {
			existingPortfolio.Status = status
		}

		dateProjectStr := c.PostForm("date_project")
		if dateProjectStr != "" {
			dateProject, err := time.Parse("2006-01-02", dateProjectStr)
			if err != nil {
				log.Printf("Error parsing start date: %v", err)
				c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Invalid end date format"))
				return
			}
			if dateProject != existingPortfolio.DateProject {
				existingPortfolio.DateProject = dateProject
			}
		}

		// Update the portfolio in the database only if changes were made
		if title != "" || subtitle != "" || content != "" || file != nil || status != "" || dateProjectStr != "" {
			err = model.UpdatePortfolio(db, existingPortfolio)
			if err != nil {
				log.Printf("Error updating portfolio: %v", err)
				c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to update portfolio"))
				return
			}
		}

		c.JSON(http.StatusOK, formatter.SuccessResponse("Portfolio updated successfully"))
	}
}

func DeletePortfolioHandler(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkUserLogin(c, jwtKey, db) {
			return
		}

		portfolioID := c.Param("id")
		if portfolioID == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Portfolio ID is required"))
			return
		}

		portfolio, err := model.GetPortfolioByID(db, portfolioID)
		if err != nil {
			log.Printf("Error retrieving portfolio: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve portfolio"))
			return
		}

		err = model.DeletePortfolioAndRelations(db, portfolioID)
		if err != nil {
			log.Printf("Error deleting portfolio and its relations: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to delete portfolio and its relations"))
			return
		}

		if portfolio.Image != "" {
			imagePath := "./uploads/portfolio/" + portfolio.Image
			if err := os.Remove(imagePath); err != nil {
				if !os.IsNotExist(err) {
					log.Printf("Error deleting image file: %v", err)
					c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to delete portfolio image"))
					return
				}
			}
		}

		c.JSON(http.StatusOK, formatter.SuccessResponse("Portfolio and its relations deleted successfully"))
	}
}

func checkUserLogin(c *gin.Context, jwtKey string, db *sql.DB) bool {
	authorizationHeader := c.GetHeader("Authorization")
	if authorizationHeader == "" {
		c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Authorization header not provided"))
		return false
	}

	tokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Token not provided"))
		return false
	}

	_, err := ValidateToken(tokenString, jwtKey, db)
	if err != nil {
		c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Invalid token"))
		return false
	}

	return true
}
