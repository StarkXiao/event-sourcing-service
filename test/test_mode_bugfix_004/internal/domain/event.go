package domain

import "time"

type Event struct {
	ID, TenantID, AggregateType, AggregateID, Type, IdempotencyKey string
	Version, Position                                              int64
	Payload                                                        map[string]any
	Metadata                                                       map[string]any
	OccurredAt                                                     time.Time
}

const (
	OrderCreated   = "OrderCreated"
	OrderPaid      = "OrderPaid"
	OrderCancelled = "OrderCancelled"
	OrderExpired   = "OrderExpired"
	AccountOpened  = "AccountOpened"
	MoneyCredited  = "MoneyCredited"
	MoneyDebited   = "MoneyDebited"
)
