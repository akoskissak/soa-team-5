package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"soa/blog-service/database"
	"soa/blog-service/models"
	"soa/blog-service/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	MaxUploadSize = 5 * 1024 * 1024
	UploadDir     = "./static/uploads"
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metoda nije dozvoljena", http.StatusMethodNotAllowed)
		return
	}

	// jwt parse
	currentUsername, userId, err := utils.GetClaimsFromJWT(r)
	if err != nil {
		http.Error(w, "Nevalidan token", http.StatusUnauthorized)
		return
	}

	var newPost models.Post

	decoder := json.NewDecoder(r.Body)
	parseErr := decoder.Decode(&newPost)
	if parseErr != nil {
		http.Error(w, "Greška pri parsiranju JSON-a: " + parseErr.Error(), http.StatusBadRequest)
		return
	}

	newPost.UserID = userId
	newPost.Username = currentUsername

	if newPost.Title == "" || newPost.Description == "" {
		http.Error(w, "Naslov i opis su obavezni.", http.StatusBadRequest)
		return
	}

	result := database.GORM_DB.Create(&newPost)
	if result.Error != nil {
		log.Printf("Greška pri čuvanju posta u bazu: %v", result.Error)
		http.Error(w, "Greška servera pri kreiranju posta.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newPost)
	fmt.Printf("Novi post kreiran: %+v\n", newPost)
}

func UploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metoda nije dozvoljena", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		http.Error(w, "Fajl je prevelik. Maksimalna veličina je 5MB.", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Greška pri dohvatanju fajla: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := handler.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		http.Error(w, "Nevažeći tip fajla. Dozvoljeni su samo JPEG, PNG, GIF.", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(UploadDir); os.IsNotExist(err) {
		err = os.MkdirAll(UploadDir, 0755)
		if err != nil {
			log.Printf("Greška pri kreiranju direktorijuma %s: %v", UploadDir, err)
			http.Error(w, "Greška servera pri kreiranju foldera za upload.", http.StatusInternalServerError)
			return
		}
	}

	fileExtension := filepath.Ext(handler.Filename)
	newFileName := uuid.New().String() + fileExtension
	filePath := filepath.Join(UploadDir, newFileName)

	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("Greška pri kreiranju fajla na disku: %v", err)
		http.Error(w, "Greška servera pri čuvanju fajla.", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Greška pri kopiranju fajla: %v", err)
		http.Error(w, "Greška servera pri čuvanju fajla.", http.StatusInternalServerError)
		return
	}

	imageURL := fmt.Sprintf("/uploads/%s", newFileName)

	response := map[string]string{"imageUrl": imageURL}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("Slika uspešno uploadovana: %s", imageURL)
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metoda nije dozvoljena", http.StatusMethodNotAllowed)
		return
	}
	
	// jwt parse
	currentUsername, _, err := utils.GetClaimsFromJWT(r)
	if err != nil {
		http.Error(w, "Nevalidan token", http.StatusUnauthorized)
		return
	}

	following, err := GetFollowing(currentUsername)
	if err != nil {
		http.Error(w, "Greska pri dohvatanju pracenih korisnika", http.StatusInternalServerError)
		return
	}

	following = append(following, currentUsername)

	var posts []models.Post

	result := database.GORM_DB.Where("username IN ?", following).Order("created_at DESC").Find(&posts)
	if result.Error != nil {
		log.Printf("Greška pri dohvatanju postova iz baze: %v", result.Error)
		http.Error(w, "Greška servera pri dohvatanju postova.", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(posts)
	fmt.Printf("Dohvaćeno %d postova.\n", len(posts))
}

// pomocna funkcija koji radi preko HTTP REST, prebacicemo na gRPC
func GetFollowing(username string) ([]string, error) {
	url := fmt.Sprintf("http://follower-service:8082/api/following/%s", username)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("followers service returned status %d", resp.StatusCode)
	}

	var result models.FollowingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Following, nil
}

func GetPostByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metoda nije dozvoljena", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "ID posta je obavezan.", http.StatusBadRequest)
		return
	}

	postID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Neispravan format ID-a posta. Mora biti validan UUID.", http.StatusBadRequest)
		return
	}

	var post models.Post

	result := database.GORM_DB.First(&post, postID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Post nije pronađen.", http.StatusNotFound)
			return
		}
		log.Printf("Greška pri dohvatanju posta po ID-ju: %v", result.Error)
		http.Error(w, "Greška servera pri dohvatanju posta.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(post)
	fmt.Printf("Dohvaćen post sa ID: %s.\n", postID.String())
}

