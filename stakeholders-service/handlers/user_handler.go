package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"stakeholders-service/db"
	"stakeholders-service/models"
	"stakeholders-service/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var input models.User

	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// provere ali bi bolje bilo da se koristi postgresql
	if input.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	fmt.Println("Pssword je: ", input.Password)
	if input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	if input.Role != models.RoleTourist && input.Role != models.RoleGuide {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role must be tourist or guide"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	input.Password = string(hashedPassword)

	collection := db.MongoClient.Database("stakeholders").Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existing models.User
	err = collection.FindOne(ctx, bson.M{"username": input.Username}).Decode(&existing)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}
	if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	_, err = collection.InsertOne(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func Login(c *gin.Context) {
	var input models.User

	// Parsiraj JSON telo zahteva
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if input.Username == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	collection := db.MongoClient.Database("stakeholders").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err = collection.FindOne(ctx, bson.M{"username": input.Username, "is_blocked": false}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := utils.GenerateJWT(user.Username, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Login successful",
		"username": user.Username,
		"role":     user.Role,
		"token":    token,
	})
}

func GetAllUsers(c *gin.Context) {
	claims, err := utils.VerifyJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims["role"] != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can access this route"})
		return
	}

	collection := db.MongoClient.Database("stakeholders").Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}

	projection := bson.M{
		"_id":        1,
		"username":   1,
		"email":      1,
		"role":       1,
		"is_blocked": 1,
	}

	findOptions := options.Find().SetProjection(projection)

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch users"})
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func BlockUser(c *gin.Context) {
	claims, err := utils.VerifyJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims["role"] != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can access this route"})
		return
	}

	var requestBody struct {
		UserID    string `json:"userId"`
		BlockUser bool   `json:"block"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(requestBody.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	collection := db.MongoClient.Database("stakeholders").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{"is_blocked": requestBody.BlockUser},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedUser models.User
	err = collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, update, opts).Decode(&updatedUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

func GetProfile(c *gin.Context) {
	claims, err := utils.VerifyJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	if claims["role"] == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only guides and tourists can access this route"})
		return
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token claims"})
		return
	}

	collection := db.MongoClient.Database("stakeholders").Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projection := bson.M{
		"profile":    1,
		"_id":        0,
		"is_blocked": 1,
	}

	var result struct {
		Profile models.UserProfile `bson:"profile" json:"profile"`
	}

	err = collection.FindOne(ctx, bson.M{"username": username}, options.FindOne().SetProjection(projection)).Decode(&result)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, result.Profile)
}

func UpdateProfile(c *gin.Context) {
	claims, err := utils.VerifyJWT(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	userRole, ok := claims["role"].(string)
	if !ok || userRole == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token claims: role missing"})
		return
	}

	if userRole == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin cannot update profile via this route"})
		return
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token claims: username missing"})
		return
	}

	var updatedProfile models.UserProfile
	if err := c.ShouldBindJSON(&updatedProfile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if updatedProfile.ProfilePicture != "" && strings.HasPrefix(updatedProfile.ProfilePicture, "data:image/") {
		imageURL, err := saveBase64Image(updatedProfile.ProfilePicture)
		if err != nil {
			log.Printf("Error saving image: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image: " + err.Error()})
			return
		}
		updatedProfile.ProfilePicture = imageURL
	}

	collection := db.MongoClient.Database("stakeholders").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"profile": updatedProfile,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var userAfterUpdate models.User
	err = collection.FindOneAndUpdate(ctx, bson.M{"username": username}, update, opts).Decode(&userAfterUpdate)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, userAfterUpdate.Profile)
}

func saveBase64Image(base64String string) (string, error) {
	parts := strings.SplitN(base64String, ",", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid base64 string format")
	}

	meta := parts[0]
	data := parts[1]

	contentType := strings.TrimPrefix(meta, "data:")
	contentType = strings.SplitN(contentType, ";", 2)[0]

	var extension string
	switch contentType {
	case "image/png":
		extension = ".png"
	case "image/jpeg", "image/jpg":
		extension = ".jpg"
	case "image/gif":
		extension = ".gif"
	default:
		return "", fmt.Errorf("unsupported image content type: %s", contentType)
	}

	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 string: %w", err)
	}

	fileName := uuid.New().String() + extension
	filePath := filepath.Join("static", "uploads", fileName)

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directories: %w", err)
	}

	err = os.WriteFile(filePath, decodedData, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save image file: %w", err)
	}

	return fmt.Sprintf("/uploads/%s", fileName), nil
}
