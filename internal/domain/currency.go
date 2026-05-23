package domain

import (
	"errors"
	"math/big"
)

// Money represents an immutable value in a specific currency with big integer precision.
type Money struct {
	amount   *big.Int
	currency string
}

func NewMoney(amount *big.Int, currency string) Money {
	if amount == nil {
		amount = big.NewInt(0)
	}
	return Money{
		amount:   new(big.Int).Set(amount),
		currency: currency,
	}
}

func (m Money) Amount() *big.Int {
	if m.amount == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Set(m.amount)
}

func (m Money) Currency() string {
	return m.currency
}

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.New("currency mismatch")
	}
	res := new(big.Int).Add(m.amount, other.amount)
	return NewMoney(res, m.currency), nil
}
