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

func AddExperiance(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validation with JWT
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
			log.Printf("Error validating token: %v\n", err)
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Invalid Token"))
			return
		}

		// Parse form data
		companyName := c.PostForm("company_name")
		position := c.PostForm("position")
		startDateStr := c.PostForm("start_date")
		endDateStr := c.PostForm("end_date")
		location := c.PostForm("location")
		file, err := c.FormFile("image")

		if err != nil {
			log.Printf("Error retrieving form file: %v\n", err)
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Image file is required"))
			return
		}

		// Generate UUID for the experience
		experienceID := uuid.New().String()

		// Save image to file
		newFileName := uuid.New().String() + filepath.Ext(file.Filename)
		imagePath := "uploads/experience/" + newFileName
		if err := c.SaveUploadedFile(file, imagePath); err != nil {
			log.Printf("Error saving uploaded file: %v\n", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to save image"))
			return
		}

		// Parse dates
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			log.Printf("Error parsing start date: %v\n", err)
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Invalid start date format, expected yyyy-mm-dd"))
			return
		}
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			log.Printf("Error parsing end date: %v\n", err)
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Invalid end date format, expected yyyy-mm-dd"))
			return
		}

		// Create experience
		experience := model.Experience{
			ID:          experienceID,
			CompanyName: companyName,
			Image:       newFileName,
			Position:    position,
			StartDate:   startDate,
			EndDate:     endDate,
			Location:    location,
		}

		// Insert experience into the database
		if err := model.InsertExperience(db, &experience); err != nil {
			log.Printf("Error inserting experience into database: %v\n", err)
			// Delete the saved image if database transaction fails
			if removeErr := os.Remove(imagePath); removeErr != nil {
				log.Printf("Error removing image file after failed database insert: %v\n", removeErr)
			}
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to insert experience"))
			return
		}

		// Retrieve skill IDs from form data
		skillIDs := c.PostFormArray("skill_ids")
		if len(skillIDs) == 0 {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("At least one skill ID is required"))
			return
		}

		// Add skills to the portfolio
		if err := experience.AddSkills(db, skillIDs); err != nil {
			log.Printf("Error adding skills to portfolio: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to add skills to portfolio"))
			return
		}

		c.JSON(http.StatusCreated, formatter.SuccessResponse(map[string]interface{}{
			"id":           experience.ID,
			"company_name": experience.CompanyName,
			"image":        experience.Image,
			"position":     experience.Position,
			"location":     experience.Location,
			"start_date":   experience.StartDate,
			"end_date":     experience.EndDate,
			"skills":       skillIDs,
		}))
	}
}

func AddSkillsToExperience(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkUserLogin(c, jwtKey, db) {
			return
		}

		experienceID := c.Param("id")
		if experienceID == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Experience ID is required"))
			return
		}

		// Retrieve skill IDs from form input
		skillIDs := c.PostFormArray("skill_ids")
		if len(skillIDs) == 0 {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Skill IDs are required"))
			return
		}

		experience, err := model.GetExperienceID(db, experienceID)
		if err != nil {
			log.Printf("Error retrieving portfolio: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve portfolio"))
			return
		}

		if err := experience.AddSkills(db, skillIDs); err != nil {
			log.Printf("Error adding skills to portfolio: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to add skills to portfolio"))
			return
		}

		c.JSON(http.StatusOK, formatter.SuccessResponse("Skills successfully added to experience"))
	}
}

func GetExperience(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		//pagination parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		offset := (page - 1) * limit

		//retrieve experience with pagination
		experiences, err := model.GetExperience(db, offset, limit)
		if err != nil {
			log.Printf("Error retrieving experience: %v\n", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve experiences"))
			return
		}

		//include server url in the image link
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}

		//retrieve experiences for each experience and inclue image paths
		for i := range experiences {
			experiences[i].Image = scheme + "://" + c.Request.Host + "/uploads/experience/" + experiences[i].Image
		}

		for i, experience := range experiences {
			skills, err := model.GetSkillByExperienceID(db, experience.ID)
			if err != nil {
				log.Printf("Error retriving skill for experience %s: %v", experience.ID, err)
				c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrive skill for experience"))
				return
			}

			for j := range skills {
				skills[j].Image = scheme + "://" + c.Request.Host + "/uploads/skills/" + skills[j].Image
			}

			experiences[i].Skills = skills
		}

		//return success response
		c.JSON(http.StatusOK, formatter.SuccessResponse(map[string]interface{}{
			"experience": experiences,
		}))
	}
}

func GetExperienceByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		experienceID := c.Param("id")
		if experienceID == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Experience ID is required"))
			return
		}

		experience, err := model.GetExperienceID(db, experienceID)
		if err != nil {
			log.Printf("Error retrieving experience by ID: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve experience"))
			return
		}

		// Include server URL in the image link
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		experience.Image = scheme + "://" + c.Request.Host + "/uploads/experience/" + experience.Image

		//get skill by experience
		// for i, experienceSkill := range experience {
		// 	skills, err := model.GetSkillByExperienceID(db, experience.ID)
		// 	if err != nil {
		// 		log.Printf("Error retriving skill for experience %s: %v", experience.ID, err)
		// 		c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrive skill for experience"))
		// 		return
		// 	}

		// 	for j := range skills {
		// 		skills[j].Image = scheme + "://" + c.Request.Host + "/uploads/skills/" + skills[j].Image
		// 	}

		// 	experience.Skills = skills
		// }

		skills, err := model.GetSkillByExperienceID(db, experience.ID)
		if err != nil {
			log.Printf("Error retriving skill for experience %s: %v", experience.ID, err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrive skill for experience"))
			return
		}

		for j := range skills {
			skills[j].Image = scheme + "://" + c.Request.Host + "/uploads/skills/" + skills[j].Image
		}

		experience.Skills = skills

		c.JSON(http.StatusOK, formatter.SuccessResponse(experience))
	}
}

