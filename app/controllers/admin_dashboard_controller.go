package controllers

import (
	"fmt"
	"github.com/novaru/go-shop/app/core/session/auth"
	"github.com/novaru/go-shop/app/utils"
	"github.com/unrolled/render"
	"net/http"
)

func (server *Server) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "admin_layout",
		Extensions: []string{".tmpl", ".html"},
	})

	user := auth.CurrentUser(server.DB, w, r)
	fmt.Println("User =>", utils.PrintJSON(user))

	_ = render.HTML(w, http.StatusOK, "admin_dashboard", map[string]interface{}{
		"user": utils.PrintJSON(user),
	})
}
