package report

// This package generates report of the rewards and winners

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/archway-network/testnet-evaluator/validators"
	"github.com/archway-network/testnet-evaluator/winners"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type WinnersListReport struct {
	Title         string
	WinnersList   *winners.WinnersList
	ValidatorOnly bool
}

func init() {
	err := os.MkdirAll(configs.Configs.Report.OutputDir, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("\nThe output directory: `%s`\n", configs.Configs.Report.OutputDir)
	}
}

func StoreWinnersCSV(winnersList winners.WinnersList) error {

	f, err := os.Create(filepath.Join(configs.Configs.Report.OutputDir, "winners.csv"))
	if err != nil {
		return err
	}
	defer f.Close()

	// The headers
	if configs.Configs.IdVerification.Required {
		_, err = f.WriteString("Email,Account,Reward")
	} else {
		_, err = f.WriteString("Account,Reward")
	}
	if err != nil {
		return err
	}

	for i := 0; i < winnersList.Length(); i++ {
		winner := winnersList.GetItem(i)

		var redcord string

		if configs.Configs.IdVerification.Required {

			redcord = fmt.Sprintf("\n%s,%s,%d", winner.VerificationData.Email, winner.Address, winner.Rewards)
		} else {
			redcord = fmt.Sprintf("\n%s,%d", winner.Address, winner.Rewards)
		}

		// the Go CSV package has some issues, it missed some records
		_, err = f.WriteString(redcord)
		if err != nil {
			return err
		}
	}

	return nil
}

// should be called like this
// allWinners := []report.WinnersListReport{
// 	{Title: "Active Validator", WinnersList: &validatorsWinnersList, ValidatorOnly: true},
// 	{Title: "Jailed Unjailed", WinnersList: &unjailWinnersList, ValidatorOnly: true},
// 	{Title: "Governance", WinnersList: &govWinnersList, ValidatorOnly: false},
// 	{Title: "Staking", WinnersList: &stakingWinnersList, ValidatorOnly: false},
// }
// GenerateHTML(totalWinnersList, allWinners)
func GenerateHTML(mergedList winners.WinnersList, challenges []WinnersListReport) error {

	homePagePath, err := filepath.Abs(filepath.Join(configs.Configs.Report.OutputDir, "start.html"))
	if err != nil {
		return err
	}

	htmlDir := filepath.Join(configs.Configs.Report.OutputDir, "html")
	err = os.MkdirAll(htmlDir, os.ModePerm)
	if err != nil {
		return err
	}
	localePrint := message.NewPrinter(language.English)

	homePageContent := getHTMLHeader("List of Winners", "<b>List of Winners</b>", homePagePath)
	var homePageTableRows [][]string

	for i := 0; i < mergedList.Length(); i++ {
		winner := mergedList.GetItem(i)

		// <!-- Details pages

		pageFileName := getHTMLFileName("details", winner.Address)
		detailsPageFilePath, err := filepath.Abs(filepath.Join(htmlDir, pageFileName))
		if err != nil {
			return err
		}

		bExplorerLink := fmt.Sprintf(`Details of the winner: <b><a target="_blank" href="%s">%s</a></b>`,
			fmt.Sprintf(configs.Configs.BlockExplorer.Account, winner.Address),
			winner.Address)

		htmlContent := getHTMLHeader("Details of the winner", bExplorerLink, homePagePath)

		//<!-- Verification Info
		if configs.Configs.IdVerification.HTMLReport {
			if winner.Verified {
				htmlContent += getHTMLInfoBox("ID Verification:", "This participant is verified.")
				htmlContent += getHTMLTable([]string{"Identification data", ""}, [][]string{
					{"Email address", winner.VerificationData.Email},
					{"KYC ID", winner.VerificationData.KYCId},
				}, nil)
			} else {
				htmlContent += getHTMLWarningBox("ID Verification:", "This participant is NOT verified!")
			}
		}
		//-->

		tableHeaders := []string{"Challenge", "Reward", "More Info"}
		var tableRows [][]string
		for chIndex := range challenges {
			index := challenges[chIndex].WinnersList.FindByAddress(winner.Address)

			if index > -1 {
				chWinner := challenges[chIndex].WinnersList.GetItem(index)

				row := []string{challenges[chIndex].Title, localePrint.Sprintf("%d", chWinner.Rewards)}

				moreInfo := ""
				if chWinner.TxResponse != nil {
					// fmt.Printf("\nRawLog of %s: %v", chIndex, chWinner.TxResponse)
					// row = append(row, chWinner.TxResponse.RawLog)
					moreInfo += fmt.Sprintf("<pre>%v</pre>", chWinner.TxResponse)
					if chWinner.TxResponse.TxHash != "" {
						moreInfo += fmt.Sprintf(`<br /><a target="_blank" href="%s">%s</a>`,
							fmt.Sprintf(configs.Configs.BlockExplorer.TxHash, chWinner.TxResponse.TxHash),
							chWinner.TxResponse.TxHash)
					}
				}
				if challenges[chIndex].ValidatorOnly {
					valAddr := validators.GetValidatorAddressFromAccountAddr(chWinner.Address)
					moreInfo += fmt.Sprintf(`<br /><a target="_blank" href="%s">View Validator on Block Explorer</a>`,
						fmt.Sprintf(configs.Configs.BlockExplorer.Validator, valAddr))
				}
				row = append(row, moreInfo)

				tableRows = append(tableRows, row)
			}

		}
		tableFooters := []string{"Total Reward", localePrint.Sprintf("%d", winner.Rewards), ""}

		htmlContent += getHTMLTable(tableHeaders, tableRows, tableFooters)
		htmlContent += getHTMLFooter()

		err = writeTextToFile(detailsPageFilePath, htmlContent)
		if err != nil {
			return err
		}

		// End of Detail page-->

		detailsLink := fmt.Sprintf(`<a href="%s">%s</a>`, detailsPageFilePath, winner.Address)
		// bExplorerLink = fmt.Sprintf(`<a target="_blank" href="%s">%s</a>`,
		// 	fmt.Sprintf(configs.Configs.BlockExplorer.Account, winner.Address),
		// 	winner.Address)
		homePageTableRows = append(homePageTableRows,
			[]string{detailsLink, localePrint.Sprintf("%d", winner.Rewards)},
		)
	}

	homePageContent += getHTMLTable(
		[]string{"Winner", "Total Rewards"},
		homePageTableRows,
		nil)
	homePageContent += getHTMLFooter()
	err = writeTextToFile(homePagePath, homePageContent)
	if err != nil {
		return err
	}

	return nil
}

func getHTMLFileName(prefix string, winnerAddr string) string {
	return strings.ToLower(strings.ReplaceAll(prefix, " ", "_")) + "_" + winnerAddr + ".html"
}

func writeTextToFile(filePath, text string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(text)
	if err != nil {
		return err
	}

	return nil
}
