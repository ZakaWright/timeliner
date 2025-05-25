package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"timeliner/internal/models"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-chi/jwtauth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthService struct {
	DBClient         *pgxpool.Pool
	CTX              context.Context
	jwtSecret        []byte
	tokenExperiation time.Duration
	JwtAuth          *jwtauth.JWTAuth
}

func NewAuthService(dbClient *pgxpool.Pool, ctx context.Context, jwtSecret []byte) *AuthService {
	return &AuthService{
		DBClient:         dbClient,
		CTX:              ctx,
		jwtSecret:        jwtSecret,
		tokenExperiation: 24 * time.Hour,
		JwtAuth:          jwtauth.New("HS256", []byte(jwtSecret), nil),
	}
}

func (auth AuthService) MakeToken(id int64) string {
	_, tokenString, _ := auth.JwtAuth.Encode(map[string]interface{}{"user_id": id})
	return tokenString

}

func (auth AuthService) LoginUser(w http.ResponseWriter, r *http.Request, userModel models.UserModel) {
	r.ParseForm()
	username := r.PostForm.Get("login-username")
	password := r.PostForm.Get("login-password")

	if strings.TrimSpace(username) == "" {
		//return nil, errors.New("username cannot be empty")
		http.Error(w, "Username cannot be empty", http.StatusBadRequest)
	}
	if strings.TrimSpace(password) == "" {
		//return nil, errors.New("password cannot be empty")
		http.Error(w, "Password cannot be empty", http.StatusBadRequest)
	}
	query := `
		SELECT user_id, username, password_hash, is_active, created_at 
		FROM users 
		WHERE username = $1
	`
	var user models.User
	err := userModel.DB.QueryRow(userModel.CTX, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Username or Password is Incorrect", http.StatusBadRequest)
		}
		http.Error(w, "Something went wrong", http.StatusBadRequest)
	}
	if !user.IsActive {
		http.Error(w, "User is disabled", http.StatusBadRequest)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		http.Error(w, "Username or Password is Incorrect", http.StatusBadRequest)
	}
	// clear password from user
	user.PasswordHash = ""

	token := auth.MakeToken(user.ID)

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
		// HTTPS only
		//Secure: true,
		Name:  "jwt",
		Value: token,
	})
	http.Redirect(w, r, "/authtest", http.StatusSeeOther)

}

func (auth AuthService) LogOutUser(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Logging Out\n")
	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
		// HTTPS only
		//Secure: true,
		Name:  "jwt",
		Value: "",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func GetUser(r *http.Request, user_model models.UserModel) (*models.User, error) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if claims["user_id"] != nil {
		id := claims["user_id"]
		var intID int64
		switch v := id.(type) {
		case float64:
			intID = int64(v)
		case string:
			parsedID, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				//http.Redirect(w, r, "/login", http.StatusSeeOther)
				return nil, errors.New("invalid user ID")
			}
			intID = parsedID
		default:
			//http.Redirect(w, r, "/login", http.StatusSeeOther)
			return nil, errors.New("invalid user ID")
		}
		user, err := user_model.GetByID(intID)
		if err != nil {
			//http.Redirect(w, r, "/login", http.StatusSeeOther)
			return nil, errors.New("invalid user")
		}
		return user, nil

	}
	return nil, errors.New("invalid user")

}
