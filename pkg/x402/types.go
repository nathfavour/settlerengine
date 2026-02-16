package x402

// PaymentDescriptor defines the standard JSON structure for 402 responses.
type PaymentDescriptor struct {
	Amount    string `json:"amount"`    // Amount in atomic units (uint256 string)
	Asset     string `json:"asset"`     // Token contract address (USDC)
	Network   string `json:"network"`   // Chain ID or network name
	Recipient string `json:"recipient"` // Merchant wallet address
	Nonce     string `json:"nonce"`     // Unique session UUID for the challenge
}

// ChallengeResponse is the body returned with a 402 status code.
type ChallengeResponse struct {
	Status      int               `json:"status"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Payment     PaymentDescriptor `json:"payment"`
}
