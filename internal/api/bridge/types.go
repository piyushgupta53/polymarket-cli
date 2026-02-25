package bridge

// DepositAddresses holds deposit addresses for supported chains.
type DepositAddresses struct {
	EVMAddress     string `json:"evmAddress"`
	SolanaAddress  string `json:"solanaAddress"`
	BitcoinAddress string `json:"bitcoinAddress"`
}

// SupportedAsset represents a supported chain/token for bridging.
type SupportedAsset struct {
	Chain   string `json:"chain"`
	Token   string `json:"token"`
	Address string `json:"address"`
}

// DepositStatus represents the status of a deposit transaction.
type DepositStatus struct {
	TxHash string `json:"txHash"`
	Status string `json:"status"`
	Amount string `json:"amount,omitempty"`
	Chain  string `json:"chain,omitempty"`
}
