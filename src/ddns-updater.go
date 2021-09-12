package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	env "github.com/atoy3731/ddns-updater/src/common/env"
	logging "github.com/atoy3731/ddns-updater/src/common/logging"

	// Providers
	cloudflare "github.com/atoy3731/ddns-updater/src/providers/cloudflare"
	dynu "github.com/atoy3731/ddns-updater/src/providers/dynu"
	noip "github.com/atoy3731/ddns-updater/src/providers/noip"
)

var (
	IpUrl        string
	ExistingIp   []byte
	IntervalMins int
	DnsProvider  string

	requiredEnvs map[string]*env.RequiredEnv
	optionalEnvs map[string]*env.OptionalEnv
)

func init() {
	logging.InfoLogger.Println("==============")
	logging.InfoLogger.Println(" DDNS Updater ")
	logging.InfoLogger.Println("==============")

	ExistingIp = []byte("N/A")

	requiredEnvs = map[string]*env.RequiredEnv{
		"DNS_PROVIDER": env.NewRequiredEnv("DNS_PROVIDER"),
	}

	optionalEnvs = map[string]*env.OptionalEnv{
		"INTERVAL_MINS": env.NewOptionalEnv("INTERVAL_MINS", 5),
		"IP_URL":        env.NewOptionalEnv("IP_URL", "https://checkip.amazonaws.com"),
		"LOG_LEVEL":     env.NewOptionalEnv("LOG_LEVEL", "info"),
	}

	env.ValidateRequired(requiredEnvs)
	env.ValidateOptional(optionalEnvs)

	DnsProvider = requiredEnvs["DNS_PROVIDER"].Value.(string)
	IpUrl = optionalEnvs["IP_URL"].Value.(string)
	IntervalMins = optionalEnvs["INTERVAL_MINS"].Value.(int)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getIp() []byte {
	logging.DebugLogger.Println(fmt.Sprintf("Getting IP from '%s'", IpUrl))
	resp, err := http.Get(IpUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	logging.DebugLogger.Println(fmt.Sprintf("Acquired IP: %s", strings.TrimSpace(string(body))))
	return body
}

func isIpChanged(current_ip []byte) bool {
	if string(ExistingIp) != string(current_ip) {
		logging.InfoLogger.Println(fmt.Sprintf("Updating IP (%s -> %s)", strings.TrimSpace(string(ExistingIp)), strings.TrimSpace(string(current_ip))))
		return true
	} else {
		logging.DebugLogger.Println(fmt.Sprintf("IP (%s) hasn't changed. No update required.", strings.TrimSpace(string(current_ip))))
	}

	return false
}

func updateDns(current_ip string) {
	switch provider := strings.ToLower(DnsProvider); provider {
	case "cloudflare":
		cloudflare.UpdateDns(current_ip)
	case "noip":
		noip.UpdateDns(current_ip)
	case "dynu":
		dynu.UpdateDns(current_ip)
	default:
		logging.ErrorLogger.Fatalln(fmt.Sprintf("Invalid provider '%s'", DnsProvider))
	}
}

func main() {
	logging.InfoLogger.Println(fmt.Sprintf("Using DNS provider: %s", strings.ToLower(DnsProvider)))
	logging.InfoLogger.Println(fmt.Sprintf("Running every %d minutes", IntervalMins))
	logging.InfoLogger.Println(fmt.Sprintf("Current IP: %s", strings.TrimSpace(string(getIp()))))

	for true {
		current_ip := getIp()

		if isIpChanged(current_ip) {
			updateDns(string(current_ip))
		}

		time.Sleep(time.Duration(IntervalMins) * time.Minute)
	}
}
