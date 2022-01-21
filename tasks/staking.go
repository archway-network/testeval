package tasks

import (

	// sdk "github.com/cosmos/cosmos-sdk/types"
	// slashing "github.com/cosmos/cosmos-sdk/x/slashing/types"

	"fmt"
	"log"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/archway-network/testnet-evaluator/progressbar"
	"github.com/archway-network/testnet-evaluator/winners"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
)

func GetStakingWinners(conn *grpc.ClientConn) (winners.WinnersList, error) {

	var winnersList winners.WinnersList

	offset := uint64(0)

	for {

		fmt.Printf("\nRetrieving delegators list [offset: %d]...\n", offset)
		delegatorsList, lastOffset, err := getListOfDelegatorsTwice(conn, configs.Configs.Tasks.Staking.MaxWinners, offset)
		offset = lastOffset
		if err != nil {
			return winners.WinnersList{}, err
		}

		fmt.Printf("\nProcessing the delegators list...\n")

		var bar progressbar.Bar
		bar.NewOption(0, int64(delegatorsList.Length()))
		bar.Play(0)

		// Let's check which delegators did at least one Redelegate or undelegate
		for i := 0; i < delegatorsList.Length(); i++ {
			delegator := delegatorsList.GetItem(i)
			bar.Play(int64(i))

			redelegated, err := hasRedelegated(conn, delegator.Address)
			if err != nil {
				return winners.WinnersList{}, err
			}

			undelegated, err := hasUndelegated(conn, delegator.Address)
			if err != nil {
				return winners.WinnersList{}, err
			}

			claimedStakingRewards, err := hasClaimedStakingRewards(conn, delegator.Address)
			if err != nil {
				return winners.WinnersList{}, err
			}

			if (redelegated || undelegated) && claimedStakingRewards {
				winnersList.Append(winners.Winner{
					Address: delegator.Address,
					Rewards: configs.Configs.Tasks.Staking.Reward,
				})
			}
		}
		bar.Finish()

		if winnersList.Length() >= configs.Configs.Tasks.Staking.MaxWinners {
			winnersList = winnersList.Trim(configs.Configs.Tasks.Staking.MaxWinners)
			break
		}
	}
	return winnersList, nil
}

// This function retrieves the list of delegators who had delegated at least two times
// to at least two distinct validators
func getListOfDelegatorsTwice(conn *grpc.ClientConn, maxDelegators int, offset uint64) (winners.WinnersList, uint64, error) {

	var winnersList winners.WinnersList
	var delegatorsList map[string]string = make(map[string]string) // map[ delegatorAddress ] validatorAdress

	var bar progressbar.Bar
	bar.NewOption(0, int64(maxDelegators))
	bar.Play(0)

	for {
		response, err := GetTxEvents(conn,
			[]string{
				"message.module='staking'",
				"message.action='/cosmos.staking.v1beta1.MsgDelegate'",
			}, uint64(100), offset)

		if err != nil {
			return winners.WinnersList{}, offset, err
		}
		offset += uint64(len(response.TxResponses))

		for i := range response.TxResponses {

			delegationMsg := staking.MsgDelegate{}
			err := proto.Unmarshal(response.Txs[i].Body.Messages[0].Value, &delegationMsg)
			if err != nil {
				log.Printf("Error unmarshaling: %s", err.Error())
				continue
			}

			// Check if we have seen this delegator before (since min delegation must be 2)
			if val, ok := delegatorsList[delegationMsg.DelegatorAddress]; ok {

				// The two delegations must be with two distinct validators
				if val != delegationMsg.ValidatorAddress {

					// This list adds items only one time
					winnersList.Append(winners.Winner{
						Address:    delegationMsg.DelegatorAddress,
						Timestamp:  response.TxResponses[i].Timestamp,
						TxResponse: response.TxResponses[i],
					})
				}
			}
			bar.Play(int64(winnersList.Length()))

			delegatorsList[delegationMsg.DelegatorAddress] = delegationMsg.ValidatorAddress
		}

		if winnersList.Length() >= maxDelegators {
			winnersList = winnersList.Trim(maxDelegators)
			break
		}
	}

	bar.Finish()
	return winnersList, offset, nil
}

func hasRedelegated(conn *grpc.ClientConn, delegatorAddress string) (bool, error) {
	response, err := GetTxEvents(conn,
		[]string{
			"message.module='staking'",
			"message.action='/cosmos.staking.v1beta1.MsgBeginRedelegate'",
			fmt.Sprintf("message.sender='%s'", delegatorAddress),
		}, 100, 0)
	if err != nil {
		return false, err
	}

	if response == nil ||
		response.TxResponses == nil ||
		len(response.TxResponses) == 0 {
		return false, nil
	}
	return true, nil
}

func hasUndelegated(conn *grpc.ClientConn, delegatorAddress string) (bool, error) {
	response, err := GetTxEvents(conn,
		[]string{
			"message.module='staking'",
			"message.action='/cosmos.staking.v1beta1.MsgUndelegate'",
			fmt.Sprintf("message.sender='%s'", delegatorAddress),
		}, 100, 0)
	if err != nil {
		return false, err
	}

	if response == nil ||
		response.TxResponses == nil ||
		len(response.TxResponses) == 0 {
		return false, nil
	}
	return true, nil
}

func hasClaimedStakingRewards(conn *grpc.ClientConn, delegatorAddress string) (bool, error) {
	response, err := GetTxEvents(conn,
		[]string{
			// "message.module='distribution'",
			"message.action='/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward'",
			fmt.Sprintf("message.sender='%s'", delegatorAddress),
		}, 100, 0)
	if err != nil {
		return false, err
	}

	if response == nil ||
		response.TxResponses == nil ||
		len(response.TxResponses) == 0 {
		return false, nil
	}
	return true, nil
}
