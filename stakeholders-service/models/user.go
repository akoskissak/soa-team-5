package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Role string

const (
	RoleTourist Role = "tourist"
	RoleGuide   Role = "guide"
	RoleAdmin   Role = "admin"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username string            `bson:"username" json:"username"`
	Email    string				`bson:"email" json:"email"`
	Password string            `bson:"password,omitempty" json:"password,omitempty"`
	Role     Role              `bson:"role" json:"role"`
}