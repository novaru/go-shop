package models

const (
	TaxPercent = 11
)

func GetTaxPercent() float64 {
	return float64(TaxPercent) / 100.0
}

func GetTaxAmount(price int) float64 {
	return GetTaxPercent() * float64(price)
}
