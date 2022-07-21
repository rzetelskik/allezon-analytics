package api

type Product struct {
	ProductID  uint64 `json:"product_id"`
	BrandID    string `json:"brand_id"`
	CategoryID string `json:"category_id"`
	Price      int32  `json:"price"`
}
