package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"stakeholders-service/models"
	"stakeholders-service/utils"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	stakeproto "api-gateway/proto/stakeholders"
)

type StakeholdersServer struct {
	stakeproto.UnimplementedStakeholdersServiceServer
	mongoClient *mongo.Client
}

func NewStakeholdersServer(mongoClient *mongo.Client) *StakeholdersServer {
	return &StakeholdersServer{
		mongoClient: mongoClient,
	}
}

func (s *StakeholdersServer) Register(ctx context.Context, req *stakeproto.RegisterRequest) (*stakeproto.RegisterResponse, error) {
	input := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     models.Role(req.Role),
	}

	if input.Username == "" || input.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username and password are required")
	}

	if input.Role != models.RoleTourist && input.Role != models.RoleGuide {
		return nil, status.Errorf(codes.InvalidArgument, "role must be tourist or guide")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to process registration")
	}
	input.Password = string(hashedPassword)

	collection := s.mongoClient.Database("stakeholders").Collection("users")

	_, err = collection.InsertOne(ctx, input)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Printf("Registration failed for username '%s': user already exists (duplicate key error)", input.Username)
			return nil, status.Errorf(codes.AlreadyExists, "user already exists")
		}

		log.Printf("MongoDB insert error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	log.Printf("User registered successfully")
	return &stakeproto.RegisterResponse{}, nil
}

func (s *StakeholdersServer) Login(ctx context.Context, req *stakeproto.LoginRequest) (*stakeproto.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username and password are required")
	}

	collection := s.mongoClient.Database("stakeholders").Collection("users")

	var user models.User
	err := collection.FindOne(ctx, bson.M{"username": req.Username, "is_blocked": false}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, status.Errorf(codes.NotFound, "invalid credentials")
	}
	if err != nil {
		log.Printf("MongoDB find error during login: %v", err)
		return nil, status.Errorf(codes.Internal, "database error")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	token, err := utils.GenerateJWT(user.Username, string(user.Role), user.ID)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	return &stakeproto.LoginResponse{
		AccessToken: token,
	}, nil
}

type UserResponse struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	Password  string             `json:"password"`
	Role      string             `json:"role"`
	IsBlocked bool               `json:"isBlocked"`
}

func (s *StakeholdersServer) GetAllUsers(ctx context.Context, req *stakeproto.GetAllUsersRequest) (*stakeproto.GetAllUsersResponse, error) {
	collection := s.mongoClient.Database("stakeholders").Collection("users")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filter := bson.M{}
	projection := bson.M{
		"_id":        1,
		"username":   1,
		"email":      1,
		"password":   1,
		"role":       1,
		"is_blocked": 1,
	}

	findOptions := options.Find().SetProjection(projection)

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("MongoDB find error: %v", err)
		return nil, status.Errorf(codes.Internal, "could not fetch users")
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		log.Printf("Error decoding users: %v", err)
		return nil, status.Errorf(codes.Internal, "error decoding users")
	}

	var adminUsersProto []*stakeproto.User
	for _, u := range users {
		adminUsersProto = append(adminUsersProto, &stakeproto.User{
			Id:        u.ID.Hex(),
			Username:  u.Username,
			Email:     u.Email,
			Password:  u.Password,
			Role:      string(u.Role),
			IsBlocked: u.IsBlocked,
		})
	}

	return &stakeproto.GetAllUsersResponse{Users: adminUsersProto}, nil
}

func (s *StakeholdersServer) BlockUser(ctx context.Context, req *stakeproto.BlockUserRequest) (*stakeproto.BlockUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "userId is required")
	}

	objID, err := primitive.ObjectIDFromHex(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid userId format")
	}

	collection := s.mongoClient.Database("stakeholders").Collection("users")
	update := bson.M{"$set": bson.M{"is_blocked": req.Block}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedUser models.User
	err = collection.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update, opts).Decode(&updatedUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		log.Printf("Failed to update user: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update user")
	}

	return &stakeproto.BlockUserResponse{Status: "User blocked/unblocked successfully"}, nil
}

func (s *StakeholdersServer) GetProfile(ctx context.Context, req *stakeproto.GetProfileRequest) (*stakeproto.UserProfileResponse, error) {
	claims, err := utils.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized: %v", err)
	}

	role, ok := claims["role"].(string)
	if !ok || role == "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "only guides and tourists can access this route")
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token claims: username not found")
	}

	collection := s.mongoClient.Database("stakeholders").Collection("users")

	projection := bson.M{
		"profile":  1,
		"username": 1,
	}

	var result models.User

	err = collection.FindOne(ctx, bson.M{"username": username}, options.FindOne().SetProjection(projection)).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, status.Errorf(codes.NotFound, "profile not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "database error")
	}

	response := utils.MapToUserProfileResponse(result)
	return response, nil
}

func (s *StakeholdersServer) GetProfileByUsername(ctx context.Context, req *stakeproto.GetProfileByUsernameRequest) (*stakeproto.UserProfileResponse, error) {
	username := req.Username

	if username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username is required")
	}

	collection := s.mongoClient.Database("stakeholders").Collection("users")
	projection := bson.M{"profile": 1, "username": 1}

	var result models.User
	err := collection.FindOne(
		ctx,
		bson.M{"username": username},
		options.FindOne().SetProjection(projection),
	).Decode(&result)

	if err == mongo.ErrNoDocuments {
		return nil, status.Errorf(codes.NotFound, "profile not found")
	} else if err != nil {
		log.Printf("Database error fetching profile by username: %v", err)
		return nil, status.Errorf(codes.Internal, "database error")
	}

	response := utils.MapToUserProfileResponse(result)
	return response, nil
}

func (s *StakeholdersServer) UpdateProfile(ctx context.Context, req *stakeproto.UpdateProfileRequest) (*stakeproto.UpdateProfileResponse, error) {
	claims, err := utils.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized: %v", err)
	}

	role, ok := claims["role"].(string)
	if !ok || role == "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "only guides and tourists can access this route")
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token claims: username not found")
	}

	var updatedProfile models.UserProfile
	if req.Profile != nil {
		updatedProfile.FirstName = req.Profile.FirstName
		updatedProfile.LastName = req.Profile.LastName
		updatedProfile.ProfilePicture = req.Profile.ProfilePicture
		updatedProfile.Biography = req.Profile.Biography
		updatedProfile.Motto = req.Profile.Motto
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "profile data is required")
	}

	if updatedProfile.ProfilePicture != "" && strings.HasPrefix(updatedProfile.ProfilePicture, "data:image/") {
		imageURL, err := saveBase64Image(updatedProfile.ProfilePicture)
		if err != nil {
			log.Printf("Error saving image: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to save image: %v", err)
		}
		updatedProfile.ProfilePicture = imageURL
	}

	collection := s.mongoClient.Database("stakeholders").Collection("users")
	update := bson.M{"$set": bson.M{"profile": updatedProfile}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var userAfterUpdate models.User
	err = collection.FindOneAndUpdate(ctx, bson.M{"username": username}, update, opts).Decode(&userAfterUpdate)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		log.Printf("Failed to update profile: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update profile: %v", err)
	}

	return &stakeproto.UpdateProfileResponse{Status: "Profile updated successfully"}, nil
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
