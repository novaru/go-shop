package controllers

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/novaru/go-shop/app/models"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func GetShoppingCartID(w http.ResponseWriter, r *http.Request) string {
	session, err := store.Get(r, sessionShoppingCart)
	fmt.Println(session)

	if err != nil {
		fmt.Println("Error getting session")
		session, _ = store.New(r, sessionShoppingCart)
	}

	if session.Values["id"] == nil {
		session.Values["id"] = uuid.New().String()
		err = session.Save(r, w)
		if err != nil {
			fmt.Println("Error getting session")
		}
	}

	cartID, ok := session.Values["id"].(string)
	if !ok {
		// If for some reason the ID is not a string, generate a new one
		cartID = uuid.New().String()
		session.Values["id"] = cartID
		err = session.Save(r, w)
		if err != nil {
			fmt.Println(err)
		}
	}

	return cartID
}

func GetShoppingCart(db *gorm.DB, cartID string) (*models.Cart, error) {
	var cart models.Cart

	existCart, err := cart.GetCart(db, cartID)
	if err != nil {
		existCart, _ = cart.CreateCart(db, cartID)
	}
	_, _ = existCart.CalculateCart(db, cartID)
	return existCart, nil
}

func (server *Server) GetCart(w http.ResponseWriter, r *http.Request) {
	var cart *models.Cart

	cartID := GetShoppingCartID(w, r)
	cart, _ = GetShoppingCart(server.DB, cartID)

	fmt.Println("cart id ===> ", cart.ID)
	fmt.Println("cart items ==>", cart.CartItems)
}

func (server *Server) AddItemToCart(w http.ResponseWriter, r *http.Request) {
	productID := r.FormValue("product_id")
	qty, _ := strconv.Atoi(r.FormValue("qty"))

	productModel := models.Product{}
	product, err := productModel.FindByID(server.DB, productID)
	if err != nil {
		http.Redirect(w, r, "/products/"+product.Slug, http.StatusSeeOther)
	}

	if qty > product.Stock {
		http.Redirect(w, r, "/products/"+product.Slug, http.StatusSeeOther)
	}

	var cart *models.Cart

	cartID := GetShoppingCartID(w, r)
	cart, _ = GetShoppingCart(server.DB, cartID)
	_, err = cart.AddItem(server.DB, models.CartItem{
		ProductID: productID,
		Qty:       qty,
	})
	if err != nil {
		http.Redirect(w, r, "/products/"+product.Slug, http.StatusSeeOther)
	}

	http.Redirect(w, r, "/carts", http.StatusSeeOther)
}
