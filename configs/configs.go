package configs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

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

		ValidatorJoin struct {
			MaxWinners int    `json:"max_winners"` // Max number of winners for this tasks
			Reward     uint64 `json:"reward"`      // Reward for each winner
		} `json:"validator_join"`

		JailUnjail struct {
			MaxWinners int    `json:"max_winners"` // Max number of winners for this tasks
			Reward     uint64 `json:"reward"`      // Reward for each winner
		} `json:"jail_unjail"`
	} `json:"tasks"`
}

var Configs Configuration

/*-------------------------*/

// This function loads the configuration file into the Configs object
func init() {

	filename := GetRootPath() + "/conf.json"
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bytes, &Configs)
	if err != nil {
		panic(err)
	}

}

// This function retrieves the root path of where the binary is being executed
func GetRootPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return dir
}
