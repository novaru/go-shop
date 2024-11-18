package controllers

import (
	"github.com/novaru/go-shop/app/core/session/auth"
	"github.com/unrolled/render"
	"net/http"
)

func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".tmpl", ".html"},
	})

	user := auth.CurrentUser(server.DB, w, r)

	_ = render.HTML(w, http.StatusOK, "home", map[string]interface{}{
		"user": user,
	})
}
