package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/novaru/go-shop/app/models"
	"github.com/shopspring/decimal"
	"github.com/unrolled/render"
	"gorm.io/gorm"
	"log"
	"math"
	"net/http"
	"os"
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
		Layout:     "layout",
		Extensions: []string{".tmpl", ".html"},
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
		"success":   GetFlash(w, r, "success"),
		"error":     GetFlash(w, r, "error"),
	})
}

func (server *Server) AddItemToCart(w http.ResponseWriter, r *http.Request) {
	productID := r.FormValue("product_id")
	qty, _ := strconv.Atoi(r.FormValue("qty"))

	productModel := models.Product{}
	product, err := productModel.FindByID(server.DB, productID)
	if err != nil {
		http.Redirect(w, r, "/products/"+product.Slug, http.StatusSeeOther)
		return
	}

	if qty > product.Stock {
		SetFlash(w, r, "error", "Stok item tidak mencukupi")
		http.Redirect(w, r, "/products/"+product.Slug, http.StatusSeeOther)
		return
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

	SetFlash(w, r, "success", "Item berhasil ditambahkan")
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

func (server *Server) GetCitiesByProvince(w http.ResponseWriter, r *http.Request) {
	provinceID := r.URL.Query().Get("province_id")

	cities, err := server.GetCitiesByProvinceID(provinceID)
	if err != nil {
		log.Fatal(err)
	}

	res := Result{Code: 200, Data: cities, Message: "Success"}
	result, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func (server *Server) CalculateShipping(w http.ResponseWriter, r *http.Request) {
	origin := os.Getenv("API_ONGKIR_ORIGIN")
	destination := r.FormValue("city_id")
	courier := r.FormValue("courier")

	if destination == "" {
		http.Error(w, "invalid destination", http.StatusInternalServerError)
	}

	cartID := GetShoppingCartID(w, r)
	cart, _ := GetShoppingCart(server.DB, cartID)

	shippingFeeOptions, err := server.CalculateShippingFee(models.ShippingFeeParams{
		Origin:      origin,
		Destination: destination,
		Weight:      cart.TotalWeight,
		Courier:     courier,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	res := Result{Code: 200, Data: shippingFeeOptions, Message: "Success"}
	result, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func (server *Server) ApplyShipping(w http.ResponseWriter, r *http.Request) {
	origin := os.Getenv("API_ONGKIR_ORIGIN")
	destination := r.FormValue("city_id")
	courier := r.FormValue("courier")
	shippingPackage := r.FormValue("shipping_package")

	cartID := GetShoppingCartID(w, r)
	cart, _ := GetShoppingCart(server.DB, cartID)

	if destination == "" {
		http.Error(w, "invalid destination", http.StatusInternalServerError)
		return
	}

	shippingFeeOptions, err := server.CalculateShippingFee(models.ShippingFeeParams{
		Origin:      origin,
		Destination: destination,
		Weight:      cart.TotalWeight,
		Courier:     courier,
	})

	if err != nil {
		http.Error(w, "invalid shipping calculation", http.StatusInternalServerError)
		return
	}

	var selectedShipping models.ShippingFeeOption

	for _, shippingOption := range shippingFeeOptions {
		if shippingOption.Service == shippingPackage {
			selectedShipping = shippingOption
			continue
		}
	}

	type ApplyShippingResponse struct {
		TotalOrder  int             `json:"total_order"`
		ShippingFee int             `json:"shipping_fee"`
		GrandTotal  int             `json:"grand_total"`
		TotalWeight decimal.Decimal `json:"total_weight"`
	}

	shippingFee := selectedShipping.Fee

	cartGrandTotal := cart.GrandTotal
	grandTotal := cartGrandTotal + shippingFee

	applyShippingResponse := ApplyShippingResponse{
		TotalOrder:  cart.GrandTotal,
		ShippingFee: shippingFee,
		GrandTotal:  grandTotal,
		TotalWeight: decimal.NewFromInt(int64(cart.TotalWeight)),
	}

	res := Result{Code: 200, Data: applyShippingResponse, Message: "Success"}
	result, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
