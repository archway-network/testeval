package tasks

import (
	"fmt"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/archway-network/testnet-evaluator/progressbar"
	"github.com/archway-network/testnet-evaluator/validators"
	"github.com/archway-network/testnet-evaluator/winners"
	"google.golang.org/grpc"
)

func GetActiveValidators(conn *grpc.ClientConn) (validators.ValidatorsList, error) {

	fmt.Printf("\nPreparing a list of active validators...\n")
	return validators.GetActiveValidators(conn)
}

func GetActiveValidatorsWinners(conn *grpc.ClientConn) (winners.WinnersList, error) {

	var winnersList winners.WinnersList

	activeValidators, err := GetActiveValidators(conn)
	if err != nil {
		return winners.WinnersList{}, err
	}

	fmt.Printf("\nCalculating rewards...\n")

	var bar progressbar.Bar
	bar.NewOption(0, int64(len(activeValidators)))
	bar.Play(0)

	for i := range activeValidators {

		bar.Play(int64(i))

		newWinner := winners.Winner{
			Address: activeValidators[i].GetAccountAddress(),
			Rewards: configs.Configs.Tasks.ValidatorJoin.Reward,
		}

		if configs.Configs.IdVerification.Required {
			verified, err := newWinner.Verify(conn)
			if err != nil {
				return winners.WinnersList{}, err
			}
			if !verified {
				continue //ignore the unverified winners
			}
		}

		winnersList.Append(newWinner)
		if winnersList.Length() >= configs.Configs.Tasks.ValidatorJoin.MaxWinners {
			break // Max winners reached
		}
	}

	bar.Finish()
	return winnersList, nil
}
