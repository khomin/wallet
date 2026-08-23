package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Server        ServerConfig     `mapstructure:"server"`
	Authorization Authorization    `mapstructure:"authorization"`
	Database      DatabaseConfig   `mapstructure:"database"`
	Redis         RedisConfig      `mapstructure:"redis"`
	Rabbit        RabbitMqConfig   `mapstructure:"rabbit"`
	Alchemy       AlchemyConfig    `mapstructure:"alchemy"`
	CoinGecko     CoinGeckoConfig  `mapstructure:"coingecko"`
	Email         EmailConfig      `mapstructure:"email"`
	Blockchain    BlockchainConfig `mapstructure:"blockchain"`
	TokenRegistry TokenRegistry    `mapstructure:"token_registry"`
}

type ServerConfig struct {
	PortHTTP    int    `mapstructure:"port_http"`
	PortGRPC    int    `mapstructure:"port_grpc"`
	Environment string `mapstructure:"environment"`
	LogPath     string `mapstructure:"log_path"`
}

type Authorization struct {
	IssuerURL string `mapstructure:"issuer_url"`
	ClientID  string `mapstructure:"client_id"`
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

type RabbitMqConfig struct {
	Url string `mapstructure:"url"`
}

type AlchemyConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type CoinGeckoConfig struct {
	APIKey       string        `mapstructure:"api_key"`
	PriceFetcher time.Duration `mapstructure:"price_fetcher"`
}

type EmailConfig struct {
	ApiKey string `mapstructure:"api_key"`
	From   string `mapstructure:"from"`
}

type BlockchainConfig struct {
	RateLimitConfig  ChainRateLimitConfig `mapstructure:"rate_limit"`
	EthereumMainnet  string               `mapstructure:"ethereum_mainnet_rpc"`
	EthereumArbitrum string               `mapstructure:"ethereum_arbitrum_rpc"`
	EthereumBase     string               `mapstructure:"ethereum_base_rpc"`
	PolygonMainnet   string               `mapstructure:"polygon_mainnet"`
	Bnb              string               `mapstructure:"bnb"`
	SolanaRPC        string               `mapstructure:"solana_rpc"`
	TronGRPC         string               `mapstructure:"tron_grpc"`
	TronAPIKey       string               `mapstructure:"tron_api_key"`
	Bitcoin          RPCConfig            `mapstructure:"bitcoin"`
	RippleMainnet    string               `mapstructure:"riplle_mainnet_rpc"`
}

type RPCConfig struct {
	Host string `mapstructure:"host"`
	User string `mapstructure:"user"`
	Pass string `mapstructure:"pass"`
}

type ChainRateLimitConfig struct {
	RPS   float64 `mapstructure:"rps"`
	Burst int     `mapstructure:"burst"`
}

type TokenConfig struct {
	Symbol      string                `yaml:"symbol"`
	Name        string                `yaml:"name"`
	Native      *NativeConfig         `yaml:"native,omitempty"`      // For L1 native gas tokens
	Deployments map[string]Deployment `yaml:"deployments,omitempty"` // For contract deployments
}

type NativeConfig struct {
	Chain    string `yaml:"chain"`
	Decimals int    `yaml:"decimals"`
}

type Deployment struct {
	Address  string `yaml:"address"`
	Decimals int    `yaml:"decimals"`
}

type TokenRegistry struct {
	Tokens []TokenConfig `yaml:"tokens"`
}

func NewConfig() *Config {
	config := Config{}
	// Safe local defaults for the mail-relay service exposed by docker-compose.
	// Every value can be overridden in config.yaml or with TRACKER_EMAIL_* env vars.
	viper.SetDefault("email.smtp_host", "localhost")
	viper.SetDefault("email.smtp_port", 25)
	viper.SetDefault("email.from", "alerts@localhost")
	viper.SetEnvPrefix("TRACKER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatalf("can't find the file: %s", err.Error())
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		logrus.Fatalf("environment can't be loaded: %s", err.Error())
	}
	logrus.Infof("environment ready")
	return &config
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
