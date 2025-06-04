package dto

import "time"

// MonthlyStat описывает агрегированные данные по месяцам для статистики.
type MonthlyStat struct {
	Month              time.Time `json:"month"`
	Income             float64   `json:"income"`
	NewClients         int       `json:"new_clients"`
	SoldSubscriptions  int       `json:"sold_subscriptions"`
	SingleVisitsIncome float64   `json:"single_visits_income"`
	SingleVisitsCount  int       `json:"single_visits_count"`
}
