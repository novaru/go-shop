package controllers

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/novaru/go-shop/app/core/session/auth"
	"github.com/novaru/go-shop/app/core/session/flash"
	"github.com/novaru/go-shop/app/models"
	"github.com/unrolled/render"
	"log"
	"net/http"
)

func (server *Server) Login(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".tmpl", ".html"},
	})

	_ = render.HTML(w, http.StatusOK, "login", map[string]interface{}{
		"error": flash.GetFlash(w, r, "error"),
	})
}

func (server *Server) DoLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	userModel := models.User{}
	user, err := userModel.FindByEmail(server.DB, email)
	if err != nil {
		flash.SetFlash(w, r, "error", "Email atau password invalid")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if !auth.ComparePassword(password, user.Password) {
		flash.SetFlash(w, r, "error", "Email atau password invalid")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	session, err := auth.GetSessionUser(r)
	if err != nil {
		fmt.Println("Error getting session")
		session, _ = store.New(r, auth.SessionUser)
	}

	session.Values["id"] = user.ID
	err = session.Save(r, w)
	if err != nil {
		flash.SetFlash(w, r, "error", "Email atau password invalid")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		log.Fatal(err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (server *Server) Register(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout:     "layout",
		Extensions: []string{".tmpl", ".html"},
	})

	_ = render.HTML(w, http.StatusOK, "register", map[string]interface{}{
		"error": flash.GetFlash(w, r, "error"),
	})
}

func (server *Server) DoRegister(w http.ResponseWriter, r *http.Request) {
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if firstName == "" || lastName == "" || email == "" || password == "" {
		flash.SetFlash(w, r, "error", "First name, last name, email atau password diperlukan!")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	userModel := models.User{}
	existUser, _ := userModel.FindByEmail(server.DB, email)
	if existUser != nil {
		flash.SetFlash(w, r, "error", "Maaf, email sudah terdaftar")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	hashedPassword, _ := auth.MakePassword(password)
	params := &models.User{
		ID:        uuid.New().String(),
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Password:  hashedPassword,
	}

	user, err := userModel.CreateUser(server.DB, params)
	if err != nil {
		flash.SetFlash(w, r, "error", "Maaf, registrasi gagal")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	// langsung login (assign user_id ke session)
	session, _ := auth.GetSessionUser(r)
	session.Values["id"] = user.ID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (server *Server) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.GetSessionUser(r)
	session.Values["id"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
