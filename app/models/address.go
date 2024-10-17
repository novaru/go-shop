package models

import "time"

type Address struct {
	ID         string `gorm:"size:36;not null;uniqueIndex;primary_key"`
	User       User
	UserID     string `gorm:"size:36;index"`
	Name       string `gorm:"size:100"`
	IsPrimary  bool
	CityID     string `gorm:"size:36"`
	ProvinceID string `gorm:"size:36"`
	Address1   string `gorm:"size:255"`
	Address2   string `gorm:"size:255"`
	Email      string `gorm:"size:100"`
	Phone      string `gorm:"size:100"`
	PostCode   string `gorm:"size:10"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
