package main

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/archway-network/testnet-evaluator/tasks"

	// sdk "github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// tendermint "github.com/cosmos/cosmos-sdk/x/bank/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {

	conn, err := Connect()
	if err != nil {
		log.Fatalf("Did not connect: %s", err)
	}
	defer conn.Close()

	/*-------------*/

	SetBech32Prefixes()

	/*-------------*/

	// winnersList, err := tasks.GetStakingWinners(conn)
	// winnersList, err := tasks.GetUnjailedValidatorsWinners(conn)
	winnersList, err := tasks.GetActiveValidatorsWinners(conn)
	// winnersList, err := tasks.GetGovAllProposalsWinners(conn)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	for i := 0; i < winnersList.Length(); i++ {
		winner := winnersList.GetItem(i)
		fmt.Printf("\nWinner: %s ==> Reward: %d  on: %s",
			winner.Address,
			winner.Rewards,
			winner.Timestamp)
		// fmt.Printf("\n%s\n\n", winner.TxResponse.TxHash)

		// accAddr, err := sdk.AccAddressFromBech32(winner.Address)
		// if err != nil {
		// 	panic(err)
		// }
		// valAddr := sdk.ValAddress(accAddr).String()
		// fmt.Printf("\t%s", valAddr)
		// fmt.Printf("\n\n\thttps://www.mintscan.io/cosmos/txs/%s\n\n", winnersList[i].TxResponse.TxHash)
	}
	fmt.Printf("\n\nLength: %d", winnersList.Length())
	fmt.Printf("\n\n\t\t-------------------------\n\n")
}

func Connect() (*grpc.ClientConn, error) {

	if configs.Configs.GRPC.TLS {
		creds := credentials.NewTLS(&tls.Config{})
		// conn, err = grpc.Dial("grpc.constantine-1.archway.tech:443", grpc.WithTransportCredentials(creds))
		return grpc.Dial(configs.Configs.GRPC.Server, grpc.WithTransportCredentials(creds))
	}
	return grpc.Dial(configs.Configs.GRPC.Server, grpc.WithInsecure())
}

func SetBech32Prefixes() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(configs.Configs.Bech32Prefix.Account.Address, configs.Configs.Bech32Prefix.Account.PubKey)
	config.SetBech32PrefixForValidator(configs.Configs.Bech32Prefix.Validator.Address, configs.Configs.Bech32Prefix.Validator.PubKey)
	config.SetBech32PrefixForConsensusNode(configs.Configs.Bech32Prefix.Consensus.Address, configs.Configs.Bech32Prefix.Consensus.PubKey)
	config.Seal()
}
