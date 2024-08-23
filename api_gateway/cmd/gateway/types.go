package main

type AuthorizePayload struct {
	CustomerWalletUserId string `json:"customer_wallet_user_id"`
	MerchantWalletUserId string `json:"merchant_wallet_user_id"`
	Cents                int64  `json:"cents"`
	Currency             string `json:"currency"`
}

type AuthorizeResponse struct {
	Pid string `json:"pid"`
}

type CapturePayload struct {
	Pid string `json:"pid"`
}
