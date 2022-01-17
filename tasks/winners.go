package tasks

import (
	"fmt"

	"github.com/archway-network/testnet-evaluator/types"
)

type WinnersList []types.Winner

/*------------------*/

// This function merges two lists of winners without considering any internal data e.g. address, reward, etc.
func (w *WinnersList) Merge(newList WinnersList) {
	for i := range newList {
		w.Append(newList[i])
	}
}

/*------------------*/

// This function sums the total reward of each address in one list and returns a distinct list
func (w WinnersList) AggregateRewards() WinnersList {
	var resWinnerList WinnersList
	for i := range w {

		foundIndex := resWinnerList.FindByAddress(w[i].Address)
		if foundIndex == -1 { // Not found

			// When combile multiple tasks, having these variables is not relevant
			w[i].Timestamp = ""
			w[i].TxResponse = nil
			resWinnerList.Append(w[i])

		} else { // Already added to the list, so let's just sum the rewards

			resWinnerList[foundIndex].Rewards += w[i].Rewards
		}
	}

	return resWinnerList
}

/*------------------*/

func (w WinnersList) Distinct() WinnersList {

	var distinctList WinnersList
	for i := range w {
		if distinctList.FindByAddress(w[i].Address) == -1 {
			distinctList.Append(w[i])
		}
	}
	return distinctList
}

/*------------------*/

// This function cuts the end of a list
func (w WinnersList) Trim(length int) WinnersList {

	var trimmedList WinnersList
	for i := range w {
		trimmedList.Append(w[i])
		if i >= length-1 {
			break
		}
	}
	return trimmedList
}

/*------------------*/

func (w WinnersList) Length() int {
	return len(w)
}

/*------------------*/

// Return result:
// 		-1 : Not found
// 		>-1: The item index
func (w WinnersList) FindByAddress(address string) int {

	for i := range w {
		if w[i].Address == address {
			return i
		}
	}
	return -1
}

/*------------------*/

func (w *WinnersList) Append(item types.Winner) {
	*w = append(*w, item)
}

/*------------------*/

func (w WinnersList) Print() {

	for i := range w {
		fmt.Printf("%d \t%s\tRewards: %d \tTx: %s\n", i+1, w[i].Address, w[i].Rewards, w[i].TxResponse.TxHash[0:8]+"..."+w[i].TxResponse.TxHash[len(w[i].TxResponse.TxHash)-8:])
	}
}
