package types

import (
	"github.com/cosmos/cosmos-sdk/types"
)

type Winner struct {
	Address    string            // Wallet Address
	Rewards    uint64            // Total Reward of a winner
	Timestamp  string            // The time of the task done, if applicable
	TxResponse *types.TxResponse // The associated Tx Response with the task, if applicable
}

type Configuration struct {
	GrpcServer   string `json:"grpc_server"`
	UseTLS       bool   `json:"use_tls"`
	APICallRetry int    `json:"api_call_retry"`
	TxHashURL    string `json:"tx_hash_url"` // The URL to the block explorer in order to find more details of a transaction via its hash

	Tasks struct {
		Gov struct {
			MaxWinners int      `json:"max_winners"` // Max number of winners for this tasks
			Proposals  []uint64 `json:"proposals"`   // The list of Proposal Ids to be investigated
			Reward     uint64   `json:"reward"`      // Reward for each winner

		} `json:"gov"`
	} `json:"tasks"`
}
