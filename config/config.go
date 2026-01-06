package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

var loaded = false

func LoadConfiguration(file string) {
	if loaded {
		return
	}

	loaded = true

	viper.SetConfigFile(file)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(fmt.Errorf("scribble: fatal error reading config: %w", err))
	}
}

func BindAddress() string {
	address := viper.GetString("server.address")
	if address == "" {
		address = ":"
		logValueWarning("server address", address)
	}

	port := viper.GetUint("server.port")
	if port == 0 {
		port = 8080
		logValueWarning("server port", port)
	}

	return fmt.Sprintf("%v%v", address, port)
}

func TokenEndpoint() string {
	endpoint := viper.GetString("micropub.token_endpoint")
	if endpoint == "" {
		endpoint := "https://tokens.indieauth.com"
		logValueWarning("token endpoint", endpoint)
	}

	if !strings.HasPrefix(endpoint, "http") {
		log.Panicf("token endpoint may be invalid - detected no http prefix")
	}

	return endpoint
}

func MeUrl() string {
	meUrl := viper.GetString("micropub.me_url")
	if meUrl == "" {
		meUrl = "http://example.org"
		logValueWarning("\"me\" url", "")
		log.Println("warning: access tokens will not be validated! please define your \"me\" url")
	}

	return meUrl
}

func Debug() bool {
	return viper.GetBool("debug")
}

func logValueWarning(name string, def any) {
	log.Printf("warning: %v not set or unreadable - using default %v", name, def)
}
