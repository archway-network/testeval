package tasks

import (
	"fmt"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/archway-network/testnet-evaluator/progressbar"
	"github.com/archway-network/testnet-evaluator/validators"
	"github.com/archway-network/testnet-evaluator/winners"
	"google.golang.org/grpc"
)

func GetAllUnjailedValidators(conn *grpc.ClientConn) (validators.ValidatorsList, error) {
	var jailedAndUnjailedValidators validators.ValidatorsList

	activeValidators, err := GetActiveValidators(conn)
	if err != nil {
		return validators.ValidatorsList{}, err
	}

	fmt.Printf("\nAnalyzing validators signing info...\n")

	var bar progressbar.Bar
	bar.NewOption(0, int64(len(activeValidators)))
	bar.Play(0)

	for i := range activeValidators {
		signingInfo, err := validators.GetValidatorsSigningInfo(conn, activeValidators[i].ConsAddress)
		bar.Play(int64(i))

		if err != nil {
			return jailedAndUnjailedValidators, err
		}

		if signingInfo.JailedUntil.Unix() > 0 {

			jailedAndUnjailedValidators = append(jailedAndUnjailedValidators,
				validators.Validator{
					activeValidators[i].Validator,
					signingInfo,
					activeValidators[i].ConsAddress,
				},
			)
		}
	}

	bar.Finish()
	return jailedAndUnjailedValidators, nil
}

func GetUnjailedValidatorsWinners(conn *grpc.ClientConn) (winners.WinnersList, error) {

	var winnersList winners.WinnersList

	activeValidators, err := GetAllUnjailedValidators(conn)
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
