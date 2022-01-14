package configs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/archway-network/testnet-evaluator/types"
)

var Configs types.Configuration

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
