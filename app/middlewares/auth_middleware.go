package middlewares

import (
	"github.com/novaru/go-shop/app/core/session/auth"
	"github.com/novaru/go-shop/app/core/session/flash"
	"net/http"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsLoggedIn(r) {
			// TODO: Set a flash message
			flash.SetFlash(w, r, "error", "Anda harus login dahulu")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	}
}
