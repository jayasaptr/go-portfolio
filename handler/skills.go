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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	formatter "github.com/ivanauliaa/response-formatter"
)

func AddSkills(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {

		//check if the user is logged in
		authorizationHeader := c.GetHeader("Authorization")
		if authorizationHeader == "" {
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Authorization header not provided"))
			return
		}

		tokenStirng := strings.TrimPrefix(authorizationHeader, "Bearer ")
		if tokenStirng == "" {
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Token not provided"))
			return
		}

		_, err := ValidateToken(tokenStirng, jwtKey, db)
		if err != nil {
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Invalid token"))
			return
		}

		c.Bind(&model.Skills{})

		skil := model.Skills{
			ID:   uuid.New().String(),
			Name: c.PostForm("name"),
		}

		//handler file upload image
		file, err := c.FormFile("image")
		if err != nil {
			log.Printf("Error retrieving file %v\n", err)
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Failed to get uploaded file"))
			return
		}

		//Generate a new filename for the image
		newFilename := uuid.New().String() + filepath.Ext(file.Filename)
		if err := c.SaveUploadedFile(file, "uploads/skills/"+newFilename); err != nil {
			log.Printf("Error saving file: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to save uploaded file"))
			return
		}
		skil.Image = newFilename

		//insert skil into the database
		if err := model.InsertSkills(db, skil); err != nil {
			log.Printf("Error inserting user into database: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to insert skill into database"))
			return
		}
		c.JSON(http.StatusCreated, formatter.SuccessResponse(skil))
	}
}

func GetSkill(db *sql.DB, jwtKey string) gin.HandlerFunc {
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
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Invalid token"))
			return
		}

		// Parse pagination query parameters
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Invalid limit value"))
			return
		}

		offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
		if err != nil {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Invalid offset value"))
			return
		}

		// Retrieve skills with pagination
		skills, err := model.GetListSkills(db, offset, limit)
		if err != nil {
			log.Printf("Error retrieving skills: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve skills"))
			return
		}

		// Include server URL in the image links
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		for i := range skills {
			skills[i].Image = scheme + "://" + c.Request.Host + "/uploads/skills/" + skills[i].Image
		}

		c.JSON(http.StatusOK, formatter.SuccessResponse(skills))
	}
}

func DeleteSkill(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		//check user login
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

		skilIDToDelete := c.Param("id")
		if skilIDToDelete == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Skill id required"))
			return
		}

		//retrive skill to get the image path
		skill, err := model.GetSkillID(db, skilIDToDelete)

		if err != nil {
			log.Printf("Error retriving skillll %v\n", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve skill"))
		}

		//delete skill from db
		if err := model.DeleteSkill(db, skilIDToDelete); err != nil {
			log.Printf("Error deleting skill from database: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to delete skill"))
			return
		}

		// Delete the skill's image file
		if skill.Image != "" {
			imagePath := "./uploads/skills/" + skill.Image
			if err := os.Remove(imagePath); err != nil {
				log.Printf("Error deleting image file: %v", err)
				c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to delete skill image"))
				return
			}
		}

		// Return success response
		c.JSON(http.StatusOK, formatter.SuccessResponse("Skill deleted successfully"))
	}
}
