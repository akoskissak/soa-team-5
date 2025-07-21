package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"soa/blog-service/database"
	"soa/blog-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
	newPost.Username = "test_user"

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
