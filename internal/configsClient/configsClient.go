package configsClient

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	env "github.com/caarlos0/env/v6"
)

// default values
const (
	defaultgrpcAddress = ":3333"
	defaultDataFolder  = "./data/keeper"
	defaultClientToken = "DefaultClientToken"
)

// ServiceConfigs is list of parameters
type ClientConfigs struct {
	GRPCaddress string `json:"grpc_address" env:"IKEEPER_GRPC_ADDRESS"`
	FileFolder  string `json:"file_folder" env:"IKEEPER_DATA_FOLDER"`
	ClientToken string `json:"client_token" env:"IKEEPER_TOKEN"`
	ConfigFile  string `json:"-" env:"IKEEPER_CONFIG_FILE"`
}

var (
	ClientConfig = ClientConfigs{}
	onceUpload   sync.Once
)

// UploadConfigs gets parameters from Flags, ENV(priority), file.json(low priority).
func UploadConfigsClient() (*ClientConfigs, string) {
	onceUpload.Do(func() {

		// from flags
		ClientConfig.Flags()
		log.Println("configs after flags:", ClientConfig)

		// from environment
		ClientConfig.Environment()
		log.Println("configs after ENV:", ClientConfig)

		// from file.fson
		if ClientConfig.ConfigFile != "" {
			ClientConfig.FromConfigFile()
		}

		ClientConfig.SetDefaultConfigs()
	})
	return &ClientConfig, fmt.Sprintf("%s\nService address:%s\nFile folder:%s\n", printVersion(), ClientConfig.GRPCaddress, ClientConfig.FileFolder)
}

// SetDefaultConfigs set default parameters
func (sc *ClientConfigs) SetDefaultConfigs() {
	if ClientConfig.GRPCaddress == "" {
		ClientConfig.GRPCaddress = defaultgrpcAddress
	}
	if ClientConfig.FileFolder == "" {
		ClientConfig.FileFolder = defaultDataFolder
	}
	if ClientConfig.ClientToken == "" {
		ClientConfig.ClientToken = defaultClientToken
	}
}

// Environment check ENV for parameters
func (sc *ClientConfigs) Environment() {
	err := env.Parse(sc)
	if err != nil {
		log.Println("env.Parse error in config package:", err)
	}
}

// Flags read flags
func (sc *ClientConfigs) Flags() {
	// Server conf flags
	flag.StringVar(&sc.GRPCaddress, "grpc", sc.GRPCaddress, "gRPC server address")

	flag.StringVar(&sc.ConfigFile, "conf", sc.ConfigFile, "config file address (if not default)")

	flag.StringVar(&sc.FileFolder, "dir", sc.FileFolder, "ikeeper's home directory (if not default)")

	flag.StringVar(&sc.ClientToken, "token", sc.ClientToken, "ikeeper token (if not default)")

	flag.Parse()
}

// low priority reading from config file
func (sc *ClientConfigs) FromConfigFile() {
	confFilePath := sc.ConfigFile

	if confFilePath != "" {
		data, err := os.ReadFile(confFilePath)
		if err != nil {
			log.Println("config file read error:", err)
			return
		}

		envFromFile := new(ClientConfigs)
		err = json.Unmarshal(data, &envFromFile)
		if err != nil {
			log.Println("config file unmarshal error:", err)
			return
		}

		log.Println("configs from file: ", envFromFile)

		if sc.GRPCaddress == "" {
			sc.GRPCaddress = envFromFile.GRPCaddress
		}

		if sc.FileFolder == "" {
			sc.FileFolder = envFromFile.FileFolder
		}

		if sc.ClientToken == "" {
			sc.ClientToken = envFromFile.ClientToken
		}
	}
}

var (
	buildVersion string = "0.7"
	buildDate    string = "170623"
	buildCommit  string = "diploma"
)

// printVersion show current app version
func printVersion() string {
	return fmt.Sprintf("Build version: %s\nBuild date: %s\nBuild commit: %s",
		buildVersion,
		buildDate,
		buildCommit,
	)
}
