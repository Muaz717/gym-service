package models

// Subscription представляет абонемент
type Subscription struct {
	ID           string  `json:"id,omitempty"`  // Номер абонемента
	Title        string  `json:"title"`         // Название тарифа
	Price        float64 `json:"price"`         // Цена тарифа
	DurationDays int     `json:"duration_days"` // Срок действия в днях
	FreezeDays   int     `json:"freeze_days"`   // Количество допустимых дней заморозки
}
