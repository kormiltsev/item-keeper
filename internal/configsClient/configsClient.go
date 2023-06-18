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

// default values
var (
	presetgrpcAddress = ""
	presetDataFolder  = ""
	presetClientToken = ""
)

// ServiceConfigs is list of parameters
type ClientConfigs struct {
	GRPCaddress string `json:"grpc_address" env:"IKEEPER_GRPC_ADDRESS"`
	FileFolder  string `json:"file_folder" env:"IKEEPER_DATA_FOLDER"`
	ClientToken string `json:"client_token" env:"IKEEPER_TOKEN"`
	ConfigFile  string `json:"-" env:"IKEEPER_CONFIG_FILE"`
}

// Client's confiogs
var (
	ClientConfig = ClientConfigs{}
	onceUpload   sync.Once
)

/*
	version and other app info. Presetted with building process:

go build -o bin/ikeeper-amd64-darwin -ldflags "
-X 'github.com/kormiltsev/item-keeper/internal/configsClient.buildVersion=1.0'
-X 'github.com/kormiltsev/item-keeper/internal/configsClient.buildDate=$(date +'%Y/%m/%d %H:%M:%S')'
-X 'github.com/kormiltsev/item-keeper/internal/configsClient.buildCommit=commit'
-X 'github.com/kormiltsev/item-keeper/internal/configsClient.presetgrpcAddress=188.227.85.116:3333'
-X 'github.com/kormiltsev/item-keeper/internal/configsClient.presetClientToken=darwin'" ./cmd/client/client.go
*/
var (
	buildVersion string = "0.7 (default)"
	buildDate    string = "no data set"
	buildCommit  string = "default"
)

// UploadConfigs gets parameters from Flags, ENV(priority), file.json(low priority).
func UploadConfigsClient() (*ClientConfigs, string) {
	onceUpload.Do(func() {

		// from flags
		ClientConfig.Flags()
		//log.Println("configs after flags:", ClientConfig)

		// from environment
		ClientConfig.Environment()
		//log.Println("configs after ENV:", ClientConfig)

		// from file.fson
		if ClientConfig.ConfigFile != "" {
			ClientConfig.FromConfigFile()
		}

		ClientConfig.SetDefaultConfigs()
	})
	return &ClientConfig, fmt.Sprintf("%s\nService address:%s\nFile folder:%s\n", PrintVersion(), ClientConfig.GRPCaddress, ClientConfig.FileFolder)
}

// SetDefaultConfigs set default parameters
func (sc *ClientConfigs) SetDefaultConfigs() {
	if presetgrpcAddress != "" {
		ClientConfig.GRPCaddress = presetgrpcAddress
	} else {
		if ClientConfig.GRPCaddress == "" {
			ClientConfig.GRPCaddress = defaultgrpcAddress
		}
	}

	if presetDataFolder != "" {
		ClientConfig.FileFolder = presetDataFolder
	} else {
		if ClientConfig.FileFolder == "" {
			ClientConfig.FileFolder = defaultDataFolder
		}
	}

	if presetClientToken != "" {
		ClientConfig.ClientToken = presetClientToken
	} else {
		if ClientConfig.ClientToken == "" {
			ClientConfig.ClientToken = defaultClientToken
		}
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

		//log.Println("configs from file: ", envFromFile)

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

// PrintVersion show current app version
func PrintVersion() string {
	return fmt.Sprintf("Build version: %s\nBuild date: %s\nBuild commit: %s",
		buildVersion,
		buildDate,
		buildCommit,
	)
}
