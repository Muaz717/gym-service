package dto

import "time"

// MonthlyStat описывает агрегированные данные по месяцам для статистики.
type MonthlyStat struct {
	Month             time.Time `json:"month"`              // Начало месяца
	Income            float64   `json:"income"`             // Доход за месяц
	NewClients        int       `json:"new_clients"`        // Новые клиенты за месяц
	SoldSubscriptions int       `json:"sold_subscriptions"` // Продано абонементов за месяц
}
