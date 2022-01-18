package tasks

import (
	"context"
	"log"
	"time"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/archway-network/testnet-evaluator/progressbar"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"google.golang.org/grpc"
)

func GetTxEvents(conn *grpc.ClientConn, events []string, limit uint64, offset uint64) (response *tx.GetTxsEventResponse, err error) {

	for retry := 0; retry < configs.Configs.APICallRetry; retry++ {

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
			// fmt.Printf("\tRetrying [ %d ]...", retry+1)

			// Ideally we want to retry after getting 502 http error, because sometimes server returns it
			// but we cannot have it as the protobuf Invoke does not return the status code
			continue
		}

		// fmt.Printf("Done")
		return response, nil
	}

	return nil, err
}

func GetWinnersByTxEvents(conn *grpc.ClientConn, events []string, maxWinners int,
	extractorFunc func(response *tx.GetTxsEventResponse) (WinnersList, error)) (WinnersList, error) {

	var bar progressbar.Bar
	bar.NewOption(0, int64(maxWinners))
	bar.Play(0)

	var totalWinners WinnersList

	// Since a user might submit a transaction a couple of times,
	// we need to make sure to get a distinct list of winners
	offset := uint64(0)
	for {
		response, err := GetTxEvents(conn, events, uint64(maxWinners), offset)

		if err != nil {
			return nil, err
		}

		thisRoundWinners, err := extractorFunc(response)
		if err != nil {
			log.Fatalf("Error in extractorFunction: %s", err)
			// return nil, err
		}

		totalWinners.Merge(thisRoundWinners)
		totalWinners = totalWinners.Distinct()

		offset += uint64(thisRoundWinners.Length())

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
