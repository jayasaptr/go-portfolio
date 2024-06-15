package handler

import (
	"database/sql"
	"io"
	"log"
	"mime/multipart"
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

func GetSkill(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

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

func GetSkillByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		skillID := c.Param("id")
		if skillID == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Skill ID is required"))
			return
		}

		skill, err := model.GetSkillID(db, skillID)
		if err != nil {
			log.Printf("Error retrieving skill by ID: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve skill"))
			return
		}

		// Include server URL in the image link
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		skill.Image = scheme + "://" + c.Request.Host + "/uploads/skills/" + skill.Image

		c.JSON(http.StatusOK, formatter.SuccessResponse(skill))
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

func UpdateSkill(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check user login
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

		skillIDToUpdate := c.Param("id")
		if skillIDToUpdate == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Skill id required"))
			return
		}

		// Retrieve existing skill to check for image update
		existingSkill, err := model.GetSkillID(db, skillIDToUpdate)
		if err != nil {
			log.Printf("Error retrieving skill: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve skill"))
			return
		}

		// Parse form data
		if err := c.Request.ParseForm(); err != nil {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Error parsing form data"))
			return
		}

		// Update skill name from form data
		newSkillName := c.PostForm("name")
		if newSkillName != "" {
			existingSkill.Name = newSkillName
		}

		// Assume new image is uploaded with form key 'image'
		file, header, err := c.Request.FormFile("image")
		if err == nil {
			defer file.Close()

			// Generate new image filename and save
			newImageName := uuid.New().String() + filepath.Ext(header.Filename)
			newImagePath := "uploads/skills/" + newImageName
			if err := saveUploadedFile(header, newImagePath); err != nil {
				log.Printf("Error saving new image: %v", err)
				c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to save new image"))
				return
			}

			// Delete old image if new one is uploaded
			if existingSkill.Image != "" {
				oldImagePath := "./uploads/skills/" + existingSkill.Image
				if err := os.Remove(oldImagePath); err != nil {
					log.Printf("Error deleting old image file: %v", err)
					c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to delete old image"))
					// Continue to update the skill even if old image deletion fails
				}
			}

			// Update skill record with new image name
			existingSkill.Image = newImageName
		}

		// Update skill in database
		if err := model.UpdateSkill(db, existingSkill); err != nil {
			log.Printf("Error updating skill: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to update skill"))
			return
		}

		c.JSON(http.StatusOK, formatter.SuccessResponse("Skill updated successfully"))
	}
}

func saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
