package auth

import (
	"github.com/gorilla/sessions"
	"github.com/novaru/go-shop/app/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"os"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
var SessionUser = "user-session"

func GetSessionUser(r *http.Request) (*sessions.Session, error) {
	return store.Get(r, SessionUser)
}

func IsLoggedIn(r *http.Request) bool {
	session, _ := store.Get(r, SessionUser)
	return session.Values["id"] != nil
}

func ComparePassword(password string, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

func MakePassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return string(hashedPassword), err
}

func CurrentUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) *models.User {
	if !IsLoggedIn(r) {
		return nil
	}

	session, _ := store.Get(r, SessionUser)

	userModel := models.User{}
	user, err := userModel.FindByID(db, session.Values["id"].(string))
	if err != nil {
		session.Values["id"] = nil
		session.Save(r, w)
		return nil
	}

	return user
}
