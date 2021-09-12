package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	env "github.com/atoy3731/ddns-updater/src/common/env"
	"github.com/atoy3731/ddns-updater/src/common/logging"
)

type CFRequest struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

type CFResult struct {
	Id string `json:"id"`
}

type CFResponse struct {
	Result []CFResult `json:"result"`
}

var (
	requiredEnvs map[string]env.RequiredEnv
	optionalEnvs map[string]env.OptionalEnv

	CloudflareZone   string
	CloudflareRecord string
	CloudflareToken  string
	CloudflareDnsTTL int
)

func init() {
	if env.IsProvider("cloudflare") {
		requiredEnvs := map[string]*env.RequiredEnv{
			"CLOUDFLARE_TOKEN":  env.NewRequiredEnv("CLOUDFLARE_TOKEN"),
			"CLOUDFLARE_ZONE":   env.NewRequiredEnv("CLOUDFLARE_ZONE"),
			"CLOUDFLARE_RECORD": env.NewRequiredEnv("CLOUDFLARE_RECORD"),
		}

		optionalEnvs := map[string]*env.OptionalEnv{
			"CLOUDFLARE_DNS_TTL": env.NewOptionalEnv("CLOUDFLARE_DNS_TTL", 1),
		}

		env.ValidateRequired(requiredEnvs)
		env.ValidateOptional(optionalEnvs)

		CloudflareZone = requiredEnvs["CLOUDFLARE_ZONE"].Value.(string)
		CloudflareRecord = requiredEnvs["CLOUDFLARE_RECORD"].Value.(string)
		CloudflareToken = requiredEnvs["CLOUDFLARE_TOKEN"].Value.(string)

		CloudflareDnsTTL = optionalEnvs["CLOUDFLARE_DNS_TTL"].Value.(int)
	}
}

func addAuthHeader(req *http.Request) {
	header := fmt.Sprintf("Bearer %s", CloudflareToken)
	req.Header.Add("Authorization", header)
}

func UpdateDns(current_ip string) {
	client := &http.Client{}

	// Get Zone ID
	var url = fmt.Sprintf("https://api.cloudflare.com/client/v4/zones?name=%s&status=active", CloudflareZone)
	var req, _ = http.NewRequest("GET", url, nil)

	addAuthHeader(req)

	var resp, resp_err = client.Do(req)
	if resp_err != nil {
		logging.ErrorLogger.Fatalln(resp_err)
		return
	} else if resp.StatusCode == 403 {
		logging.ErrorLogger.Fatalln("[ERROR] Unauthorized. Check your Cloudflare token!")
		return
	} else if resp.StatusCode != 200 {
		logging.ErrorLogger.Printf("[ERROR] Issue contacting Cloudflare")
		b, _ := io.ReadAll(resp.Body)
		logging.ErrorLogger.Fatalln(fmt.Sprintf("  -> Body: %s", string(b)))
		return
	}

	var cfResp = CFResponse{}

	defer resp.Body.Close()
	var body, _ = io.ReadAll(resp.Body)
	json.Unmarshal(body, &cfResp)
	zoneId := cfResp.Result[0].Id
	logging.DebugLogger.Println(fmt.Sprintf("Zone ID for '%s' is '%s'", CloudflareZone, zoneId))

	// Get Record ID
	url = fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=A&name=%s", string(zoneId), CloudflareRecord)
	req, _ = http.NewRequest("GET", url, nil)
	addAuthHeader(req)

	logging.DebugLogger.Println(fmt.Sprintf("Getting Record ID for '%s'", CloudflareRecord))
	resp, resp_err = client.Do(req)
	if resp_err != nil {
		logging.ErrorLogger.Fatalln(resp_err)
		return
	} else if resp.StatusCode == 403 {
		logging.ErrorLogger.Fatalln("[ERROR] Unauthorized. Check your Cloudflare token!")
		return
	} else if resp.StatusCode != 200 {
		logging.ErrorLogger.Fatalln("[ERROR] Issue contacting Cloudflare")
		b, _ := io.ReadAll(resp.Body)
		logging.ErrorLogger.Println(fmt.Sprintf("  -> Body: %s", string(b)))
		return
	}
	cfResp = CFResponse{}

	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body)
	json.Unmarshal(body, &cfResp)
	recordId := cfResp.Result[0].Id
	logging.DebugLogger.Println(fmt.Sprintf("Record ID for '%s' is '%s'", CloudflareRecord, recordId))

	// Update record
	var cfReq = CFRequest{}
	cfReq.Type = "A"
	cfReq.Name = CloudflareRecord
	cfReq.Content = strings.TrimSpace(string(current_ip))
	cfReq.TTL = CloudflareDnsTTL
	cfReq.Proxied = false

	cfReqJson, _ := json.Marshal(cfReq)
	url = fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", string(zoneId), string(recordId))
	req, _ = http.NewRequest("PUT", url, bytes.NewBuffer([]byte(cfReqJson)))
	addAuthHeader(req)
	req.Header.Add("Content-type", "application/json")

	logging.DebugLogger.Println(fmt.Sprintf("Updating DNS record for '%s'", CloudflareRecord))
	resp, resp_err = client.Do(req)
	if resp_err != nil {
		logging.ErrorLogger.Fatalln(resp_err)
		return
	} else if resp.StatusCode == 403 {
		logging.ErrorLogger.Fatalln("[ERROR] Unauthorized. Check your Cloudflare token!")
		return
	} else if resp.StatusCode != 200 {
		logging.ErrorLogger.Printf("[ERROR] Issue contacting Cloudflare")
		b, _ := io.ReadAll(resp.Body)
		logging.ErrorLogger.Fatalln(fmt.Sprintf("  -> Body: %s", string(b)))
		return
	}

	logging.InfoLogger.Println(fmt.Sprintf("DNS Updated (%s -> %s)", CloudflareRecord, strings.TrimSpace(string(string(current_ip)))))

}
