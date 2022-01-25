package winners

import (
	"fmt"
	"log"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/archway-network/testnet-evaluator/events"
	"github.com/archway-network/testnet-evaluator/progressbar"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"google.golang.org/grpc"
)

func GetWinnersByTxEvents(conn *grpc.ClientConn, listOfEvents []string, maxWinners int,
	extractorFunc func(response *tx.GetTxsEventResponse) (WinnersList, error)) (WinnersList, error) {

	if maxWinners < 1 {
		return WinnersList{}, fmt.Errorf("`max_winners` must be greater than zero")
	}

	var bar progressbar.Bar
	bar.NewOption(0, int64(maxWinners))
	bar.Play(0)

	var totalWinners WinnersList

	// Since a user might submit a transaction a couple of times,
	// we need to make sure to get a distinct list of winners
	offset := uint64(0)
	for {
		response, err := events.GetTxEvents(conn, listOfEvents, uint64(maxWinners), offset)

		if err != nil {
			return WinnersList{}, err
		}

		thisRoundWinners, err := extractorFunc(response)
		if err != nil {
			// log.Fatalf("Error in extractorFunction: %s", err)
			return WinnersList{}, err
		}

		// <!-- Verification process
		if configs.Configs.IdVerification.Required {

			fmt.Printf("\nVerifying the identity of the winners...\n")
			err = thisRoundWinners.VerifyAll(conn)
			if err != nil {
				log.Fatalf("Error: %s", err)
			}
			fmt.Printf("\nDone\n")

			// Only new winners add to the list
			totalWinners.Merge(thisRoundWinners.GetVerifiedOnly())

		} else {

			// Only new winners add to the list
			totalWinners.Merge(thisRoundWinners)
		}

		// -->

		offset += uint64(thisRoundWinners.Length())
		// thisRoundWinners.Print()

		bar.Play(int64(totalWinners.Length()))

		if thisRoundWinners.Length() == 0 ||
			totalWinners.Length() >= maxWinners {
			totalWinners = totalWinners.Trim(maxWinners)
			bar.Finish()
			break
		}
	}
	return totalWinners, nil
}
