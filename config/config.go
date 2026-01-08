package config

import (
	"fmt"
	"log"
	"net/url"

	"github.com/spf13/viper"
)

var loaded = false

func LoadAndValidateConfiguration(file string) {
	if loaded {
		return
	}

	loaded = true

	viper.SetConfigFile(file)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(fmt.Errorf("fatal error reading config: %w", err))
	}

	validateConfiguration()
}

func validateConfiguration() {
	validateServerSection()
	validateMicropubSection()
	validatePersistenceSection()
}

func validateServerSection() {
	address := viper.GetString("server.address")
	if address == "" || address[len(address)-1] != ':' {
		log.Fatal("server.address: invalid value, should not be empty and should end with colon (\":\")")
	}

	port := viper.GetUint("server.port")
	if port == 0 {
		log.Fatal("server.port: invalid value, should be a positive whole number (a valid network port)")
	}

	maxPayloadSize := MaxPayloadSize()
	if maxPayloadSize == 0 {
		log.Fatal("server.limits.maxPayloadSize: invalid value, should be defined as positive bytes value")
	}

	maxFileSize := MaxFileSize()
	if maxFileSize == 0 {
		log.Fatal("server.limits.maxFileSize: invalid value, should be defined as positive bytes value")
	}

	maxMultipartSize := MaxMultipartSize()
	if maxMultipartSize == 0 {
		log.Fatal("server.limits.maxMultipartSize: invalid value, should be defined as positive bytes value")
	}
}

func validateMicropubSection() {
	meUrl := MeUrl()
	if meUrl == "" {
		log.Fatal("micropub.me_url: should be defined")
	} else if _, err := url.ParseRequestURI(meUrl); err != nil {
		log.Fatal("micropub.me_url: should be a valid URL")
	}

	tokenEndpoint := TokenEndpoint()
	if tokenEndpoint == "" {
		log.Fatal("micropub.token_endpoint: should be defined")
	} else if _, err := url.ParseRequestURI(tokenEndpoint); err != nil {
		log.Fatal("micropub.token_endpoint: should be a valid URL")
	}
}

func validatePersistenceSection() {
	strategy := PersistenceStrategy()
	if strategy != "static" {
		log.Fatal("persistence.strategy: invalid value, should be \"static\"")
	}

	validatePersistenceStaticSection()
}

func validatePersistenceStaticSection() {
	method := PersistenceStaticMethod()
	if method != "git" {
		log.Fatal("persistence.static.method: invalid value, should be \"git\"")
	}

	validatePersistenceStaticGitSection()

	// TODO: add other static method checking -- sftp, http
}

func validatePersistenceStaticGitSection() {
	repository := GitRepository()
	if repository == "" {
		log.Fatal("persistence.static.git.repository: should be defined")
	} else if _, err := url.ParseRequestURI(repository); err != nil {
		log.Fatal("persistence.static.git.repository: should be a valid URL")
	}

	username := GitUsername()
	if username == "" {
		log.Fatal("persistence.static.git.username: should be defined")
	}

	password := GitPassword()
	if password == "" {
		log.Fatal("persistence.static.git.password: should be defined")
	}

	directory := GitDirectory()
	if directory == "" {
		log.Fatal("persistence.static.git.directory: should be defined")
	}
}

func BindAddress() string {
	return fmt.Sprintf("%v%v", viper.GetString("server.address"), viper.GetUint("server.port"))
}

func MaxPayloadSize() uint {
	return viper.GetUint("server.limits.maxPayloadSize")
}

func MaxFileSize() uint {
	return viper.GetUint("server.limits.maxFileSize")
}

func MaxMultipartSize() uint {
	return viper.GetUint("server.limits.maxMultipartSize")
}

func TokenEndpoint() string {
	return viper.GetString("micropub.token_endpoint")
}

func MeUrl() string {
	return viper.GetString("micropub.me_url")
}

func Debug() bool {
	return viper.GetBool("debug")
}

func PersistenceStrategy() string {
	return viper.GetString("persistence.strategy")
}

func PersistenceStaticMethod() string {
	return viper.GetString("persistence.static.method")
}

func GitRepository() string {
	return viper.GetString("persistence.static.git.repository")
}

func GitUsername() string {
	return viper.GetString("persistence.static.git.username")
}

func GitPassword() string {
	return viper.GetString("persistence.static.git.password")
}

func GitDirectory() string {
	return viper.GetString("persistence.static.git.directory")
}
