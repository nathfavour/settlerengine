package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/nathfavour/settlerengine/core/domain/model"
	"github.com/nathfavour/settlerengine/pkg/money"
)

func main() {
	fmt.Println("SettlerEngine: The Payment Engine for the Agentic Era")
	log.Println("Starting server...")

	m := money.New(big.NewInt(100), "USD")
	inv := model.NewInvoice("inv_1", m, 3600)
	log.Printf("Created invoice: %s with amount %s %s", inv.ID, inv.Amount.Amount().String(), inv.Amount.Currency())
}
