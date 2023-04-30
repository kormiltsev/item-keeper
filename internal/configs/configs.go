package configs

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	env "github.com/caarlos0/env/v6"
)

// default values
const (
	defaultgrpcport = ":3333"
	defaultdb       = "mock"
)

// ServiceConfigs is list of parameters
type ServiceConfigs struct {
	GRPCport   string `json:"grpc_port" env:"GRPC_PORT"`
	DBlink     string `json:"database_dsn" env:"DATABASE_DSN"`
	ConfigFile string `json:"-" env:"CONFIG_FILE"`
}

var (
	ServiceConfig = ServiceConfigs{}
	onceUpload    sync.Once
	debugmod      bool
)

// UploadConfigs gets parameters from Flags, ENV(priority), file.json(low priority).
func UploadConfigs() *ServiceConfigs {
	onceUpload.Do(func() {

		// from flags
		ServiceConfig.Flags()
		log.Println("configs after flags:", ServiceConfig)

		// from environment
		ServiceConfig.Environment()
		log.Println("configs after ENV:", ServiceConfig)

		// from file.fson
		if ServiceConfig.ConfigFile != "" {
			ServiceConfig.FromConfigFile()
		}

		ServiceConfig.SetDefaultConfigs()

		// in case `env:"GRPC_PORT"` = 8080 (port only) or host:port/something
		host, port, err := net.SplitHostPort(ServiceConfig.GRPCport)
		if err != nil {
			a := strings.Split(ServiceConfig.GRPCport, ":")
			switch len(a) {
			case 0:
				ServiceConfig.GRPCport = defaultgrpcport //default
			case 1:
				ServiceConfig.GRPCport = ":" + a[0]
			case 3:
				ServiceConfig.GRPCport = a[1] + ":" + a[2]
			}
		} else {
			ServiceConfig.GRPCport = host + ":" + port
		}
	})
	return &ServiceConfig
}

// SetDefaultConfigs set default parameters
func (sc ServiceConfigs) SetDefaultConfigs() {
	if ServiceConfig.GRPCport == "" {
		ServiceConfig.GRPCport = defaultgrpcport
	}
	if ServiceConfig.DBlink == "" || debugmod {
		ServiceConfig.DBlink = defaultdb
	}
}

// Environment check ENV for parameters
func (sc *ServiceConfigs) Environment() {
	err := env.Parse(sc)
	if err != nil {
		log.Println("env.Parse error in config package:", err)
	}
}

// Flags read flags
func (sc *ServiceConfigs) Flags() {
	// Server conf flags
	flag.StringVar(&sc.GRPCport, "grpcport", sc.GRPCport, "gRPC server port")
	flag.StringVar(&sc.GRPCport, "g", sc.GRPCport, "gRPC server port (shorthand)")

	flag.StringVar(&sc.ConfigFile, "configfile", sc.ConfigFile, "config file address")
	flag.StringVar(&sc.ConfigFile, "c", sc.ConfigFile, "config file address (shorthand)")

	flag.StringVar(&sc.DBlink, "dblink", sc.DBlink, "database dsn")
	flag.StringVar(&sc.DBlink, "d", sc.DBlink, "database dsn (shorthand)")

	flag.BoolVar(&debugmod, "mock", debugmod, "debug case. using mockDB")

	flag.Parse()
}

// low priority reading from config file
func (sc *ServiceConfigs) FromConfigFile() {
	confFilePath := sc.ConfigFile

	if confFilePath != "" {
		data, err := os.ReadFile(confFilePath)
		if err != nil {
			log.Println("config file read error:", err)
			return
		}

		envFromFile := new(ServiceConfigs)
		err = json.Unmarshal(data, &envFromFile)
		if err != nil {
			log.Println("config file unmarshal error:", err)
			return
		}

		log.Println("configs from file: ", envFromFile)

		if sc.GRPCport == "" {
			sc.GRPCport = envFromFile.GRPCport
		}

		if sc.DBlink == "" {
			sc.DBlink = envFromFile.DBlink
		}
	}
}
