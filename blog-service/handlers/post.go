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

	var newPost models.Post

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newPost)
	if err != nil {
		http.Error(w, "Greška pri parsiranju JSON-a: "+err.Error(), http.StatusBadRequest)
		return
	}

	newPost.UserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	newPost.Username = "djurdjevic_m"

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

	var posts []models.Post

	result := database.GORM_DB.Find(&posts)
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
