package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"timeliner/internal/app"

	"github.com/go-chi/chi/v5"
)

func GetUserById(w http.ResponseWriter, r *http.Request, app *app.App) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	user, err := app.Models.Users.GetByID(intID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.Write([]byte(fmt.Sprintf("User: %v", user)))
}
