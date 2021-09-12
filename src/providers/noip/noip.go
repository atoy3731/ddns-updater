package noip

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

	NoIpEmail    string
	NoIpPassword string
	NoIpHostname string
)

func init() {
	if env.IsProvider("noip") {
		requiredEnvs := map[string]*env.RequiredEnv{
			"NOIP_EMAIL":    env.NewRequiredEnv("NOIP_EMAIL"),
			"NOIP_PASSWORD": env.NewRequiredEnv("NOIP_PASSWORD"),
			"NOIP_HOSTNAME": env.NewRequiredEnv("NOIP_HOSTNAME"),
		}

		optionalEnvs := map[string]*env.OptionalEnv{}

		env.ValidateRequired(requiredEnvs)
		env.ValidateOptional(optionalEnvs)

		NoIpEmail = requiredEnvs["NOIP_EMAIL"].Value.(string)
		NoIpPassword = requiredEnvs["NOIP_PASSWORD"].Value.(string)
		NoIpHostname = requiredEnvs["NOIP_HOSTNAME"].Value.(string)
	}
}

func addAuthHeader(req *http.Request) {
	token := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", NoIpEmail, NoIpPassword)))
	header := fmt.Sprintf("Bearer %s", token)
	req.Header.Add("Authorization", header)
}

func UpdateDns(current_ip string) {
	client := &http.Client{}

	// Update NoIP hostname
	var url = fmt.Sprintf("http://dynupdate.no-ip.com/nic/update?hostname=%s&myip=%s", NoIpHostname, current_ip)
	var req, _ = http.NewRequest("GET", url, nil)

	addAuthHeader(req)

	var resp, resp_err = client.Do(req)
	if resp_err != nil {
		logging.ErrorLogger.Fatalln(resp_err)
		return
	} else if resp.StatusCode == 401 {
		logging.ErrorLogger.Fatalln("[ERROR] Unauthorized. Check your NoIP credentials!")
		return
	} else if resp.StatusCode != 200 {
		logging.ErrorLogger.Printf("[ERROR] Issue contacting NoIP")
		b, _ := io.ReadAll(resp.Body)
		logging.ErrorLogger.Fatalln(fmt.Sprintf("  -> Body: %s", string(b)))
		return
	}

	logging.InfoLogger.Println(fmt.Sprintf("DNS Updated (%s -> %s)", NoIpHostname, strings.TrimSpace(string(string(current_ip)))))
}
