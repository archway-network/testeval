package report

// This package generates report of the rewards and winners

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/archway-network/testnet-evaluator/configs"
	"github.com/archway-network/testnet-evaluator/winners"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

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
	_, err = f.WriteString("Account,Reward")
	if err != nil {
		return err
	}

	for i := 0; i < winnersList.Length(); i++ {
		winner := winnersList.GetItem(i)

		redcord := fmt.Sprintf("\n%s,%d", winner.Address, winner.Rewards)

		// the Go CSV package has some issues, it missed some records
		_, err = f.WriteString(redcord)
		if err != nil {
			return err
		}
	}

	return nil
}

// should be called like this
// allWinners := map[string]*winners.WinnersList{
// 	"Active Validator": &validatorsWinnersList,
// 	"Jailed Unjailed":  &unjailWinnersList,
// 	"Governance":       &govWinnersList,
// 	"Staking":          &stakingWinnersList,
// }
// GenerateHTML(totalWinnersList, allWinners)
func GenerateHTML(mergedList winners.WinnersList, challenges map[string]*winners.WinnersList) error {

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

	homePageContent := getHTMLHeader("List of Winners", homePagePath)
	var homePageTableRows [][]string

	for i := 0; i < mergedList.Length(); i++ {
		winner := mergedList.GetItem(i)

		// <!-- Details pages

		pageFileName := getHTMLFileName("details", winner.Address)
		detailsPageFilePath, err := filepath.Abs(filepath.Join(htmlDir, pageFileName))
		if err != nil {
			return err
		}

		htmlContent := getHTMLHeader(fmt.Sprintf("Details of the winner: %s", winner.Address), homePagePath)
		tableHeaders := []string{"Challenge", "Reward", "More Info"}
		var tableRows [][]string
		for chName := range challenges {
			index := challenges[chName].FindByAddress(winner.Address)

			if index > -1 {
				chWinner := challenges[chName].GetItem(index)

				row := []string{chName, localePrint.Sprintf("%d", chWinner.Rewards)}

				if chWinner.TxResponse != nil {
					// fmt.Printf("\nRawLog of %s: %v", chName, chWinner.TxResponse)
					// row = append(row, chWinner.TxResponse.RawLog)
					row = append(row, fmt.Sprintf("%v", chWinner.TxResponse))
				} else {
					row = append(row, "")
				}

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
