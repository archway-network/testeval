package winners

import (
	"fmt"

	"github.com/archway-network/testnet-evaluator/events"
	"google.golang.org/grpc"
)

// This function inspects the received tx to the `verifier_account`
// and checks its memo field for a particular data (e.g. `email address`)
// in order to match it with already existing data
// If the existing data matches with the account holder is sent, the account is verified
func (w *Winner) Verify(conn *grpc.ClientConn) (bool, error) {
	// events.GetTxEvents()

	response, err := events.GetTxEvents(conn,
		[]string{
			// "message.module='bank'",
			"message.action='/cosmos.bank.v1beta1.MsgSend'",
			fmt.Sprintf("message.sender='%s'", w.Address), //cosmos155svs6sgxe55rnvs6ghprtqu0mh69kehrn0dqr
			// fmt.Sprintf("coin_received.receiver='%s'", configs.Configs.IdVerification.VerifierAccount),
		}, 100, 0)
	if err != nil {
		return false, err
	}

	fmt.Printf("\nAddress: %s\n\n", w.Address)

	if response == nil ||
		response.TxResponses == nil ||
		len(response.TxResponses) == 0 {
		return false, nil
	}

	fmt.Println(response.TxResponses)
	return false, nil
}
