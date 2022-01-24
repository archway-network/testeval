package tasks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/archway-network/testnet-evaluator/progressbar"
	"github.com/archway-network/testnet-evaluator/winners"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"google.golang.org/grpc"
)

func GetTxEvents(conn *grpc.ClientConn, events []string, limit uint64, offset uint64) (response *tx.GetTxsEventResponse, err error) {

	for retry := 0; retry < configs.Configs.GRPC.APICallRetry; retry++ {

		c := tx.NewServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		response, err = c.GetTxsEvent(
			ctx,
			&tx.GetTxsEventRequest{
				Events:  events,
				OrderBy: tx.OrderBy_ORDER_BY_ASC,
				Pagination: &query.PageRequest{
					// Key:    nextKey,
					Limit:  limit,
					Offset: offset,
					// Reverse: false,
				},
			})

		if err != nil {
			fmt.Printf("\r[%d", retry+1)
			// fmt.Printf("\tRetrying [ %d ]...", retry+1)
			// fmt.Printf("\tErr: %s", err)

			// Ideally we want to retry after getting 502 http error, because sometimes server returns it
			// but we cannot have it as the protobuf Invoke does not return the status code
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// fmt.Printf("Done")
		return response, nil
	}

	return nil, err
}

func GetWinnersByTxEvents(conn *grpc.ClientConn, events []string, maxWinners int,
	extractorFunc func(response *tx.GetTxsEventResponse) (winners.WinnersList, error)) (winners.WinnersList, error) {

	if maxWinners < 1 {
		return winners.WinnersList{}, fmt.Errorf("`max_winners` must be greater than zero")
	}

	var bar progressbar.Bar
	bar.NewOption(0, int64(maxWinners))
	bar.Play(0)

	var totalWinners winners.WinnersList

	// Since a user might submit a transaction a couple of times,
	// we need to make sure to get a distinct list of winners
	offset := uint64(0)
	for {
		response, err := GetTxEvents(conn, events, uint64(maxWinners), offset)

		if err != nil {
			return winners.WinnersList{}, err
		}

		thisRoundWinners, err := extractorFunc(response)
		if err != nil {
			log.Fatalf("Error in extractorFunction: %s", err)
			// return nil, err
		}

		// Only new winners add to the list
		totalWinners.Merge(thisRoundWinners)

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
