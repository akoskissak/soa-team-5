package models

type TourStatus string

const (
	Draft TourStatus = "Draft"
	Published TourStatus = "Published"
	Archived TourStatus = "Archived"
)