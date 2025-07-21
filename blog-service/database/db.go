package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"soa/blog-service/models"
)

var GORM_DB *gorm.DB

func InitDB() {
	connStr := "user=postgres password=super host=localhost port=5432 dbname=blog_db sslmode=disable"

	var err error
	GORM_DB, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("Greška pri povezivanju sa bazom podataka koristeći GORM: %v", err)
	}

	fmt.Println("Uspješno povezano sa PostgreSQL bazom podataka koristeći GORM!")

	err = GORM_DB.AutoMigrate(&models.Post{})
	if err != nil {
		log.Fatalf("Greška pri automatskoj migraciji šeme baze podataka: %v", err)
	}
	fmt.Println("Migracija tabele 'posts' uspješno završena (GORM AutoMigrate).")

	sqlDB, err := GORM_DB.DB()
	if err != nil {
		log.Fatalf("Greška pri dobijanju underlying *sql.DB iz GORM-a: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
}

func CloseDB() {
	if GORM_DB != nil {
		sqlDB, err := GORM_DB.DB()
		if err != nil {
			log.Printf("Error getting underlying *sql.DB to close: %v", err)
		} else {
			err = sqlDB.Close()
			if err != nil {
				log.Printf("Error closing GORM database connection: %v", err)
			}
			fmt.Println("GORM konekcija sa bazom podataka zatvorena.")
		}
	}
}
