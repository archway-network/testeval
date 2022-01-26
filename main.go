package main

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/archway-network/testeval/configs"
	"github.com/archway-network/testeval/report"
	"github.com/archway-network/testeval/tasks"
	"github.com/archway-network/testeval/winners"

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

	var totalWinnersList winners.WinnersList

	fmt.Printf("\nFinding the active validators winners...\n")
	validatorsWinnersList, err := tasks.GetActiveValidatorsWinners(conn)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	fmt.Printf("\nDone\n")

	fmt.Printf("\nFinding the jailed-unjailed validators winners...\n")
	unjailWinnersList, err := tasks.GetUnjailedValidatorsWinners(conn)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	fmt.Printf("\nDone\n")

	fmt.Printf("\nFinding the governance winners...\n")
	govWinnersList, err := tasks.GetGovAllProposalsWinners(conn)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	fmt.Printf("\nDone\n")

	fmt.Printf("\nFinding the staking winners...\n")
	stakingWinnersList, err := tasks.GetStakingWinners(conn)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	fmt.Printf("\nDone\n")

	fmt.Printf("\nMerging all the winners...\n")
	totalWinnersList.MergeWithAggregateRewards(stakingWinnersList)
	totalWinnersList.MergeWithAggregateRewards(unjailWinnersList)
	totalWinnersList.MergeWithAggregateRewards(govWinnersList)
	totalWinnersList.MergeWithAggregateRewards(validatorsWinnersList)
	fmt.Printf("\nDone\n")

	err = report.StoreWinnersCSV(totalWinnersList)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	allWinners := []report.WinnersListReport{
		{Title: "Active Validator", WinnersList: &validatorsWinnersList, ValidatorOnly: true},
		{Title: "Jailed Unjailed", WinnersList: &unjailWinnersList, ValidatorOnly: true},
		{Title: "Governance", WinnersList: &govWinnersList, ValidatorOnly: false},
		{Title: "Staking", WinnersList: &stakingWinnersList, ValidatorOnly: false},
	}
	err = report.GenerateHTML(totalWinnersList, allWinners)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	fmt.Printf("\n\nTotal winners: %d", totalWinnersList.Length())
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
