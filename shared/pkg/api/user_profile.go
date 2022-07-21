package api

type UserProfile struct {
	Views []UserTag `json:"views"`
	Buys  []UserTag `json:"buys"`
}
