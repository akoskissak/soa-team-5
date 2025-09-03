package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	blogproto "api-gateway/proto/blog"
	followerproto "api-gateway/proto/follower"
	"soa/blog-service/database"
	"soa/blog-service/models"
	"soa/blog-service/utils"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const (
	MaxUploadSize = 5 * 1024 * 1024
	UploadDir     = "./static/uploads"
)

var followerClient followerproto.FollowerServiceClient

func InitFollowerClient(c followerproto.FollowerServiceClient) {
	followerClient = c
}

type BlogServer struct {
	blogproto.UnimplementedBlogServiceServer
}

func NewBlogServer() *BlogServer {
	return &BlogServer{}
}

func (s *BlogServer) CreatePost(ctx context.Context, req *blogproto.CreatePostRequest) (*blogproto.Post, error) {
	currentUsername, userId, err := utils.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Nevalidan token: %v", err)
	}

	if req.GetTitle() == "" || req.GetDescription() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Naslov i opis su obavezni.")
	}

	newPost := models.Post{
		UserID:      userId,
		Username:    currentUsername,
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		CreatedAt:   time.Now(),
		ImageURLs:   req.GetImageUrls(),
	}

	result := database.GORM_DB.Create(&newPost)
	if result.Error != nil {
		log.Printf("Greška pri čuvanju posta u bazu: %v", result.Error)
		return nil, status.Errorf(codes.Internal, "Greška servera pri kreiranju posta.")
	}

	protoPost := convertPostToProto(&newPost)
	fmt.Printf("Novi post kreiran: %+v\n", protoPost)
	return protoPost, nil
}

func HandleImageUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metoda nije dozvoljena.", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(MaxUploadSize)
	if err != nil {
		http.Error(w, "Fajl je prevelik. Maksimalna veličina je 5MB.", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Fajl nije pronađen.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if _, err := os.Stat(UploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(UploadDir, 0755); err != nil {
			log.Printf("Greška pri kreiranju direktorijuma %s: %v", UploadDir, err)
			http.Error(w, "Greška servera pri kreiranju foldera za upload.", http.StatusInternalServerError)
			return
		}
	}

	fileExtension := filepath.Ext(fileHeader.Filename)
	newFileName := uuid.New().String() + fileExtension
	filePath := filepath.Join(UploadDir, newFileName)

	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("Greška pri kreiranju fajla na disku: %v", err)
		http.Error(w, "Greška servera pri čuvanju fajla.", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		log.Printf("Greška pri kopiranju fajla na disk: %v", err)
		http.Error(w, "Greška servera pri čuvanju fajla.", http.StatusInternalServerError)
		return
	}

	imageURL := fmt.Sprintf("/uploads/%s", newFileName)
	log.Printf("Slika uspešno uploadovana: %s", imageURL)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"imageUrl": "%s"}`, imageURL)))
}

func (s *BlogServer) GetPosts(ctx context.Context, req *blogproto.GetPostsRequest) (*blogproto.GetPostsResponse, error) {
	currentUsername, _, err := utils.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Nevalidan token: %v", err)
	}

	followerResp, err := followerClient.GetFollowing(ctx, &followerproto.GetFollowingRequest{Username: currentUsername})
	if err != nil {
		log.Printf("Greška pri dohvatanju pracenih korisnika: %v", err)
		return nil, status.Errorf(codes.Internal, "Greška pri dohvatanju pracenih korisnika.")
	}

	following := append(followerResp.Following, currentUsername)

	var posts []models.Post
	result := database.GORM_DB.Where("username IN ?", following).Order("created_at DESC").Find(&posts)
	if result.Error != nil {
		log.Printf("Greška pri dohvatanju postova iz baze: %v", result.Error)
		return nil, status.Errorf(codes.Internal, "Greška servera pri dohvatanju postova.")
	}

	protoPosts := make([]*blogproto.Post, len(posts))
	for i, post := range posts {
		protoPosts[i] = convertPostToProto(&post)
	}

	fmt.Printf("Dohvaćeno %d postova.\n", len(protoPosts))
	return &blogproto.GetPostsResponse{Posts: protoPosts}, nil
}

func (s *BlogServer) GetPostByID(ctx context.Context, req *blogproto.GetPostByIDRequest) (*blogproto.Post, error) {
	postID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Neispravan format ID-a posta.")
	}

	var post models.Post
	result := database.GORM_DB.First(&post, postID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "Post nije pronađen.")
		}
		log.Printf("Greška pri dohvatanju posta po ID-ju: %v", result.Error)
		return nil, status.Errorf(codes.Internal, "Greška servera pri dohvatanju posta.")
	}

	protoPost := convertPostToProto(&post)
	fmt.Printf("Dohvaćen post sa ID: %s.\n", postID.String())
	return protoPost, nil
}

func (s *BlogServer) ToggleLike(ctx context.Context, req *blogproto.ToggleLikeRequest) (*blogproto.ToggleLikeResponse, error) {
	fmt.Printf("ToggleLike - Primljen zahtev. PostID: %s, UserID: %s\n", req.GetPostId(), req.GetUserId())

	postID, err := uuid.Parse(req.GetPostId())
	if err != nil {
		fmt.Printf("ToggleLike - Greška pri parsiranju PostID-a: %v\n", err)
		return nil, status.Errorf(codes.InvalidArgument, "Neispravan format ID-a posta.")
	}
	fmt.Printf("ToggleLike - PostID uspešno parsiran: %s\n", postID.String())

	userID := req.GetUserId()
	fmt.Printf("ToggleLike - UserID je string: %s\n", userID)

	var existingLike models.Like
	result := database.GORM_DB.Where("post_id = ? AND user_id = ?", postID, userID).First(&existingLike)

	var likesCount int32
	var responseStatus string

	switch result.Error {
	case nil:
		deleteResult := database.GORM_DB.Delete(&existingLike)
		if deleteResult.Error != nil {
			log.Printf("Greška pri brisanju lajka: %v", deleteResult.Error)
			return nil, status.Errorf(codes.Internal, "Greška servera pri uklanjanju lajka.")
		}
		database.GORM_DB.Model(&models.Post{}).Where("id = ?", postID).Update("likes_count", gorm.Expr("likes_count - ?", 1))

		var updatedPost models.Post
		database.GORM_DB.Select("likes_count").First(&updatedPost, postID)
		likesCount = int32(updatedPost.LikesCount)
		responseStatus = "Lajk uklonjen"
	case gorm.ErrRecordNotFound:
		newLike := models.Like{
			PostID: postID,
			UserID: userID,
		}
		createResult := database.GORM_DB.Create(&newLike)
		if createResult.Error != nil {
			log.Printf("Greška pri dodavanju lajka: %v", createResult.Error)
			return nil, status.Errorf(codes.Internal, "Greška servera pri dodavanju lajka.")
		}
		database.GORM_DB.Model(&models.Post{}).Where("id = ?", postID).Update("likes_count", gorm.Expr("likes_count + ?", 1))

		var updatedPost models.Post
		database.GORM_DB.Select("likes_count").First(&updatedPost, postID)
		likesCount = int32(updatedPost.LikesCount)
		responseStatus = "Lajk dodat"
	default:
		log.Printf("Greška pri proveri postojećeg lajka: %v", result.Error)
		return nil, status.Errorf(codes.Internal, "Greška servera pri obradi lajka.")
	}

	fmt.Printf("Status lajka: %s, lajkova: %d\n", responseStatus, likesCount)
	return &blogproto.ToggleLikeResponse{Status: responseStatus, LikesCount: likesCount}, nil
}

func (s *BlogServer) AddCommentToPost(ctx context.Context, req *blogproto.AddCommentToPostRequest) (*blogproto.Comment, error) {
	fmt.Printf("AddCommentToPost - Primljen zahtev. PostID: %s, UserID: %s\n", req.GetPostId(), req.GetUserId())

	postID, err := uuid.Parse(req.GetPostId())
	if err != nil {
		fmt.Printf("AddCommentToPost - Greška pri parsiranju PostID-a: %v\n", err)
		return nil, status.Errorf(codes.InvalidArgument, "Neispravan format ID-a posta.")
	}
	fmt.Printf("AddCommentToPost - PostID uspešno parsiran: %s\n", postID.String())

	parsedUserID := req.GetUserId()
	fmt.Printf("AddCommentToPost - UserID je string: %s\n", parsedUserID)

	newComment := models.Comment{
		PostID:    postID,
		UserID:    parsedUserID,
		Username:  req.Username,
		Text:      req.GetText(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if newComment.Text == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Tekst komentara je obavezan.")
	}

	result := database.GORM_DB.Create(&newComment)
	if result.Error != nil {
		log.Printf("Greška pri čuvanju komentara u bazu: %v", result.Error)
		return nil, status.Errorf(codes.Internal, "Greška servera pri kreiranju komentara.")
	}

	protoComment := convertCommentToProto(&newComment)
	fmt.Printf("Novi komentar kreiran za post %s: %+v\n", postID.String(), protoComment)
	return protoComment, nil
}

func (s *BlogServer) GetCommentsForPost(ctx context.Context, req *blogproto.GetCommentsForPostRequest) (*blogproto.GetCommentsForPostResponse, error) {
	fmt.Printf("GetCommentsForPost - Primljen zahtev. PostID: %s\n", req.GetPostId())

	postID, err := uuid.Parse(req.GetPostId())
	if err != nil {
		fmt.Printf("GetCommentsForPost - Greška pri parsiranju PostID-a: %v\n", err)
		return nil, status.Errorf(codes.InvalidArgument, "Neispravan format ID-a posta.")
	}
	fmt.Printf("GetCommentsForPost - PostID uspešno parsiran: %s\n", postID.String())

	var comments []models.Comment
	result := database.GORM_DB.Where("post_id = ?", postID).Order("created_at asc").Find(&comments)
	if result.Error != nil {
		log.Printf("Greška pri dohvatanju komentara za post %s: %v", postID.String(), result.Error)
		return nil, status.Errorf(codes.Internal, "Greška servera pri dohvatanju komentara.")
	}

	protoComments := make([]*blogproto.Comment, len(comments))
	for i, comment := range comments {
		protoComments[i] = convertCommentToProto(&comment)
	}

	fmt.Printf("Dohvaćeno %d komentara za post %s.\n", len(protoComments), postID.String())
	return &blogproto.GetCommentsForPostResponse{Comments: protoComments}, nil
}

func convertPostToProto(post *models.Post) *blogproto.Post {
	return &blogproto.Post{
		Id:          post.ID.String(),
		UserId:      post.UserID,
		Username:    post.Username,
		Title:       post.Title,
		Description: post.Description,
		CreatedAt:   post.CreatedAt.Format(time.RFC3339),
		ImageUrls:   post.ImageURLs,
		LikesCount:  int32(post.LikesCount),
	}
}

func convertCommentToProto(comment *models.Comment) *blogproto.Comment {
	return &blogproto.Comment{
		Id:        comment.ID.String(),
		PostId:    comment.PostID.String(),
		UserId:    comment.UserID,
		Username:  comment.Username,
		Text:      comment.Text,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
	}
}
