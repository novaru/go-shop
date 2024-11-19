package middlewares

import (
	"github.com/novaru/go-shop/app/core/session/auth"
	"gorm.io/gorm"
	"net/http"
	"slices"
)

func RoleMiddleware(next http.HandlerFunc, db *gorm.DB, roles ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is authenticated
		user := auth.CurrentUser(db, w, r)
		if !slices.Contains(roles, user.Role.Name) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}
