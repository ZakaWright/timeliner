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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
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
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	// Routes
	router.Group(func(router chi.Router) {
		router.Use(jwtauth.Verifier(app.Auth.JwtAuth))
		router.Use(jwtauth.Authenticator)

		router.Get("/authtest", func(w http.ResponseWriter, r *http.Request) {
			//w.Write([]byte(fmt.Sprintf("Working")))
			fmt.Println("Calling auth test")
			routes.AuthTest(w, r, app.Models.Users)
		})
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		routes.Index(w, r)
	})

	router.Route("/login", func(router chi.Router) {
		router.Get("/", routes.Login)
		router.Post("/", func(w http.ResponseWriter, r *http.Request) {
			app.Auth.LoginUser(w, r, app.Models.Users)
		})
	})

	router.Route("/register", func(router chi.Router) {
		router.Get("/", routes.RegisterUser)
		router.Post("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Calling Register User")
			app.Auth.RegisterUser(w, r, app.Models.Users)
		})

	})

	router.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		routes.Logout(w, r)
	})
	router.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
		app.Auth.LogOutUser(w, r)
	})
	router.Get("/incidents/new", func(w http.ResponseWriter, r *http.Request) {
		routes.NewIncident(w, r)
	})

	router.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		routes.GetUser(w, r, &app.Models.Users)
	})

	// API
	router.Route("/api", (func(router chi.Router) {
		router.Route("/user", (func(router chi.Router) {
			router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				routes.GetUserById(w, r, app)
			})
		}))
	}))

	http.ListenAndServe(":8000", router)
}
