package controllers

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/novaru/go-shop/app/models"
	"github.com/unrolled/render"
	"gorm.io/gorm"
	"log"
	"math"
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

	updatedCart, _ := cart.GetCart(db, cartID)

	totalWeight := 0
	productModel := models.Product{}
	for _, cartItem := range updatedCart.CartItems {
		product, _ := productModel.FindByID(db, cartItem.ProductID)

		productWeight, _ := product.Weight.Float64()
		ceilWeight := math.Ceil(productWeight)

		itemWeight := cartItem.Qty * int(ceilWeight)

		totalWeight += itemWeight
	}

	updatedCart.TotalWeight = totalWeight

	return updatedCart, nil
}

func (server *Server) GetCart(w http.ResponseWriter, r *http.Request) {
	render := render.New(render.Options{
		Layout: "layout",
	})

	var cart *models.Cart

	cartID := GetShoppingCartID(w, r)
	cart, _ = GetShoppingCart(server.DB, cartID)
	items, _ := cart.GetItems(server.DB, cartID)

	provinces, err := server.GetProvinces()
	if err != nil {
		log.Fatal(err)
	}

	_ = render.HTML(w, http.StatusOK, "cart", map[string]interface{}{
		"cart":      cart,
		"items":     items,
		"provinces": provinces,
	})
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

func (server *Server) UpdateCart(w http.ResponseWriter, r *http.Request) {
	cartID := GetShoppingCartID(w, r)
	cart, _ := GetShoppingCart(server.DB, cartID)

	for _, item := range cart.CartItems {
		qty, _ := strconv.Atoi(r.FormValue(item.ID))

		_, err := cart.UpdateItemQty(server.DB, item.ID, qty)
		if err != nil {
			http.Redirect(w, r, "/carts", http.StatusSeeOther)
		}
	}

	http.Redirect(w, r, "/carts", http.StatusSeeOther)
}

func (server *Server) RemoveItemByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars["id"] == "" {
		http.Redirect(w, r, "/carts", http.StatusSeeOther)
	}

	cartID := GetShoppingCartID(w, r)
	cart, _ := GetShoppingCart(server.DB, cartID)

	err := cart.RemoveItemByID(server.DB, vars["id"])
	if err != nil {
		http.Redirect(w, r, "/carts", http.StatusSeeOther)
	}

	http.Redirect(w, r, "/carts", http.StatusSeeOther)
}

func ClearCart(db *gorm.DB, cartID string) error {
	var cart models.Cart

	err := cart.ClearCart(db, cartID)
	if err != nil {
		return err
	}

	return nil
}
