package application

import (
	"context"
	"event-sourcing-service/internal/domain"
	"event-sourcing-service/internal/eventstore"
	"math"
)

type Commands struct{ Store eventstore.Store }

func (s Commands) Execute(c context.Context, t, typ, id, cmd, key string, expected int64, p map[string]any) ([]domain.Event, error) {
	if t == "" || id == "" || key == "" || p == nil {
		return nil, domain.ErrInvalid
	}
	events, err := s.Store.Load(c, t, typ, id)
	if err != nil {
		return nil, err
	}
	if typ == "order" {
		o := domain.Order{ID: id}
		for _, e := range events {
			o.Apply(e)
		}
		if cmd == "create" && expected != 0 {
			return nil, domain.ErrConflict
		}
		if cmd == "create" && len(events) > 0 {
			return nil, domain.ErrConflict
		}
		if cmd == "create" {
			if stringValue(p, "buyer_id") == "" || stringValue(p, "currency") == "" || amount(p) <= 0 {
				return nil, domain.ErrInvalid
			}
		}
		if cmd == "pay" && o.Status == "paid" {
			return nil, nil
		}
		et := cmd
		if cmd == "create" {
			et = domain.OrderCreated
		}
		if cmd == "expire" {
			et = domain.OrderExpired
			if o.Status != "pending" {
				return nil, domain.ErrInvalid
			}
		}
		if cmd == "pay" {
			et, err = o.Pay()
		} else if cmd == "cancel" {
			et, err = o.Cancel()
		} else if cmd != "create" && cmd != "expire" {
			return nil, domain.ErrInvalid
		}
		if err != nil {
			return nil, err
		}
		return s.Store.Append(c, eventstore.AppendRequest{TenantID: t, Type: typ, ID: id, IdempotencyKey: key, ExpectedVersion: expected, Events: []domain.Event{{Type: et, Payload: p}}})
	}
	if typ != "account" || (cmd != "open" && cmd != "credit" && cmd != "debit") {
		return nil, domain.ErrInvalid
	}
	account := domain.Account{ID: id}
	for _, event := range events {
		account.Apply(event)
	}
	if cmd == "open" {
		if expected != 0 || len(events) > 0 || stringValue(p, "currency") == "" {
			return nil, domain.ErrInvalid
		}
		cmd = domain.AccountOpened
	}
	if cmd == "credit" {
		if len(events) == 0 || account.Currency == "" || amount(p) <= 0 || (p["currency"] != nil && stringValue(p, "currency") != account.Currency) {
			return nil, domain.ErrInvalid
		}
		cmd = domain.MoneyCredited
	}
	if cmd == "debit" {
		if len(events) == 0 || account.Currency == "" || amount(p) <= 0 || (p["currency"] != nil && stringValue(p, "currency") != account.Currency) {
			return nil, domain.ErrInvalid
		}
		cmd = domain.MoneyDebited
	}
	if cmd == domain.MoneyDebited {
		if err := account.Debit(amount(p)); err != nil {
			return nil, err
		}
	}
	return s.Store.Append(c, eventstore.AppendRequest{TenantID: t, Type: typ, ID: id, IdempotencyKey: key, ExpectedVersion: expected, Events: []domain.Event{{Type: cmd, Payload: p}}})
}

func stringValue(p map[string]any, k string) string { v, _ := p[k].(string); return v }

func amount(p map[string]any) int64 {
	switch v := p["amount"].(type) {
	case float64:
		if v != math.Trunc(v) {
			return 0
		}
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	}
	return 0
}
