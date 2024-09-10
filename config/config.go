package config

import (
	"github.com/spf13/viper"
)

var AppConfig Config

func InitConfig() {
	viper.AutomaticEnv()

	// Default config
	viper.SetDefault("LIBP2P_PORT", 4001)
	viper.SetDefault("LIBP2P_BOOT_NODES", "")
	viper.SetDefault("THRESHOLD", 2)
	viper.SetDefault("PROPOSER", false)
	viper.SetDefault("RELAYER_PRI_KEY", "node_private_key.pem")

	AppConfig = Config{
		Libp2pPort:      viper.GetInt("LIBP2P_PORT"),
		Libp2pBootNodes: viper.GetString("LIBP2P_BOOT_NODES"),
		Threshold:       viper.GetInt("THRESHOLD"),
		Proposer:        viper.GetBool("PROPOSER"),
		RelayerPriKey:   viper.GetString("RELAYER_PRI_KEY"),
	}

}

type Config struct {
	Libp2pPort      int
	Libp2pBootNodes string
	Threshold       int
	Proposer        bool
	RelayerPriKey   string
}
