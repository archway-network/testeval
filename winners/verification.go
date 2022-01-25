package winners

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/archway-network/testnet-evaluator/events"
	"github.com/archway-network/testnet-evaluator/progressbar"

	// staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	// "github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
)

/*----------------*/

type verificationDataType struct { //map[email]...
	Email string
	KYCId string
}

// Used for quick search
var verificationData map[string]verificationDataType

/*----------------*/

// This function inspects the received tx to the `verifier_account`
// and checks its memo field for a particular data (e.g. `email address`)
// in order to match it with already existing data
// If the existing data matches with the account holder is sent, the account is verified
func (w *Winner) Verify(conn *grpc.ClientConn) (bool, error) {

	// Check the cache first
	if verified, found := readVerificationCache(w.Address); found {
		return verified, nil
	}

	response, err := events.GetTxEvents(conn,
		[]string{
			"message.module='bank'",
			"message.action='/cosmos.bank.v1beta1.MsgSend'",
			fmt.Sprintf("message.sender='%s'", w.Address),
			fmt.Sprintf("coin_received.receiver='%s'", configs.Configs.IdVerification.VerifierAccount),
		}, 100, 0)
	if err != nil {
		return false, err
	}

	if response == nil ||
		response.TxResponses == nil ||
		len(response.TxResponses) == 0 {
		return false, nil
	}

	// We extract multiple transactions as the user might re-send a tx to correct a wrong data
	for i := range response.Txs {
		usersVerificationData, err := extractVerificationDataFromTxMemo(response.Txs[i].Body.Memo)
		if err != nil {
			return false, err
		}

		foundUserVerificationData, err := findVerificationDataByEmail(usersVerificationData.Email)
		if err != nil {
			return false, err
		}

		if foundUserVerificationData.Email != "" &&
			foundUserVerificationData.KYCId != "" &&
			foundUserVerificationData.Email == usersVerificationData.Email &&
			foundUserVerificationData.KYCId == usersVerificationData.KYCId {
			w.Verified = true
			w.VerificationData = foundUserVerificationData
			addToVerificationCache(w.Address, true)
			return true, nil
		}
	}

	addToVerificationCache(w.Address, false)
	return false, nil
}

/*-------------*/

var verificationCache map[string]bool // to cache the verification results
func addToVerificationCache(address string, verified bool) {

	if verificationCache == nil {
		verificationCache = make(map[string]bool)
	}
	verificationCache[address] = verified
}

func readVerificationCache(address string) (bool, bool) {
	verified, found := verificationCache[address]
	return verified, found
}

/*-------------*/

func loadVerificationData() (map[string]verificationDataType, error) {

	dataList := make(map[string]verificationDataType)

	f, err := os.Open(configs.Configs.IdVerification.InputFile.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)

	headers, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	emailIndex := -1
	kycIndex := -1
	for i := range headers {

		if headers[i] == configs.Configs.IdVerification.InputFile.Fields.Email {
			emailIndex = i
		}

		if headers[i] == configs.Configs.IdVerification.InputFile.Fields.KYCId {
			kycIndex = i
		}
	}

	for {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		dataList[line[emailIndex]] = verificationDataType{
			Email: line[emailIndex],
			KYCId: line[kycIndex],
		}
	}
	return dataList, nil
}

func findVerificationDataByEmail(email string) (verificationDataType, error) {

	// Load the csv file if we have not done it yet
	var err error
	if verificationData == nil {
		verificationData, err = loadVerificationData()
		if err != nil {
			return verificationDataType{}, err
		}
	}

	if foundItem, ok := verificationData[email]; ok {
		return foundItem, nil
	}

	return verificationDataType{}, nil
}

// We have to set a rule for example saying users need
// to enter their email address then and space followed by their KYC code
func extractVerificationDataFromTxMemo(memo string) (verificationDataType, error) {

	memo = strings.Trim(memo, " \r\n\t")
	if len(memo) == 0 {
		return verificationDataType{}, nil
	}

	memoFields := strings.Fields(memo)
	if len(memoFields) != 2 {
		return verificationDataType{}, fmt.Errorf("invalid Tx memo format")
	}
	return verificationDataType{
		Email: memoFields[0],
		KYCId: memoFields[1],
	}, nil
}

func (w *WinnersList) VerifyAll(conn *grpc.ClientConn) error {

	var bar progressbar.Bar
	bar.NewOption(0, int64(w.Length()))

	for i := 0; i < w.Length(); i++ {
		_, err := w.list[i].Verify(conn)
		if err != nil {
			return err
		}
		bar.Play(int64(i))
	}
	bar.Finish()
	return nil
}
