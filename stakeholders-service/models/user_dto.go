package models

type UserProfileResponse struct {
	Username       string `json:"username"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	ProfilePicture string `json:"profilePicture"`
	Biography      string `json:"biography"`
	Motto          string `json:"motto"`
}