func UpdateExperience(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkUserLogin(c, jwtKey, db) {
			return
		}

		experienceID := c.Param("id")
		if experienceID == "" {
			c.JSON(http.StatusBadGateway, formatter.BadRequestResponse("Experience id is required"))
			return
		}

		if err := c.Request.ParseForm(); err != nil {
			log.Printf("Error parsing form data: %v\n", err)
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Error parsing form data"))
			return
		}

		existingExperience, err := model.GetExperienceID(db, experienceID)
		if err != nil {
			log.Printf("Error retrieving exsisting experience %v\n", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve experience"))
			return
		}

		//update
		companyName := c.PostForm("company_name")
		if companyName != "" && companyName != existingExperience.CompanyName {
			existingExperience.CompanyName = companyName
		}

		//handle update image
		file, header, err := c.Request.FormFile("image")
		if err == nil {
			// Generate new filename and save the file
			newFilename := uuid.New().String() + filepath.Ext(header.Filename)
			if err := c.SaveUploadedFile(header, "./uploads/experience/"+newFilename); err != nil {
				log.Printf("Error saving uploaded file: %v", err)
				c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to save image"))
				return
			}

			// Delete old image if new one is uploaded
			if existingExperience.Image != "" {
				oldImagePath := "./uploads/experience/" + existingExperience.Image
				if err := os.Remove(oldImagePath); err != nil {
					if !os.IsNotExist(err) {
						log.Printf("Error deleting old image file: %v", err)
						c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to delete old image"))
						return
					}
				}
			}

			// Update experience image to new filename
			existingExperience.Image = newFilename
		} else if err != http.ErrMissingFile {
			log.Printf("Error retrieving image file: %v", err)
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Failed to retrieve image file"))
			return
		}

		position := c.PostForm("position")
		if position != "" && position != existingExperience.Position {
			existingExperience.Position = position
		}

		startDateStr := c.PostForm("start_date")
		if startDateStr != "" {
			startDate, err := time.Parse("2006-01-02", startDateStr)
			if err != nil {
				log.Printf("Error parsing start date: %v", err)
				c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Invalid start date format"))
				return
			}
			if startDate != existingExperience.StartDate {
				existingExperience.StartDate = startDate
			}
		}

		endDateStr := c.PostForm("end_date")
		if endDateStr != "" {
			endDate, err := time.Parse("2006-01-02", endDateStr)
			if err != nil {
				log.Printf("Error parsing start date: %v", err)
				c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Invalid end date format"))
				return
			}
			if endDate != existingExperience.EndDate {
				existingExperience.EndDate = endDate
			}
		}

		location := c.PostForm("location")
		if location != "" && location != existingExperience.Location {
			existingExperience.Location = location
		}

		//update data
		if companyName != "" || file != nil || position != "" || startDateStr != "" || endDateStr != "" || location != "" {
			err = model.UpdateExperience(db, existingExperience)
			if err != nil {
				log.Printf("Error updating experience: %v", err)
				c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to update experiemce"))
				return
			}
		}

		c.JSON(http.StatusOK, formatter.SuccessResponse("Experience updated successfully"))
	}
}

func DeleteExperience(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkUserLogin(c, jwtKey, db) {
			return
		}

		experienceID := c.Param("id")

		if experienceID == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Experience id is required"))
			return
		}

		experience, err := model.GetExperienceID(db, experienceID)
		if err != nil {
			log.Printf("Err retriveing experience: %v\n", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve experience"))
			return
		}

		err = model.DeleteExperienceAndRelations(db, experienceID)
		if err != nil {
			log.Printf("Error deleting skill with relations: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to delete experience and its relations"))
			return
		}

		if experience.ID != "" {
			imagePath := "./uploads/experience/" + experience.Image
			if err := os.Remove(imagePath); err != nil {
				if !os.IsNotExist(err) {
					log.Printf("Error deleting image file: %v", err)
					c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to delete portfolio image"))
					return
				}
			}
		}

		c.JSON(http.StatusOK, formatter.SuccessResponse("Experience deleted successfully"))
	}
}

func DeleteSkillExperienceWithRelationsHandler(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkUserLogin(c, jwtKey, db) {
			return
		}

		experienceID := c.Param("id")
		if experienceID == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Experience ID is required"))
			return
		}

		skillID := c.PostForm("skill_id")
		if skillID == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Skill ID is required"))
			return
		}

		err := model.DeleteSkillAndExperienceRelations(db, skillID, experienceID)
		if err != nil {
			log.Printf("Error deleting skill with relations: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to delete skill from portfolio"))
			return
		}

		c.JSON(http.StatusOK, formatter.SuccessResponse("Skill successfully deleted from portfolio"))
	}
}
