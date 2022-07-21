package api

type UserProfileResponse struct {
	Cookie string `json:"cookie"`
	UserProfile
}
