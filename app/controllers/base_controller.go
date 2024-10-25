package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/novaru/go-shop/app/config"
	"github.com/novaru/go-shop/app/models"
	"github.com/novaru/go-shop/database/seeders"
	"github.com/urfave/cli"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type Server struct {
	DB     *gorm.DB
	Router *mux.Router
	Config *config.Env
}

type PageLink struct {
	Page          int32
	Url           string
	IsCurrentPage bool
}

type PaginationLinks struct {
	CurrentPage string
	NextPage    string
	PrevPage    string
	TotalRows   int32
	TotalPages  int32
	Links       []PageLink
}

type PaginationParams struct {
	Path        string
	TotalRows   int32
	PerPage     int32
	CurrentPage int32
}

type Result struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

var (
	store               *sessions.CookieStore
	sessionShoppingCart = "shopping-cart-session"
	sessionFlash        = "flash-session"
	sessionUser         = "user-session"
)

func (server *Server) Initialize(c config.Env) {
	fmt.Printf("Welcome to the App %s\n", c.AppName)

	server.Config = &c
	server.InitializeDB(c)
	server.InitializeRoutes()
	server.InitializeStore([]byte(os.Getenv("SESSION_KEY")))
}

func (server *Server) InitializeDB(c config.Env) {
	var err error
	var (
		dbDriver = c.DBDriver
		dbHost   = c.DBHost
		dbPort   = c.DBPort
		dbName   = c.DBName
		dbUser   = c.DBUser
		dbPass   = c.DBPass
	)

	if dbDriver == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUser, dbPass, dbHost, dbPort, dbName)
		server.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	} else {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
			dbHost, dbUser, dbPass, dbName, dbPort)
		server.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		fmt.Println("Failed to connect to database")
	}
}

func (server *Server) DBMigrate() {
	var err error
	for _, model := range models.RegisterModels() {
		err = server.DB.Debug().AutoMigrate(model.Model)

		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Database migrated successfully")
}

func (server *Server) InitCommands(c config.Env) {
	server.InitializeDB(c)

	cmdApp := cli.NewApp()
	cmdApp.Commands = []cli.Command{
		{
			Name: "db:migrate",
			Action: func(c *cli.Context) error {
				server.DBMigrate()
				return nil
			},
		},
		{
			Name: "db:seed",
			Action: func(c *cli.Context) error {
				db := server.DB
				if err := seeders.DBSeed(db); err != nil {
					log.Fatal(err)
				}
				return nil
			},
		},
	}

	if err := cmdApp.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func (server *Server) Run(addr string) {
	fmt.Printf("Starting server at localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, server.Router))
}

func (server *Server) InitializeStore(secret []byte) {
	store = sessions.NewCookieStore(secret)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}
}

func GetPaginationLinks(c *config.Env, params PaginationParams) (PaginationLinks, error) {
	var links []PageLink

	totalPages := int32(math.Ceil(float64(params.TotalRows) / float64(params.PerPage)))

	for i := 1; int32(i) <= totalPages; i++ {
		links = append(links, PageLink{
			Page:          int32(i),
			Url:           fmt.Sprintf("%s/%s?page=%s", c.AppURL, params.Path, fmt.Sprint(i)),
			IsCurrentPage: int32(i) == params.CurrentPage,
		})
	}

	var nextPage int32
	var prevPage int32

	prevPage = 1
	nextPage = totalPages

	if params.CurrentPage > 2 {
		prevPage = params.CurrentPage - 1
	}

	if params.CurrentPage < totalPages {
		nextPage = params.CurrentPage + 1
	}

	return PaginationLinks{
		CurrentPage: fmt.Sprintf("%s/%s?page=%s", c.AppURL, params.Path, fmt.Sprint(params.CurrentPage)),
		NextPage:    fmt.Sprintf("%s/%s?page=%s", c.AppURL, params.Path, fmt.Sprint(nextPage)),
		PrevPage:    fmt.Sprintf("%s/%s?page=%s", c.AppURL, params.Path, fmt.Sprint(prevPage)),
		TotalRows:   params.TotalRows,
		TotalPages:  totalPages,
		Links:       links,
	}, nil
}

func (server *Server) GetProvinces() ([]models.Province, error) {
	response, err := http.Get(os.Getenv("API_ONGKIR_BASE_URL") + "/province?key=" + os.Getenv("API_ONGKIR_KEY"))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	provinceResponse := models.ProvinceResponse{}
	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}

	jsonErr := json.Unmarshal(body, &provinceResponse)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return provinceResponse.ProvinceData.Results, nil
}

func (server *Server) GetCitiesByProvinceID(provinceID string) ([]models.City, error) {
	response, err := http.Get(os.Getenv("API_ONGKIR_BASE_URL") + "city?key=" + os.Getenv("API_ONGKIR_KEY") + "&province=" + provinceID)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	cityResponse := models.CityResponse{}

	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}

	jsonErr := json.Unmarshal(body, &cityResponse)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return cityResponse.CityData.Results, nil
}

func (server *Server) CalculateShippingFee(shippingParams models.ShippingFeeParams) ([]models.ShippingFeeOption, error) {
	if shippingParams.Origin == "" || shippingParams.Destination == "" || shippingParams.Weight <= 0 || shippingParams.Courier == "" {
		return nil, errors.New("invalid params")
	}

	params := url.Values{}
	params.Add("key", os.Getenv("API_ONGKIR_KEY"))
	params.Add("origin", shippingParams.Origin)
	params.Add("destination", shippingParams.Destination)
	params.Add("weight", strconv.Itoa(shippingParams.Weight))
	params.Add("courier", shippingParams.Courier)

	response, err := http.PostForm(os.Getenv("API_ONGKIR_BASE_URL")+"cost", params)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	ongkirResponse := models.OngkirResponse{}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	jsonErr := json.Unmarshal(body, &ongkirResponse)
	if jsonErr != nil {
		return nil, jsonErr
	}

	var shippingFeeOptions []models.ShippingFeeOption
	for _, result := range ongkirResponse.OngkirData.Results {
		for _, cost := range result.Costs {
			shippingFeeOptions = append(shippingFeeOptions, models.ShippingFeeOption{
				Service: cost.Service,
				Fee:     cost.Cost[0].Value,
			})
		}
	}

	return shippingFeeOptions, nil
}

func SetFlash(w http.ResponseWriter, r *http.Request, name string, value string) {
	session, err := store.Get(r, sessionFlash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.AddFlash(value, name)
	session.Save(r, w)
}

func GetFlash(w http.ResponseWriter, r *http.Request, name string) []string {
	session, err := store.Get(r, sessionFlash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	fm := session.Flashes(name)
	if len(fm) == 0 {
		return nil
	}

	session.Save(r, w)
	var flashes []string
	for _, fl := range fm {
		flashes = append(flashes, fl.(string))
	}

	return flashes
}

func IsLoggedIn(r *http.Request) bool {
	session, _ := store.Get(r, sessionUser)
	if session.Values["id"] == nil {
		return false
	}

	return true
}

func ComparePassword(password string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func MakePassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return string(hashedPassword), err
}

func (server *Server) CurrentUser(w http.ResponseWriter, r *http.Request) *models.User {
	if !IsLoggedIn(r) {
		return nil
	}

	session, _ := store.Get(r, sessionUser)
	userModel := models.User{}
	user, err := userModel.FindByID(server.DB, session.Values["id"].(string))
	if err != nil {
		session.Values["id"] = nil
		session.Save(r, w)
		return nil
	}

	return user
}
