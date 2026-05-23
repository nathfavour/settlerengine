package db

type Invoice struct {
	ID        string
	Amount    string
	Currency  string
	Status    string
	CreatedAt int64
	ExpiresAt int64
}

type VerifiedPayment struct {
	Signature  string
	Signer     string
	Amount     string
	Asset      string
	Nonce      string
	VerifiedAt int64
}
