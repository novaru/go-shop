package controllers

import (
	"github.com/gorilla/mux"
	"github.com/novaru/go-shop/app/consts"
	"github.com/novaru/go-shop/app/middlewares"
	"net/http"
)

func (server *Server) InitializeRoutes() {
	server.Router = mux.NewRouter()
	server.Router.HandleFunc("/", server.Home).Methods("GET")

	server.Router.HandleFunc("/login", server.Login).Methods("GET")
	server.Router.HandleFunc("/login", server.DoLogin).Methods("POST")
	server.Router.HandleFunc("/register", server.Register).Methods("GET")
	server.Router.HandleFunc("/register", server.DoRegister).Methods("POST")
	server.Router.HandleFunc("/logout", server.Logout).Methods("GET")

	server.Router.HandleFunc("/admin/dashboard",
		middlewares.AuthMiddleware(
			middlewares.RoleMiddleware(
				server.AdminDashboard,
				server.DB,
				consts.RoleAdmin,
				consts.RoleOperator))).
		Methods("GET")

	server.Router.HandleFunc("/products", server.Products).Methods("GET")
	server.Router.HandleFunc("/products/{slug}", server.GetProductBySlug).Methods("GET")

	//server.Router.HandleFunc("/upload", server.UploadProduct).Methods("GET")
	//server.Router.HandleFunc("/upload", server.DoUploadProduct).Methods("POST")

	server.Router.HandleFunc("/carts", server.GetCart).Methods("GET")
	server.Router.HandleFunc("/carts", server.AddItemToCart).Methods("POST")
	server.Router.HandleFunc("/carts/update", server.UpdateCart).Methods("POST")
	server.Router.HandleFunc("/carts/cities", server.GetCitiesByProvince).Methods("GET")
	server.Router.HandleFunc("/carts/calculate-shipping", server.CalculateShipping).Methods("POST")
	server.Router.HandleFunc("/carts/apply-shipping", server.ApplyShipping).Methods("POST")
	server.Router.HandleFunc("/carts/remove/{id}", server.RemoveItemByID).Methods("GET")

	server.Router.HandleFunc("/orders/checkout", middlewares.AuthMiddleware(server.Checkout)).Methods("POST")
	server.Router.HandleFunc("/orders/{id}", middlewares.AuthMiddleware(server.ShowOrder)).Methods("GET")

	server.Router.HandleFunc("/payments/midtrans", server.Midtrans).Methods("POST")

	staticFileDirectory := http.Dir("./assets/")
	staticFileHandler := http.StripPrefix("/public/", http.FileServer(staticFileDirectory))
	server.Router.PathPrefix("/public/").Handler(staticFileHandler).Methods("GET")
}
