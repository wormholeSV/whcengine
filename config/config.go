package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bcext/gcash/chaincfg"
	"github.com/copernet/whccommon/model"
	"github.com/spf13/viper"
)

const (
	ConfEnv        = "WHC_CONF"
	ConfTestEnv    = "WHC_TEST_CONF"
	ProjectLastDir = "whcengine"
	LockKey        = "chain_fetch_lock"
)

var conf *configuration

type configuration struct {
	GoVersion string `mapstructure:"go_version"`
	Version   string `mapstructure:"version"`
	TestNet   bool   `mapstructure:"testnet"`
	Private   struct {
		TickerSeconds    time.Duration    `mapstructure:"ticker_seconds"`
		FirstBlockHeight map[string]int64 `mapstructure:"first_block_height"`
	}
	DB    *model.DBOption
	Redis *model.RedisOption
	Log   *model.LogOption
	RPC   *model.RPCOption
	Tx    struct {
		MiniOutput float64 `mapstructure:"mini_output"`
	}
}

func GetFirstBlockHeight() int64 {
	conf = GetConf()

	net := "mainnet"
	if conf.TestNet {
		net = "testnet"
	}
	return conf.Private.FirstBlockHeight[net]
}

func GetConf() *configuration {
	if conf != nil {
		return conf
	}

	config := &configuration{}
	viper.SetEnvPrefix("whc")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	viper.SetDefault("conf", "./conf.yml")

	// get config file path from environment
	confFile := viper.GetString("conf")

	var realPath string

	// conf.go unit testing
	if viper.GetString("test_conf") != "" {
		realPath = viper.GetString("test_conf")
	} else {
		path, err := filepath.Abs("./")
		if err != nil {
			panic(err)
		}

		lastIndex := strings.Index(path, ProjectLastDir) + len(ProjectLastDir)
		correctPath := path[:lastIndex]
		realPath = filepath.Join(correctPath, confFile)
	}

	// parse config
	file, err := os.Open(realPath)
	if err != nil {
		panic("Open config file error: " + err.Error())
	}
	defer file.Close()

	err = viper.ReadConfig(file)
	if err != nil {
		panic("Read config file error: " + err.Error())
	}

	err = viper.Unmarshal(config)
	if err != nil {
		panic("Parse config file error: " + err.Error())
	}

	// TODO validate configuration
	//helper.Must(nil, config.Validate())

	conf = config
	return config
}

func GetChainParam() *chaincfg.Params {
	conf := GetConf()
	if conf.TestNet {
		return &chaincfg.TestNet3Params
	}

	return &chaincfg.MainNetParams
}
