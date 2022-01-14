package tasks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/archway-network/testnet-evaluator/configs"

	// sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gogo/protobuf/proto"

	"github.com/archway-network/testnet-evaluator/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	"google.golang.org/grpc"
)

func GetGovProposalWinners(conn *grpc.ClientConn, proposalId uint64) ([]types.Winner, error) {

	var winnersList []types.Winner

	for retry := 0; retry < configs.Configs.APICallRetry; retry++ {

		c := tx.NewServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		response, err := c.GetTxsEvent(
			ctx,
			&tx.GetTxsEventRequest{
				Events: []string{
					"message.module='governance'",
					"message.action='/cosmos.gov.v1beta1.MsgVote'", //TODO: Maybe we need to find the proper constant instead
					fmt.Sprintf("proposal_vote.proposal_id='%d'", proposalId),
				},
				OrderBy: tx.OrderBy_ORDER_BY_ASC,
				Pagination: &query.PageRequest{
					Offset:  0,
					Limit:   configs.Configs.Tasks.Gov.MaxWinners,
					Reverse: false,
				},
			})

		if err != nil {
			log.Printf("Error in API call: %s\n\tRetrying [ %d ]...", err, retry+1)

			// Ideally we want to retry after getting 502 http error, because sometimes server returns it
			// but we cannot have it as the protobuf Invoke does not return the status code
			continue
		}

		for i := range response.TxResponses {
			// fmt.Printf("\n\nTx (%d): %v", response.TxResponses[i].Height, response.TxResponses[i].TxHash)
			// fmt.Printf("\nTimestamp: %#v", response.TxResponses[i].Timestamp)

			voteMsg := gov.MsgVote{}
			err := proto.Unmarshal(response.Txs[i].Body.Messages[0].Value, &voteMsg)
			if err != nil {
				log.Printf("Error unmarshaling: %s", err.Error())
				continue
			}

			winnersList = append(winnersList,
				types.Winner{
					Address: voteMsg.Voter,
					Rewards: configs.Configs.Tasks.Gov.Reward,
				})
		}

		return winnersList, nil
	}

	return winnersList, fmt.Errorf("something went wrong")
}
