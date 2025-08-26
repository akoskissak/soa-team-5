package models

type FollowReq struct {
	To string `json:"to"` // koga prati
}

type RecDTO struct {
	Username string `json:"username"`
	Mutuals  int64  `json:"mutuals"`
}