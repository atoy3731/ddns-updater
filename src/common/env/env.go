package env

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type RequiredEnv struct {
	Name  string
	Value interface{}
}

type OptionalEnv struct {
	Name    string
	Default interface{}
	Value   interface{}
}

func NewRequiredEnv(name string) *RequiredEnv {
	return &RequiredEnv{
		name,
		nil,
	}
}

func NewOptionalEnv(name string, defaultValue interface{}) *OptionalEnv {
	return &OptionalEnv{
		name,
		defaultValue,
		nil,
	}
}

func ValidateRequired(requiredEnvs map[string]*RequiredEnv) {
	var missingEnvs []string
	for _, requiredEnv := range requiredEnvs {
		if os.Getenv(requiredEnv.Name) == "" {
			missingEnvs = append(missingEnvs, requiredEnv.Name)
		} else {
			requiredEnv.Value = os.Getenv(requiredEnv.Name)
		}
	}

	if len(missingEnvs) > 0 {
		log.Fatalln(fmt.Sprintf("Missing required Envs: %s", strings.Join(missingEnvs, ",")))
	}
}

func ValidateOptional(optionalEnvs map[string]*OptionalEnv) {
	for _, optionalEnv := range optionalEnvs {
		if os.Getenv(optionalEnv.Name) == "" {
			optionalEnv.Value = optionalEnv.Default
		} else {
			optionalEnv.Value = os.Getenv(optionalEnv.Name)
		}
	}
}

func IsProvider(provider string) bool {
	defProvider := os.Getenv("DNS_PROVIDER")

	if strings.ToLower(provider) == strings.ToLower(defProvider) {
		return true
	}

	return false
}
