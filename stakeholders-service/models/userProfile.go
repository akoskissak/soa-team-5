package models

type UserProfile struct {
	FirstName      string `bson:"first_name" json:"firstName"`
	LastName       string `bson:"last_name" json:"lastName"`
	ProfilePicture string `bson:"profile_picture" json:"profilePicture"`
	Biography      string `bson:"biography" json:"biography"`
	Motto          string `bson:"motto" json:"motto"`
}
