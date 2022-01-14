package types

type Winner struct {
	Address string // Wallet Address
	Rewards uint64 // Total Reward of a winner
}

type Configuration struct {
	GrpcServer   string `json:"grpc_server"`
	UseTLS       bool   `json:"use_tls"`
	APICallRetry int    `json:"api_call_retry"`

	Tasks struct {
		Gov struct {
			MaxWinners uint64   `json:"max_winners"` // Max number of winners for this tasks
			Proposals  []uint64 `json:"proposals"`   // The list of Proposal Ids to be investigated
			Reward     uint64   `json:"reward"`      // Reward for each winner

		} `json:"gov"`
	} `json:"tasks"`
}
