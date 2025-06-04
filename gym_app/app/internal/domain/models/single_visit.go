package models

import "time"

type SingleVisit struct {
	Id         int       `json:"id"`
	VisitDate  time.Time `json:"visit_date"`
	FinalPrice float64   `json:"final_price"`
}
