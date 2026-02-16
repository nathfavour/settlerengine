package money

import (
	"errors"
	"math/big"
)

// Money represents a value in a specific currency with precision.
type Money struct {
	amount   *big.Int
	currency string
}

func New(amount *big.Int, currency string) Money {
	return Money{
		amount:   new(big.Int).Set(amount),
		currency: currency,
	}
}

func (m Money) Amount() *big.Int {
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
	return New(res, m.currency), nil
}
