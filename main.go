package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"timeliner/internal/app"
	"timeliner/internal/database"
	"timeliner/internal/routes"

	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database connection
	dbConfig := database.ConfigFromEnv()
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}
	defer db.Close()

	log.Println("Application started")

	jwtSecret := os.Getenv("JWT_SECRET")
	app := app.NewApp(db, ctx, []byte(jwtSecret))

	// test user creation
	/*
		u, err := RegisterUser(models.Users, "test_1", "password")
		if err != nil {
			fmt.Printf("Error in user creation: %v", err)
		}
		fmt.Printf("User Created: %v", u)
	*/

	http.ListenAndServe(":8000", routes.NewRouter(app))
	fmt.Println("Server running")
}
