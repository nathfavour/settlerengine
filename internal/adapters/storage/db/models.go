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

type WebhookConfig struct {
	ID        string
	Url       string
	Secret    string
	Events    string
	CreatedAt int64
}

type WebhookDelivery struct {
	ID            string
	ConfigID      string
	Payload       string
	Event         string
	Status        string
	Attempts      int32
	NextAttemptAt int64
	CreatedAt     int64
}

type ClientReputation struct {
	ClientAddress  string
	Score          int32
	TotalPayments  string
	LastPaymentAt  int64
}

type PricingPolicy struct {
	ResourcePath    string
	BasePrice       string
	Currency        string
	SurgeMultiplier float64
}

type Escrow struct {
	ID           string
	InvoiceID    string
	Amount       string
	Currency     string
	Status       string
	DeliveryHash string
	CreatedAt    int64
}

type LsatChallenge struct {
	MacaroonID   string
	PreimageHash string
	Preimage     string
	ResourcePath string
	Amount       int64
	CreatedAt    int64
}

type YieldStrategy struct {
	ID            string
	Provider      string
	VaultAddress  string
	Asset         string
	Tvl           string
	Apy           float64
	LastHarvestAt int64
	Status        string
}

type YieldHarvest struct {
	ID          string
	StrategyID  string
	Amount      string
	Asset       string
	TxHash      string
	Status      string
	HarvestedAt int64
}

