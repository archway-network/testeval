package events

import (
	"context"
	"fmt"
	"time"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"google.golang.org/grpc"
)

func GetTxEvents(conn *grpc.ClientConn, events []string, limit uint64, offset uint64) (response *tx.GetTxsEventResponse, err error) {

	for retry := 0; retry < configs.Configs.GRPC.APICallRetry; retry++ {

		c := tx.NewServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(configs.Configs.GRPC.CallTimeout))
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
