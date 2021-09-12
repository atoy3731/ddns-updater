package dynu

import (
	b64 "encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	env "github.com/atoy3731/ddns-updater/src/common/env"
	logging "github.com/atoy3731/ddns-updater/src/common/logging"
)

var (
	requiredEnvs map[string]env.RequiredEnv
	optionalEnvs map[string]env.OptionalEnv

	DynuUsername string
	DynuPassword string
	DynuHostname string
)

func init() {
	if env.IsProvider("noip") {
		requiredEnvs := map[string]*env.RequiredEnv{
			"DYNU_USERNAME": env.NewRequiredEnv("DYNU_USERNAME"),
			"DYNU_PASSWORD": env.NewRequiredEnv("DYNU_PASSWORD"),
			"DYNU_HOSTNAME": env.NewRequiredEnv("DYNU_HOSTNAME"),
		}

		optionalEnvs := map[string]*env.OptionalEnv{}

		env.ValidateRequired(requiredEnvs)
		env.ValidateOptional(optionalEnvs)

		DynuUsername = strings.TrimSpace(requiredEnvs["DYNU_USERNAME"].Value.(string))
		DynuPassword = strings.TrimSpace(requiredEnvs["DYNU_PASSWORD"].Value.(string))
		DynuHostname = strings.TrimSpace(requiredEnvs["DYNU_HOSTNAME"].Value.(string))
	}
}

func addAuthHeader(req *http.Request) {
	token := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", DynuUsername, DynuPassword)))
	header := fmt.Sprintf("Bearer %s", token)
	req.Header.Add("Authorization", header)
}

func UpdateDns(current_ip string) {
	client := &http.Client{}

	// Update NoIP hostname
	var url = fmt.Sprintf("https://api.dynu.com/nic/update?hostname=%s&myip=%s", DynuHostname, current_ip)
	var req, _ = http.NewRequest("GET", url, nil)

	addAuthHeader(req)

	var resp, resp_err = client.Do(req)
	if resp_err != nil {
		logging.ErrorLogger.Fatalln(resp_err)
		return
	} else if resp.StatusCode == 401 {
		logging.ErrorLogger.Fatalln("[ERROR] Unauthorized. Check your Dynu credentials!")
		return
	} else if resp.StatusCode != 200 {
		logging.ErrorLogger.Printf("[ERROR] Issue contacting Dynu")
		b, _ := io.ReadAll(resp.Body)
		logging.ErrorLogger.Fatalln(fmt.Sprintf("  -> Body: %s", string(b)))
		return
	}

	logging.InfoLogger.Println(fmt.Sprintf("DNS Updated (%s -> %s)", DynuHostname, strings.TrimSpace(string(string(current_ip)))))
}
