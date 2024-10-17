package controllers

import (
    "fmt"
    "github.com/novaru/go-shop/app/models"
    "github.com/unrolled/render"
    "net/http"
    "strconv"
)

func (server *Server) Products(w http.ResponseWriter, r *http.Request) {
    render := render.New(render.Options{
        Layout: "layout",
    })

    q := r.URL.Query()
    page, _ := strconv.Atoi(q.Get("page"))
    if page <= 0 {
        page = 1
    }
    perPage := 9

    productModel := models.Product{}
    products, totalRows, err := productModel.GetProducts(server.DB, perPage, page)
    if err != nil {
        return
    }

    pagination, _ := GetPaginationLinks(server.Config, PaginationParams{
        Path:        "products",
        TotalRows:   int32(totalRows),
        PerPage:     int32(perPage),
        CurrentPage: int32(page),
    })

    fmt.Println("rows: ", totalRows)
    fmt.Println("pagination: ", pagination)

    _ = render.HTML(w, http.StatusOK, "products", map[string]interface{}{
        "products":   products,
        "pagination": pagination,
    })
}
