package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"timeliner/internal/database"
	"timeliner/internal/models"
	"timeliner/internal/routes"

	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"timeliner/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
)

type App struct {
	DBClient *pgxpool.Pool
	CTX      context.Context
}

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

	models := models.GetModels(db, ctx)
	jwtSecret := os.Getenv("JWT_SECRET")
	auth := services.NewAuthService(db, ctx, []byte(jwtSecret))

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

	// If I migrate to another file - http.ListenAndServe(":"+port, whatever.Router())
	// Routes
	router.Group(func(router chi.Router) {
		router.Use(jwtauth.Verifier(auth.JwtAuth))
		router.Use(jwtauth.Authenticator)
		router.Get("/authtest", func(w http.ResponseWriter, r *http.Request) {
			//w.Write([]byte(fmt.Sprintf("Working")))
			fmt.Println("Calling auth test")
			routes.AuthTest(w, r, models.Users)
		})
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		routes.Index(w, r)
	})

	router.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		routes.Login(w, r)
	})

	router.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		/*
			r.ParseForm()
			username := r.PostForm.Get("username")
			password := r.PostForm.Get("password")
			auth.LoginUser(models.Users, username, password)
		*/
		auth.LoginUser(w, r, models.Users)
	})

	router.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		routes.Logout(w, r)
	})
	router.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
		auth.LogOutUser(w, r)
	})
	router.Get("/incidents/new", func(w http.ResponseWriter, r *http.Request) {
		routes.NewIncident(w, r)
	})

	router.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		routes.GetUser(w, r, models.Users)
	})

	// API
	router.Get("/api/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		routes.GetUserById(w, r, models.Users)
	})

	// start server
	http.ListenAndServe(":8000", router)
}

func /*(auth *App)*/ RegisterUser(userModel models.UserModel, username, password string) (*models.User, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("username cannot be empty")
	}
	user, err := userModel.Insert(username, password)
	if err != nil {
		return nil, err
	}
	return user, nil
}
