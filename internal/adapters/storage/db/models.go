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
