package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"presto_tou_service/constants"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"presto_tou_service/handler"
	"presto_tou_service/repository"
	"presto_tou_service/router"
	"presto_tou_service/service"
)

// @title           Presto TOU Service API
// @version         1.0
// @description     This is a time-of-use pricing service for EV chargers.
// @host            localhost:8080
// @BasePath        /
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	db := initDB()
	defer db.Close()
	repo := repository.NewPostgresRepo(db)
	svc := service.NewPricingService(repo)
	httpHandler := handler.NewHttpHandler(svc)

	mux := router.NewRouter(httpHandler)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func initDB() *sql.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == constants.Empty {
		log.Fatalf("DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}
