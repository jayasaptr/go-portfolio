package handler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"portfolio/model"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	formatter "github.com/ivanauliaa/response-formatter"
	"golang.org/x/crypto/bcrypt"
)

func RegisterAuth(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Bind(&model.User{})
		user := model.User{
			ID:       uuid.New().String(),
			Name:     c.PostForm("name"),
			Email:    c.PostForm("email"),
			Password: c.PostForm("password"),
		}

		// Check if the email is already registered
		if _, err := model.GetUserByEmail(db, user.Email); err == nil {
			log.Printf("Email already registered: %s", user.Email)
			c.JSON(http.StatusConflict, formatter.BadRequestResponse("Email already registered"))
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to hash password"))
			return
		}
		user.Password = string(hashedPassword)

		// Handle file upload
		file, err := c.FormFile("image")
		if err != nil {
			log.Printf("Error retrieving file: %v", err)
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Failed to get uploaded file"))
			return
		}

		// Generate a new filename for the image
		newFileName := uuid.New().String() + filepath.Ext(file.Filename)
		if err := c.SaveUploadedFile(file, "uploads/users/"+newFileName); err != nil {
			log.Printf("Error saving file: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to save uploaded file"))
			return
		}
		user.Image = newFileName

		// Insert user into the database
		if err := model.InsertUser(db, user); err != nil {
			log.Printf("Error inserting user into database: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to insert user into database"))
			return
		}

		c.JSON(http.StatusCreated, formatter.SuccessResponse(user))
	}
}

func LoginAuth(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")

		if email == "" || password == "" {
			log.Println("Email and password are required")
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("Email and password are required"))
			return
		}

		user, err := model.GetUserByEmail(db, email)
		if err != nil {
			log.Printf("Error retrieving user by email: %v", err)
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Invalid email or password"))
			} else {
				c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Error retrieving user"))
			}
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			log.Printf("Error comparing hash and password: %v", err)
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Invalid email or password"))
			return
		}

		// Generate JWT token using the helper function
		tokenString, err := generateJWT(user.ID, jwtKey)
		if err != nil {
			log.Printf("Error generating JWT token: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to generate token"))
			return
		}

		// Update user token in the database
		user.Token = &tokenString
		if err := model.UpdateUser(db, *user); err != nil {
			log.Printf("Error updating user token in database: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to update user token"))
			return
		}

		// Include server URL in the image link
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		user.Image = scheme + "://" + c.Request.Host + "/uploads/users/" + user.Image

		//password omitempty
		user.Password = ""

		// Return the user info along with the token
		c.JSON(http.StatusOK, formatter.SuccessResponse(user))
	}
}

func GetUserWithJWT(db *sql.DB, jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract JWT token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("Authorization token not provided")
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Authorization token not provided"))
			return
		}

		// Split the token type and the token value
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Println("Invalid Authorization token format")
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Invalid Authorization token format"))
			return
		}

		// Parse the JWT token
		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtKey), nil
		})

		if err != nil {
			log.Printf("Invalid token: %v", err)
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Invalid token"))
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID := claims["userID"].(string)
			user, err := model.GetUserID(db, userID)
			if err != nil {
				if err == sql.ErrNoRows {
					c.JSON(http.StatusNotFound, formatter.NotFoundResponse("User not found"))
				} else {
					c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Authorization token not provided"))
				}
				return
			}

			// Check if the token matches the token stored in the user table
			if user.Token == nil || *user.Token != tokenString {
				c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Token mismatch"))
				return
			}

			// Include server URL in the image link
			scheme := "http"
			if c.Request.TLS != nil {
				scheme = "https"
			}
			user.Image = scheme + "://" + c.Request.Host + "/uploads/users/" + user.Image

			// Omit token and password from the response
			user.Token = nil
			user.Password = ""

			c.JSON(http.StatusOK, formatter.SuccessResponse(user))
		} else {
			c.JSON(http.StatusUnauthorized, formatter.UnauthorizedResponse("Invalid token claims"))
		}
	}
}

func generateJWT(userID, jwtKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func DeleteUser(db *sql.DB, jwtKey string) gin.HandlerFunc {
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

		userIDToDelete := c.PostForm("user_id")
		if userIDToDelete == "" {
			c.JSON(http.StatusBadRequest, formatter.BadRequestResponse("User ID to delete is required"))
			return
		}

		// Retrieve user to get the image path
		user, err := model.GetUserID(db, userIDToDelete)
		if err != nil {
			log.Printf("Error retrieving user: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to retrieve user"))
			return
		}

		// Delete the user from the database
		if err := model.DeleteUser(db, userIDToDelete); err != nil {
			log.Printf("Error deleting user from database: %v", err)
			c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to delete user"))
			return
		}

		// Delete the user's image file
		if user.Image != "" {
			imagePath := "./uploads/users/" + user.Image
			if err := os.Remove(imagePath); err != nil {
				log.Printf("Error deleting image file: %v", err)
				c.JSON(http.StatusInternalServerError, formatter.InternalServerErrorResponse("Failed to delete user image"))
				return
			}
		}

		// Return success response
		c.JSON(http.StatusOK, formatter.SuccessResponse("User deleted successfully"))

	}
}

func ValidateToken(tokenString, jwtKey string, db *sql.DB) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the token signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtKey), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["userID"].(string)
		if !ok {
			return "", fmt.Errorf("userID claim is not a string")
		}
		// Retrieve the user from the database to check the token
		user, err := model.GetUserID(db, userID)
		if err != nil {
			return "", fmt.Errorf("failed to retrieve user: %v", err)
		}
		if user.Token != nil && *user.Token != tokenString {
			return "", fmt.Errorf("token does not match user's token")
		}
		return userID, nil
	}

	return "", fmt.Errorf("invalid token")
}
