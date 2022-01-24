package validators

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/archway-network/testnet-evaluator/progressbar"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc"
)

type Validator struct {
	staking.Validator
	slashing.ValidatorSigningInfo
	ConsAddress string
}
type ValidatorsList []Validator

/*-----------------------*/

func getValidatorsSetByOffset(conn *grpc.ClientConn, offset int, status string) (response *staking.QueryValidatorsResponse, err error) {

	for retry := 0; retry < configs.Configs.GRPC.APICallRetry; retry++ {

		c := staking.NewQueryClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		response, err := c.Validators(ctx,
			&staking.QueryValidatorsRequest{
				Status: status,
				Pagination: &query.PageRequest{
					// Key:    nextKey,
					// Limit:  limit,
					Offset: uint64(offset),
					// Reverse: false,
				},
			})
		if err != nil {
			fmt.Printf("\r[%d", retry+1)
			// fmt.Printf("\r\tRetrying [ %d ]...", retry+1)
			// fmt.Printf("\tErr: %s", err)

			// Ideally we want to retry after getting 502 http error, because sometimes server returns it
			// but we cannot have it as the protobuf Invoke does not return the status code
			time.Sleep(50 * time.Millisecond)
			continue
		}

		return response, nil
	}

	return nil, err
}

// This function provides a list of all validators including the inactive ones
func GetValidatorsList(conn *grpc.ClientConn, status string) (validatorsList ValidatorsList, err error) {

	offset := 0
	var bar progressbar.Bar
	bar.NewOption(0, 1)
	bar.Play(0)

	for {
		response, err := getValidatorsSetByOffset(conn, offset, status)
		if err != nil {
			return nil, err
		}

		if response == nil || len(response.Validators) == 0 {
			bar.Finish()
			break
		}
		offset += len(response.Validators)
		bar.NewOption(0, int64(response.Pagination.Total))
		bar.Play(int64(offset))

		// // There is an strange behavior of the API which retrieves the redundant data for offset 100 and 150
		// if uint64(offset) >= response.Pagination.Total {
		// 	break
		// }

		for i := range response.Validators {
			validatorsList = append(validatorsList,
				Validator{
					response.Validators[i],
					slashing.ValidatorSigningInfo{},
					GetConsAddressFromConsPubKey(response.Validators[i].ConsensusPubkey.Value),
				})
		}
	}

	bar.Finish()
	return validatorsList, nil
}

func GetValidatorsSigningInfo(conn *grpc.ClientConn, consAddress string) (result slashing.ValidatorSigningInfo, err error) {

	for retry := 0; retry < configs.Configs.GRPC.APICallRetry; retry++ {

		c := slashing.NewQueryClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		response, err := c.SigningInfo(ctx,
			&slashing.QuerySigningInfoRequest{ConsAddress: consAddress})
		if err != nil {
			fmt.Printf("\r[%d", retry+1)
			// fmt.Printf("\r\tRetrying [ %d ]...", retry+1)
			// fmt.Printf("\tErr: %s", err)

			// Ideally we want to retry after getting 502 http error, because sometimes server returns it
			// but we cannot have it as the protobuf Invoke does not return the status code
			time.Sleep(100 * time.Millisecond)
			continue
		}

		return response.ValSigningInfo, nil
	}

	return result, err
}

func GetActiveValidators(conn *grpc.ClientConn) (ValidatorsList, error) {
	return GetValidatorsList(conn, staking.BondStatusBonded)
}

// This function retrieves the consensus address from the consensus public key
func GetConsAddressFromConsPubKey(inKey []byte) string {

	// For some unknown reasons there are two extra bytes in the begining of the key
	// which cause the size error, so we remove them
	pubkey := &ed25519.PubKey{Key: inKey[2:]}
	return sdk.ConsAddress(pubkey.Address().Bytes()).String()
}

func GetValidatorInfoByAddress(conn *grpc.ClientConn, address string) (Validator, error) {

	for retry := 0; retry < configs.Configs.GRPC.APICallRetry; retry++ {

		c := staking.NewQueryClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		response, err := c.Validator(ctx,
			&staking.QueryValidatorRequest{
				ValidatorAddr: address,
			})
		if err != nil {
			fmt.Printf("\r[%d", retry+1)
			// fmt.Printf("\r\tRetrying [ %d ]...", retry+1)
			// fmt.Printf("\tErr: %s", err)

			// Ideally we want to retry after getting 502 http error, because sometimes server returns it
			// but we cannot have it as the protobuf Invoke does not return the status code
			time.Sleep(50 * time.Millisecond)
			continue
		}

		consAddr := GetConsAddressFromConsPubKey(response.Validator.ConsensusPubkey.Value)
		signingInfo, err := GetValidatorsSigningInfo(conn, consAddr)
		if err != nil {
			panic(err)
		}
		return Validator{response.Validator, signingInfo, consAddr}, nil
	}
	return Validator{}, fmt.Errorf("something went wrong")
}

func (v *Validator) GetAccountAddress() string {

	valAddr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
	if err != nil {
		log.Print(err)
	}
	return sdk.AccAddress(valAddr).String()
}
