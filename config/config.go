package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Alchemy    AlchemyConfig    `mapstructure:"alchemy"`
	CoinGecko  CoinGeckoConfig  `mapstructure:"coingecko"`
	Blockchain BlockchainConfig `mapstructure:"blockchain"`
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
	LogPath     string `mapstructure:"log_path"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AlchemyConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type CoinGeckoConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type BlockchainConfig struct {
	EthereumMainnet  string `mapstructure:"ethereum_mainnet_rpc"`
	EthereumArbitrum string `mapstructure:"ethereum_arbitrum_rpc"`
	EthereumBase     string `mapstructure:"ethereum_base_rpc"`
	PolygonMainnet   string `mapstructure:"polygon_mainnet"`
	Bnb              string `mapstructure:"bnb"`
	SolanaRPC        string `mapstructure:"solana_rpc"`
	BitcoinRPCHost   string `mapstructure:"bitcoin_rpc_host"`
	BitcoinRPCUser   string `mapstructure:"bitcoin_rpc_user"`
	BitcoinRPCPass   string `mapstructure:"bitcoin_rpc_pass"`
	TronGRPC         string `mapstructure:"tron_grpc"`
	TronAPIKey       string `mapstructure:"tron_api_key"`
}

func NewConfig() *Config {
	env := Config{}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatalf("can't find the file: %s", err.Error())
	}
	err = viper.Unmarshal(&env)
	if err != nil {
		logrus.Fatalf("environment can't be loaded: %s", err.Error())
	}
	logrus.Infof("environment ready")
	return &env
}

func (c *Config) DSN() string {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
	)
	return dsn
}
