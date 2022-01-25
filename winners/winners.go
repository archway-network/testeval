package winners

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Winner struct {
	Address          string               // Account Address
	Rewards          uint64               // Total Reward of a winner
	Timestamp        string               // The time of the task done, if applicable
	TxResponse       *sdk.TxResponse      // The associated Tx Response with the task, if applicable
	Verified         bool                 // If the ID of this winner account is verified
	VerificationData verificationDataType // When we verify the user's data, we keep a copy of the verification data here for further investigation

}

type hashMapType map[string]int // This map is used for quick search { string: Winner.Address, int: index to the item in the WinnersList.list slice}

type WinnersList struct {
	list    []Winner // To keep the order of items we use the int index
	hashMap hashMapType
}

/*------------------*/

func (w *WinnersList) GetItem(index int) Winner {
	return w.list[index]
}

/*------------------*/

// This function merges two lists of winners if a winner already exist, it is not added
func (w *WinnersList) Merge(newList WinnersList) {
	for i := range newList.list {
		w.Append(newList.list[i])
	}
}

/*------------------*/

// This function merges two lists of winners and aggregates the rewards
// If a winner already exist, her/his rewards will be aggregated
func (w *WinnersList) MergeWithAggregateRewards(newList WinnersList) {
	for i := range newList.list {
		w.AppendWithAggregateRewards(newList.list[i])
	}
}

/*------------------*/

// This function cuts the end of a list
func (w WinnersList) Trim(length int) WinnersList {

	if w.Length() == length {
		return w
	}

	var trimmedList WinnersList
	for i := range w.list {
		trimmedList.Append(w.list[i])
		if i >= length-1 {
			break
		}
	}
	return trimmedList
}

/*------------------*/

func (w WinnersList) Length() int {
	return len(w.list)
}

/*------------------*/

// Return result:
// 		-1 : Not found
// 		>-1: The item index
func (w WinnersList) FindByAddress(address string) int {

	if index, ok := w.hashMap[address]; ok {
		return index
	}
	return -1
}

/*------------------*/

func (w *WinnersList) Append(item Winner) {

	if index := w.FindByAddress(item.Address); index == -1 {
		newIndex := w.Length()
		w.list = append(w.list, item)

		if w.hashMap == nil {
			w.hashMap = make(hashMapType)
		}
		w.hashMap[item.Address] = newIndex
	}
}

/*------------------*/

func (w *WinnersList) AppendWithAggregateRewards(item Winner) {

	if index := w.FindByAddress(item.Address); index != -1 {
		w.list[index].Rewards += item.Rewards
	} else {
		newIndex := w.Length()
		w.list = append(w.list, item)

		if w.hashMap == nil {
			w.hashMap = make(hashMapType)
		}
		w.hashMap[item.Address] = newIndex
	}
}

/*------------------*/

func (w WinnersList) Print() {

	for i := range w.list {
		fmt.Printf("%d \t%s\tRewards: %d \tTx: %s\n",
			i+1,
			w.list[i].Address,
			w.list[i].Rewards,
			w.list[i].TxResponse.TxHash[0:8]+"..."+w.list[i].TxResponse.TxHash[len(w.list[i].TxResponse.TxHash)-8:],
		)
	}
}

/*------------------*/

func (w *WinnersList) GetVerifiedOnly() WinnersList {

	var output WinnersList
	for i := range w.list {
		if w.list[i].Verified {
			output.Append(w.list[i])
		}
	}
	return output
}
