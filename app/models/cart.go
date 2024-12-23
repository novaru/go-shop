package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Cart struct {
	ID              string `gorm:"size:36;not null;uniqueIndex;primary_key"`
	CartItems       []CartItem
	BaseTotalPrice  int
	TaxAmount       decimal.Decimal `gorm:"type:decimal(16,2)"`
	TaxPercent      decimal.Decimal `gorm:"type:decimal(10,2)"`
	DiscountAmount  decimal.Decimal `gorm:"type:decimal(16,2)"`
	DiscountPercent decimal.Decimal `gorm:"type:decimal(10,2)"`
	GrandTotal      int
	TotalWeight     int `gorm:"-"`
}

func (c *Cart) GetCart(db *gorm.DB, cartID string) (*Cart, error) {
	var err error
	var cart Cart

	err = db.Debug().
		Preload("CartItems").
		Preload("CartItems.Product").
		Model(Cart{}).
		Where("id = ?", cartID).
		First(&cart).Error
	if err != nil {
		return nil, err
	}

	return &cart, nil
}

func (c *Cart) CreateCart(db *gorm.DB, cartID string) (*Cart, error) {
	cart := &Cart{
		ID:              cartID,
		BaseTotalPrice:  0,
		TaxAmount:       decimal.NewFromInt(0),
		TaxPercent:      decimal.NewFromInt(11),
		DiscountAmount:  decimal.NewFromInt(0),
		DiscountPercent: decimal.NewFromInt(0),
		GrandTotal:      0,
	}

	err := db.Debug().Create(&cart).Error
	if err != nil {
		return nil, err
	}

	return cart, nil
}

func (c *Cart) CalculateCart(db *gorm.DB, cartID string) (*Cart, error) {
	cartBaseTotalPrice := 0
	cartTaxAmount := 0.0
	cartDiscountAmount := 0.0
	cartGrandTotal := 0
	for _, item := range c.CartItems {
		itemBaseTotal := item.BaseTotal
		itemTaxAmount, _ := item.TaxAmount.Float64()
		itemSubTotalTaxAmount := itemTaxAmount * float64(item.Qty)
		itemDiscountAmount, _ := item.DiscountAmount.Float64()
		itemSubTotalDiscountAmount := itemDiscountAmount * float64(item.Qty)
		itemSubTotal := item.SubTotal
		cartBaseTotalPrice += itemBaseTotal
		cartTaxAmount += itemSubTotalTaxAmount
		cartDiscountAmount += itemSubTotalDiscountAmount
		cartGrandTotal += itemSubTotal
	}
	var updateCart, cart Cart
	updateCart.BaseTotalPrice = cartBaseTotalPrice
	updateCart.TaxAmount = decimal.NewFromFloat(cartTaxAmount)
	updateCart.DiscountAmount = decimal.NewFromFloat(cartDiscountAmount)
	updateCart.GrandTotal = cartGrandTotal
	err := db.Debug().First(&cart, "id = ?", c.ID).Updates(updateCart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (c *Cart) GetItems(db *gorm.DB, cartID string) ([]CartItem, error) {
	var items []CartItem

	err := db.Debug().Preload("Product").
		Model(&CartItem{}).
		Where("cart_id = ?", cartID).
		Order("created_at desc").
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

func (c *Cart) AddItem(db *gorm.DB, item CartItem) (*CartItem, error) {
	var existItem, updateItem CartItem
	var product Product

	// Find the product first
	err := db.Model(Product{}).Where("id = ?", item.ProductID).First(&product).Error
	if err != nil {
		return nil, err
	}

	basePrice := product.Price
	taxAmount := GetTaxAmount(basePrice)
	discountAmount := 0.0

	// Check if item exists in cart
	err = db.Model(CartItem{}).
		Where("cart_id = ?", c.ID).
		Where("product_id = ?", product.ID).
		First(&existItem).Error

	if err != nil {
		// Generate a new UUID for the cart item
		item.ID = uuid.New().String() // Add this line to generate a unique ID
		item.CartID = c.ID
		item.BasePrice = product.Price
		item.BaseTotal = basePrice * item.Qty
		item.TaxPercent = decimal.NewFromFloat(GetTaxPercent())
		item.TaxAmount = decimal.NewFromFloat(taxAmount)
		item.DiscountPercent = decimal.NewFromFloat(0)
		item.DiscountAmount = decimal.NewFromFloat(discountAmount)
		subTotal := float64(item.Qty) * (float64(basePrice) + taxAmount - discountAmount)
		item.SubTotal = int(subTotal)

		err = db.Create(&item).Error
		if err != nil {
			return nil, err
		}
		return &item, nil
	}

	// Update existing item
	updateItem.Qty = existItem.Qty + item.Qty
	updateItem.BaseTotal = basePrice * updateItem.Qty
	subTotal := float64(updateItem.Qty) * (float64(basePrice) + taxAmount - discountAmount)
	updateItem.SubTotal = int(subTotal)

	err = db.Model(&existItem).Updates(updateItem).Error
	if err != nil {
		return nil, err
	}

	return &existItem, nil // Return the existing item instead of input item
}

func (c *Cart) UpdateItemQty(db *gorm.DB, itemID string, qty int) (*CartItem, error) {
	// First find the existing item
	var existItem CartItem
	err := db.Debug().Model(CartItem{}).
		Where("id = ?", itemID).
		First(&existItem).Error
	if err != nil {
		return nil, err
	}

	// Get the product
	var product Product
	err = db.Debug().Model(Product{}).
		Where("id = ?", existItem.ProductID).
		First(&product).Error
	if err != nil {
		return nil, err
	}

	// Calculate all values
	basePrice := product.Price
	taxAmount := GetTaxAmount(basePrice)
	discountAmount := 0.0

	// Calculate new totals
	newBaseTotal := basePrice * qty
	newSubTotal := float64(qty) * (float64(basePrice) + taxAmount - discountAmount)

	// Update using map to ensure all fields are properly set
	updates := map[string]interface{}{
		"qty":        qty,
		"base_total": newBaseTotal,
		"sub_total":  int(newSubTotal),
	}

	// Perform the update
	err = db.Debug().Model(&existItem).Updates(updates).Error
	if err != nil {
		return nil, err
	}

	// Fetch the updated item to return
	var updatedItem CartItem
	err = db.Debug().Model(CartItem{}).
		Where("id = ?", itemID).
		First(&updatedItem).Error
	if err != nil {
		return nil, err
	}

	return &updatedItem, nil
}

func (c *Cart) RemoveItemByID(db *gorm.DB, itemID string) error {
	var err error
	var item CartItem

	err = db.Debug().Model(&CartItem{}).Where("id = ?", itemID).First(&item).Error
	if err != nil {
		return err
	}

	err = db.Debug().Delete(&item).Error
	if err != nil {
		return err
	}

	return nil
}

func (c *Cart) ClearCart(db *gorm.DB, cartID string) error {
	err := db.Debug().Where("cart_id = ?", cartID).Delete(&CartItem{}).Error
	if err != nil {
		return err
	}

	err = db.Debug().Where("id = ?", cartID).Delete(&Cart{}).Error
	if err != nil {
		return err
	}

	return nil
}
