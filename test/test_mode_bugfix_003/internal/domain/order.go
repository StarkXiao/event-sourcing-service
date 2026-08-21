package domain

type Order struct {
	ID, BuyerID, Currency, Status string
	Amount, Version               int64
}

func (o *Order) Apply(e Event) {
	o.Version = e.Version
	switch e.Type {
	case OrderCreated:
		o.BuyerID = stringVal(e.Payload, "buyer_id")
		o.Currency = stringVal(e.Payload, "currency")
		o.Amount = intVal(e.Payload, "amount")
		o.Status = "pending"
	case OrderPaid:
		o.Status = "paid"
	case OrderCancelled:
		o.Status = "cancelled"
	case OrderExpired:
		o.Status = "expired"
	}
}
func (o Order) Pay() (string, error) {
	if o.Status == "paid" {
		return "", nil
	}
	if o.Status != "pending" {
		return "", ErrInvalid
	}
	return OrderPaid, nil
}
func (o Order) Cancel() (string, error) {
	if o.Status != "pending" {
		return "", ErrInvalid
	}
	return OrderCancelled, nil
}
func stringVal(m map[string]any, k string) string { v, _ := m[k].(string); return v }
func intVal(m map[string]any, k string) int64 {
	switch v := m[k].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	}
	return 0
}
