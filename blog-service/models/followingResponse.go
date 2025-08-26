package models

type FollowingResponse struct {
	User      string   `json:"user"`
	Following []string `json:"following"`
}