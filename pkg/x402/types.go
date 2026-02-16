package x402

// PaymentDescriptor defines the standard JSON structure for 402 responses.
type PaymentDescriptor struct {
	Scheme  string `json:"scheme"`  // e.g., "x402"
	Price   string `json:"price"`   // Amount in atomic units (uint256 string)
	Asset   string `json:"asset"`   // Token contract address (USDC)
	Network string `json:"network"` // Chain ID or network name
	PayTo   string `json:"payTo"`   // Merchant wallet address
	Nonce   string `json:"nonce"`   // Unique session UUID for the challenge
}

// ChallengeResponse is the body returned with a 402 status code.
type ChallengeResponse struct {
	Status      int                 `json:"status"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Accepts     []PaymentDescriptor `json:"accepts"`
	Resource    string              `json:"resource"`
}
