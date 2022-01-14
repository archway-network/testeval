package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/archway-network/testnet-evaluator/tasks"

	// sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gogo/protobuf/proto"

	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

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

	winnersList, err := tasks.GetGovProposalWinners(conn, 60)
	if err != nil {
		log.Fatalf("Error in GetGovProposalWinners: %s", err)
	}
	for i := range winnersList {
		fmt.Printf("\nWinner: %s ==> Reward: %d", winnersList[i].Address, winnersList[i].Rewards)
	}

	fmt.Println()

	return
	/*-------------*/

	{
		c := tx.NewServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		// response, err := c.GetTx(ctx, &tx.GetTxRequest{Hash: "xxxx"})
		response, err := c.GetTxsEvent(
			ctx,
			&tx.GetTxsEventRequest{
				Events: []string{
					// "message.module='governance'",
					// "message.action='/cosmos.gov.v1beta1.MsgVote'",
					"proposal_vote.proposal_id='60'",
					// gov.EventTypeProposalVote+"."+  proposal_id='60'",
				},
				OrderBy: tx.OrderBy_ORDER_BY_ASC,
				Pagination: &query.PageRequest{
					Offset:  0,
					Limit:   10,
					Reverse: false,
				},
			})

		// GetTx(ctx, &tx.GetTxRequest{Hash: "xxxx"})
		if err != nil {
			log.Fatalf("Error in API call: %s", err)
		}
		fmt.Printf("\n---------------------\n\n")

		for i := range response.TxResponses {
			fmt.Printf("\n\nTx (%d): %v", response.TxResponses[i].Height, response.TxResponses[i].TxHash)
			fmt.Printf("\nTimestamp: %#v", response.TxResponses[i].Timestamp)

			voteMsg := gov.MsgVote{}
			err := proto.Unmarshal(response.Txs[i].Body.Messages[0].Value, &voteMsg)
			if err != nil {
				log.Printf("Error unmarshaling: %s", err.Error())
			}
			fmt.Printf("\n\tVoter: %#v", voteMsg.Voter)
			// for j := range response.Txs[i].AuthInfo.SignerInfos {

			// 	fmt.Printf("\n\tSigner: %x", response.Txs[i].AuthInfo.SignerInfos[j].PublicKey.Value)

			// 	// fmt.Printf("\tOption: %v", response.Votes[i].Options[0].Option)
			// }
		}

		fmt.Println()

		// for i := range response.Votes {

		// 	fmt.Printf("\n\tVoter: %s", response.Votes[i].Voter)
		// 	fmt.Printf("\tOption: %v", response.Votes[i].Options[0].Option)
		// }

		// fmt.Printf("\n\n%v\n", response.Votes[0])
		// // fmt.Printf("\n\n%v\n", response.Votes[0].String())
		// fmt.Printf("\n\n%v\n", response.String())

		// // fmt.Printf("\n\n\tResponse from server: %#v\n\n", response)
		// fmt.Printf("\n\n---------------------\n")

		return
	}

	/*-------------*/

	{
		c := gov.NewQueryClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		// response, err := c.Account(ctx, &auth.QueryAccountRequest{Address: "archway1d7krrujhwlkjd5mmv5g6hnqpzpa0dt2x8hcnys"})
		response, err := c.Votes(ctx, &gov.QueryVotesRequest{ProposalId: 60, Pagination: &query.PageRequest{Limit: 10, Reverse: true}})
		if err != nil {
			log.Fatalf("Error in API call: %s", err)
		}
		fmt.Printf("\n---------------------\n\n")

		for i := range response.Votes {

			fmt.Printf("\n\tVoter: %s", response.Votes[i].Voter)
			fmt.Printf("\tOption: %v", response.Votes[i].Options[0].Option)
		}

		fmt.Printf("\n\n%v\n", response.Votes[0])
		// fmt.Printf("\n\n%v\n", response.Votes[0].String())
		fmt.Printf("\n\n%v\n", response.String())

		// fmt.Printf("\n\n\tResponse from server: %#v\n\n", response)
		fmt.Printf("\n\n---------------------\n")

		return
	}

	/*-------------*/

	{
		c := auth.NewQueryClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		// response, err := c.Account(ctx, &auth.QueryAccountRequest{Address: "archway1d7krrujhwlkjd5mmv5g6hnqpzpa0dt2x8hcnys"})
		response, err := c.Account(ctx, &auth.QueryAccountRequest{Address: "archway1ps3v673l3uhg563zddedsm6zju335tq7tsn5na"})
		if err != nil {
			log.Fatalf("Error in API call: %s", err)
		}
		fmt.Printf("\n\n\tResponse from server: %q\n\n", response.Account.TypeUrl)

		return
	}

	/*-------------*/

	c := bank.NewQueryClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	response, err := c.TotalSupply(ctx, &bank.QueryTotalSupplyRequest{})

	if err != nil {
		log.Fatalf("Error in API call: %s", err)
	}
	fmt.Printf("\n\n\tResponse from server: %q\n\n", response.Supply.String())

	/*-------------*/
}

func Connect() (*grpc.ClientConn, error) {

	if configs.Configs.UseTLS {
		creds := credentials.NewTLS(&tls.Config{})
		// conn, err = grpc.Dial("grpc.constantine-1.archway.tech:443", grpc.WithTransportCredentials(creds))
		return grpc.Dial(configs.Configs.GrpcServer, grpc.WithTransportCredentials(creds))
	}
	return grpc.Dial(configs.Configs.GrpcServer, grpc.WithInsecure())
}
