package domain

type Account struct {
	ID, Currency                        string
	Balance, Credited, Debited, Version int64
}

func (a *Account) Apply(e Event) {
	a.Version = e.Version
	switch e.Type {
	case AccountOpened:
		a.Currency = stringVal(e.Payload, "currency")
	case MoneyCredited:
		n := intVal(e.Payload, "amount")
		a.Balance += n
		a.Credited += n
	case MoneyDebited:
		n := intVal(e.Payload, "amount")
		a.Balance -= n
		a.Debited += n
	}
}
func (a Account) Debit(n int64) error {
	if n <= 0 {
		return ErrInvalid
	}
	if a.Balance < n {
		return ErrInsufficient
	}
	return nil
}
