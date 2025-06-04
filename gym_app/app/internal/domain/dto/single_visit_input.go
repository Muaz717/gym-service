package dto

type SingleVisitInput struct {
	VisitDate  string  `json:"visit_date"`
	FinalPrice float64 `json:"final_price"`
}
