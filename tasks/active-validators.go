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
		winnersList.Append(
			winners.Winner{
				Address: activeValidators[i].GetAccountAddress(),
				Rewards: configs.Configs.Tasks.ValidatorJoin.Reward,
			})
	}

	bar.Finish()
	return winnersList, nil
}
