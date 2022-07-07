package allezon_analytics

type Product struct {
	ProductID  uint64 `json:"product_id"` // FIXME: it's supposed to be a string
	BrandID    string `json:"brand_id"`
	CategoryID string `json:"category_id"`
	Price      int32  `json:"price"`
}
