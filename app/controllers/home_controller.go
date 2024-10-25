package controllers

import (
	"github.com/unrolled/render"
	"net/http"
)

func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".tmpl", ".html"},
	})

	_ = render.HTML(w, http.StatusOK, "home", map[string]interface{}{
		"title":       "Home",
		"description": "This is my home page",
	})
}
