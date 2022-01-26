package tasks

import (
	"fmt"

	"github.com/archway-network/testeval/configs"
	"github.com/archway-network/testeval/progressbar"
	"github.com/archway-network/testeval/validators"
	"github.com/archway-network/testeval/winners"
	"google.golang.org/grpc"
)

func GetActiveValidatorsWinners(conn *grpc.ClientConn) (winners.WinnersList, error) {

	return GetValidatorsWinners(conn, true)
}

func GetAllValidatorsWinners(conn *grpc.ClientConn) (winners.WinnersList, error) {

	return GetValidatorsWinners(conn, false)
}

func GetValidatorsWinners(conn *grpc.ClientConn, onlyActiveValidators bool) (winners.WinnersList, error) {

	var winnersList winners.WinnersList
	var listOfValidators validators.ValidatorsList
	var err error

	if onlyActiveValidators {
		fmt.Printf("\nPreparing a list of Active Validators...\n")
		listOfValidators, err = validators.GetActiveValidators(conn)
	} else {
		fmt.Printf("\nPreparing a list of Inactive Validators...\n")
		listOfValidators, err = validators.GetInactiveValidators(conn)
	}

	if err != nil {
		return winners.WinnersList{}, err
	}

	fmt.Printf("\nCalculating rewards...\n")

	var bar progressbar.Bar
	bar.NewOption(0, int64(len(listOfValidators)))
	bar.Play(0)

	for i := range listOfValidators {

		bar.Play(int64(i))

		newWinner := winners.Winner{
			Address: listOfValidators[i].GetAccountAddress(),
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
