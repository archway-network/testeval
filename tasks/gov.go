package tasks

import (
	"fmt"
	"log"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/archway-network/testnet-evaluator/progressbar"

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

	var bar progressbar.Bar
	bar.NewOption(0, int64(configs.Configs.Tasks.Gov.MaxWinners))

	var totalProposalWinners WinnersList

	// Since a user might vote couple of times, we need to make sure to get a distinct list of winners
	offset := uint64(0)
	for {
		response, err := GetTxEvents(conn, []string{
			"message.module='governance'",
			"message.action='/cosmos.gov.v1beta1.MsgVote'", //TODO: Maybe we need to find the proper constant instead
			fmt.Sprintf("proposal_vote.proposal_id='%d'", proposalId),
		},
			uint64(configs.Configs.Tasks.Gov.MaxWinners),
			offset)

		if err != nil {
			return nil, err
		}

		proposalWinners, err := retrieveProposalWinnersFromResponse(response)
		if err != nil {
			log.Fatalf("Error in retrieveProposalWinnersFromResponse: %s", err)
			// return nil, err
		}

		totalProposalWinners.Merge(proposalWinners)
		totalProposalWinners = totalProposalWinners.Distinct()

		offset += uint64(proposalWinners.Length())

		// progress := (float64(totalProposalWinners.Length()) / float64(configs.Configs.Tasks.Gov.MaxWinners)) * 100

		// fmt.Printf("\r\tProgress: %.2f %%", progress)
		// bar.Play(int64(progress))
		bar.Play(int64(totalProposalWinners.Length()))

		if proposalWinners.Length() == 0 ||
			totalProposalWinners.Length() >= configs.Configs.Tasks.Gov.MaxWinners {
			totalProposalWinners = totalProposalWinners.Trim(configs.Configs.Tasks.Gov.MaxWinners)
			// fmt.Print("\r\tProgress: 100%                     ")
			bar.Finish()
			break
		}
	}
	return totalProposalWinners, nil
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
