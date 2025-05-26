package routes

import (
	"net/http"
	"strconv"

	"timeliner/internal/models"
	"timeliner/internal/services"
	"timeliner/web/components"

	"github.com/go-chi/chi/v5"
)

func Index(w http.ResponseWriter, r *http.Request) {
	component := components.Index()
	component.Render(r.Context(), w)
}

func Login(w http.ResponseWriter, r *http.Request) {
	component := components.Login()
	component.Render(r.Context(), w)
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	component := components.RegisterUser()
	component.Render(r.Context(), w)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	component := components.LogOut()
	component.Render(r.Context(), w)
}

func GetUser(w http.ResponseWriter, r *http.Request, user_model *models.UserModel) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	user, err := user_model.GetByID(intID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	component := components.User(*user)
	component.Render(r.Context(), w)

}

func NewIncident(w http.ResponseWriter, r *http.Request) {
	component := components.NewIncident()
	component.Render(r.Context(), w)
}

func AuthTest(w http.ResponseWriter, r *http.Request, user_model models.UserModel) {
	// get user
	/*
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
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
				intID = parsedID
			default:
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			user, err := user_model.GetByID(intID)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			w.Write([]byte(fmt.Sprintf("Working %v", user)))
		}
	*/
	user, err := services.GetUser(r, user_model)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	component := components.User(*user)
	component.Render(r.Context(), w)
}
