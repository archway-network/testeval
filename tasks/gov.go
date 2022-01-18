package tasks

import (
	"fmt"
	"log"

	"github.com/archway-network/testnet-evaluator/configs"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gogo/protobuf/proto"

	"github.com/archway-network/testnet-evaluator/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	"google.golang.org/grpc"
)

func retrieveProposalWinnersFromResponse(response *tx.GetTxsEventResponse) (WinnersList, error) {
	var winnersList WinnersList
	for i := range response.TxResponses {
		// fmt.Printf("\n\nTx (%d): %v", response.TxResponses[i].Height, response.TxResponses[i].TxHash)
		// fmt.Printf("\nTimestamp: %#v", response.TxResponses[i].Timestamp)

		voteMsg := gov.MsgVote{}
		err := proto.Unmarshal(response.Txs[i].Body.Messages[0].Value, &voteMsg)
		if err != nil {
			log.Printf("Error unmarshaling: %s", err.Error())
			continue
		}

		winnersList.Append(
			types.Winner{
				Address:    voteMsg.Voter,
				Rewards:    configs.Configs.Tasks.Gov.Reward,
				Timestamp:  response.TxResponses[i].Timestamp,
				TxResponse: response.TxResponses[i],
			})
	}

	return winnersList, nil

}

func GetGovProposalWinners(conn *grpc.ClientConn, proposalId uint64) (WinnersList, error) {

	return GetWinnersByTxEvents(conn, []string{
		"message.module='governance'",
		"message.action='/cosmos.gov.v1beta1.MsgVote'", //TODO: Maybe we need to find the proper constant instead
		fmt.Sprintf("proposal_vote.proposal_id='%d'", proposalId),
	},
		configs.Configs.Tasks.Gov.MaxWinners,
		retrieveProposalWinnersFromResponse)

}

func GetGovAllProposalsWinners(conn *grpc.ClientConn) (WinnersList, error) {

	var winnersList WinnersList

	for i := range configs.Configs.Tasks.Gov.Proposals {

		proposalId := uint64(configs.Configs.Tasks.Gov.Proposals[i])
		fmt.Printf("\nProcessing proposal # %d\n", proposalId)
		proposalWinnerList, err := GetGovProposalWinners(conn, proposalId)
		if err != nil {
			return winnersList, err
		}

		winnersList.Merge(proposalWinnerList)
	}

	return winnersList, nil
}
