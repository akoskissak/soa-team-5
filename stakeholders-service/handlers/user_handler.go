package handlers

import (
	"context"
	"fmt"
	"net/http"
	"stakeholders-service/db"
	"stakeholders-service/models"
	"stakeholders-service/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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
	err = collection.FindOne(ctx, bson.M{"username": input.Username}).Decode(&user)
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
		"token": token,
	})
}


func GetAllUsers(c *gin.Context){
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
        "_id":      1,
        "username": 1,
        "email":    1,
        "role":     1,
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