func ToggleLike(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metoda nije dozvoljena", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "ID posta je obavezan.", http.StatusBadRequest)
		return
	}
	postID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Neispravan format ID-a posta. Mora biti validan UUID.", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		UserID uuid.UUID `json:"userId"`
	}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&requestBody)
	if err != nil {
		log.Printf("Greška pri dekodiranju ToggleLike JSON-a: %v", err)
		http.Error(w, "Greška pri parsiranju JSON-a (očekivan userId): "+err.Error(), http.StatusBadRequest)
		return
	}
	if requestBody.UserID == uuid.Nil {
		http.Error(w, "UserID je obavezan za lajk.", http.StatusBadRequest)
		return
	}
	userID := requestBody.UserID

	var existingLike models.Like
	result := database.GORM_DB.Where("post_id = ? AND user_id = ?", postID, userID).First(&existingLike)

	switch result.Error {
	case nil:
		deleteResult := database.GORM_DB.Delete(&existingLike)
		if deleteResult.Error != nil {
			log.Printf("Greška pri brisanju lajka: %v", deleteResult.Error)
			http.Error(w, "Greška servera pri uklanjanju lajka.", http.StatusInternalServerError)
			return
		}
		database.GORM_DB.Model(&models.Post{}).Where("id = ?", postID).Update("likes_count", gorm.Expr("likes_count - ?", 1))
		fmt.Printf("Lajk uklonjen za post %s od korisnika %s\n", postID.String(), userID.String())
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Lajk uklonjen"})
		return
	case gorm.ErrRecordNotFound:
		newLike := models.Like{
			PostID: postID,
			UserID: userID,
		}
		createResult := database.GORM_DB.Create(&newLike)
		if createResult.Error != nil {
			log.Printf("Greška pri dodavanju lajka: %v", createResult.Error)
			http.Error(w, "Greška servera pri dodavanju lajka.", http.StatusInternalServerError)
			return
		}
		database.GORM_DB.Model(&models.Post{}).Where("id = ?", postID).Update("likes_count", gorm.Expr("likes_count + ?", 1))
		fmt.Printf("Lajk dodat za post %s od korisnika %s\n", postID.String(), userID.String())
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Lajk dodat"})
		return
	default:
		log.Printf("Greška pri proveri postojećeg lajka: %v", result.Error)
		http.Error(w, "Greška servera pri obradi lajka.", http.StatusInternalServerError)
		return
	}
}

func GetCommentsForPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metoda nije dozvoljena", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "ID posta je obavezan.", http.StatusBadRequest)
		return
	}
	postID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Neispravan format ID-a posta. Mora biti validan UUID.", http.StatusBadRequest)
		return
	}

	var comments []models.Comment
	result := database.GORM_DB.Where("post_id = ?", postID).Order("created_at asc").Find(&comments)
	if result.Error != nil {
		log.Printf("Greška pri dohvatanju komentara za post %s: %v", postID.String(), result.Error)
		http.Error(w, "Greška servera pri dohvatanju komentara.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comments)
	fmt.Printf("Dohvaćeno %d komentara za post %s.\n", len(comments), postID.String())
}

func AddCommentToPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metoda nije dozvoljena", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id") // ID posta iz URL putanje
	if idStr == "" {
		http.Error(w, "ID posta je obavezan.", http.StatusBadRequest)
		return
	}
	postID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Neispravan format ID-a posta. Mora biti validan UUID.", http.StatusBadRequest)
		return
	}

	var newComment models.Comment
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&newComment)
	if err != nil {
		http.Error(w, "Greška pri parsiranju JSON-a: "+err.Error(), http.StatusBadRequest)
		return
	}

	newComment.PostID = postID
	if newComment.Text == "" || newComment.UserID == uuid.Nil || newComment.Username == "" {
		http.Error(w, "Tekst komentara, UserID i Username su obavezni.", http.StatusBadRequest)
		return
	}
	newComment.CreatedAt = time.Now()
	newComment.UpdatedAt = time.Now()

	result := database.GORM_DB.Create(&newComment)
	if result.Error != nil {
		log.Printf("Greška pri čuvanju komentara u bazu: %v", result.Error)
		http.Error(w, "Greška servera pri kreiranju komentara.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newComment)
	fmt.Printf("Novi komentar kreiran za post %s: %+v\n", postID.String(), newComment)
}