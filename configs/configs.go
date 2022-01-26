package configs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Configuration struct {
	GRPC struct {
		Server       string `json:"server"`
		TLS          bool   `json:"tls"`
		APICallRetry int    `json:"api_call_retry"`
		CallTimeout  int    `json:"call_timeout"`
	} `json:"grpc"`

	Tasks struct {
		Gov struct {
			MaxWinners     int      `json:"max_winners"`     // Max number of winners for this tasks
			Proposals      []uint64 `json:"proposals"`       // The list of Proposal Ids to be investigated
			Reward         uint64   `json:"reward"`          // Reward for each winner
			ValidatorsOnly bool     `json:"validators_only"` // If this task is for Validators only
		} `json:"gov"`

		ValidatorJoin struct { // Active validators
			MaxWinners     int    `json:"max_winners"`     // Max number of winners for this tasks
			Reward         uint64 `json:"reward"`          // Reward for each winner
			ValidatorsOnly bool   `json:"validators_only"` // If this task is for Validators only
		} `json:"validator_join"`

		ValidatorRun struct { // The participants who just run the validator (it could be inactive)
			MaxWinners     int    `json:"max_winners"`     // Max number of winners for this tasks
			Reward         uint64 `json:"reward"`          // Reward for each winner
			ValidatorsOnly bool   `json:"validators_only"` // If this task is for Validators only
		} `json:"validator_run"`

		JailUnjail struct {
			MaxWinners     int    `json:"max_winners"`     // Max number of winners for this tasks
			Reward         uint64 `json:"reward"`          // Reward for each winner
			ValidatorsOnly bool   `json:"validators_only"` // If this task is for Validators only
		} `json:"jail_unjail"`

		Staking struct {
			MaxWinners     int    `json:"max_winners"`     // Max number of winners for this tasks
			Reward         uint64 `json:"reward"`          // Reward for each winner
			ValidatorsOnly bool   `json:"validators_only"` // If this task is for Validators only
		} `json:"staking"`
	} `json:"tasks"`

	Bech32Prefix struct {
		Account struct {
			Address string `json:"address"`
			PubKey  string `json:"pubkey"`
		} `json:"account"`

		Validator struct {
			Address string `json:"address"`
			PubKey  string `json:"pubkey"`
		} `json:"validator"`

		Consensus struct {
			Address string `json:"address"`
			PubKey  string `json:"pubkey"`
		} `json:"consensus"`
	} `json:"bech32_prefix"`

	BlockExplorer struct {
		TxHash    string `json:"tx_hash"`
		Account   string `json:"account"`
		Validator string `json:"validator"`
	} `json:"block_explorer"`

	Report struct {
		OutputDir string `json:"output_dir"`
	} `json:"report"`

	IdVerification struct {
		Required   bool `json:"required"`    // If it is required to do an ID verification and filter out the not-verified users
		HTMLReport bool `json:"html_report"` // If the ID verification data should be shown in the HTML report
		InputFile  struct {
			Path   string `json:"path"` // Path to the CSV file containing the verification data
			Fields struct {
				Email string `json:"email"`
				KYCId string `json:"kyc_id"`
			} `json:"fields"`
		} `json:"input_file"`
		VerifierAccount string `json:"verifier_account"` // An account that all ID verification tx is sent to (in its Memo field)
	} `json:"id_verification"`
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
