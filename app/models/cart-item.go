package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type CartItem struct {
	ID              string `gorm:"size:36;not null;uniqueIndex;primary_key"`
	Cart            Cart
	CartID          string `gorm:"size:36;index"`
	Product         Product
	ProductID       string `gorm:"size:36;index"`
	Qty             int
	BasePrice       int
	BaseTotal       int
	TaxAmount       decimal.Decimal `gorm:"type:decimal(16,2)"`
	TaxPercent      decimal.Decimal `gorm:"type:decimal(10,2)"`
	DiscountAmount  decimal.Decimal `gorm:"type:decimal(16,2)"`
	DiscountPercent decimal.Decimal `gorm:"type:decimal(10,2)"`
	SubTotal        int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
