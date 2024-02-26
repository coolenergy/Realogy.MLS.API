package config

import (
	"fmt"
	"os"
	"strings"

	awsssm "github.com/PaddleHQ/go-aws-ssm"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	prom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	// Metrics registry
	PromRegistry = prometheus.NewRegistry()

	// Grpc server metrics
	PrometheusGrpcMetrics = prom.NewServerMetrics()
)

type Config struct {
	Api        ApiConfig        `mapstructure:"api"`
	MongoDB    MongoDBConfig    `mapstructure:"mongodb"`
	Log        LogConfig        `mapstructure:"log"`
	Grpc       GrpcConfig       `mapstructure:"grpc"`
	Gateway    GatewayConfig    `mapstructure:"gateway"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	Tracing    bool             `mapstructure:"tracing"`
}

type ApiConfig struct {
	Pagination PaginationConfig `mapstructure:"pagination"`
	Stream     StreamConfig     `mapstructure:"stream"`
	BySource   BySource         `mapstructure:"by_source"`
	ByAddress  ByAddress        `mapstructure:"by_address"`
	Auth       Auth             `mapstructure:"auth"`
}

type PaginationConfig struct {
	LimitDefault int32 `mapstructure:"limit_default"`
	LimitMax     int32 `mapstructure:"limit_max"`
}

type StreamConfig struct {
	DeadlineSecs int32 `mapstructure:"deadline_secs"`
}

type BySource struct {
	AllowedLastChangeDays int `mapstructure:"allowed_last_change_days"`
}

type ByAddress struct {
	SearchIndex string `mapstructure:"search_index"`
}

type GrpcConfig struct {
	Port    uint16 `mapstructure:"port"`
	Network string `mapstructure:"network"`
}

type GatewayConfig struct {
	Port    uint16 `mapstructure:"port"`
	Network string `mapstructure:"network"`
}

type PrometheusConfig struct {
	Port uint16 `mapstructure:"port"`
}

type Auth struct {
	AccessRules string `mapstructure:"accessRules"`
}

type MongoDBConfig struct {
	Prefix           string            `mapstructure:"prefix"`
	Url              string            `mapstructure:"url"`
	Name             string            `mapstructure:"name"`
	User             string            `mapstructure:"user"`
	Pass             string            `mapstructure:"pass"`
	Collections      map[string]string `mapstructure:"collections"`
	Options          string            `mapstructure:"options"`
	MaxQueryTimeSecs int               `mapstructure:"max_query_time_secs"`
}

type LogConfig struct {
	Formatter string `mapstructure:"formatter"`
	Level     string `mapstructure:"level"`
}

func (c *Config) initConfig(path string) error {

	// configs
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.SetConfigName("config") // name of configs file (without extension)
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)
	viper.SetEnvPrefix("go.mls")                           // Converts to env in the format, "GO_MLS_<PROPERTY_NAME>". Example: For "mongodb.url", env should be "GO_MLS_MONGODB_URL"
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // replace "_" to "." in the env var
	// defaults
	viper.SetDefault("grpc.port", 9080)
	viper.SetDefault("grpc.network", "tcp")
	viper.SetDefault("gateway.port", 9081)
	viper.SetDefault("gateway.network", "tcp")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("aws.region", "us-west-2")
	viper.SetDefault("aws.ssm.basepath", "")
	viper.SetDefault("aws.local", false)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("Unable to find configuration file: %s", err)
		} else {
			log.Fatalf("Unable to process configuration file: %s", err)
		}
	}

	awsRegion := viper.GetString("aws.region")
	awsConfig := &aws.Config{
		Region: aws.String(awsRegion),
	}
	awsLocal := viper.GetBool("aws.local.active")
	if awsLocal {
		log.Infof("Using localstack aws env [%v] for testing...", viper.GetString("aws.local.endpoint.ssm"))
		awsEndpointResolver := localAwsEndpointResolver("ssm", viper.GetString("aws.local.endpoint.ssm")) // for development testing.
		awsConfig.WithEndpointResolver(endpoints.ResolverFunc(awsEndpointResolver))
		awsConfig.WithRegion("us-west-2")
	}
	err := ReadAndMapParameterStore()(viper.GetViper(), awsConfig)
	if err != nil {
		log.Printf("Unable to access aws parameter store: %v. Is localstack enabled ? : %v, Using default values from config file.", err, awsLocal)

	}

	// Unmarshall configs
	err1 := viper.Unmarshal(c)
	if err1 != nil {
		return err1
	}

	return nil
}

func localAwsEndpointResolver(awsService string, awsEndpoint string) func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
	return func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		if service == awsService {
			return endpoints.ResolvedEndpoint{
				URL:           awsEndpoint,
				SigningRegion: region,
			}, nil
		}
		return endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
	}
}

func (c *Config) initLogging() error {
	level, err := log.ParseLevel(c.Log.Level)
	if err != nil {
		return err
	}
	log.SetOutput(os.Stdout)
	log.SetLevel(level)
	if strings.EqualFold(c.Log.Formatter, "json") {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	}
	return nil
}

func Load(path string) *Config {

	config := Config{}

	// prometheus grpc metrics
	PromRegistry.MustRegister(PrometheusGrpcMetrics)

	// configs
	if err := config.initConfig(path); err != nil {
		log.Fatalf("Error while decoding mls configs to struct, %v", err)
	}

	// logging
	if err := config.initLogging(); err != nil {
		log.Fatalf("Error while parsing the log level from the configs : %s", err)
	}

	return &config
}

// Read AWS parameter store for a given base path. If the config "aws.ssm.basepath" exists, then the function assumes the values has to be read from aws.
// In case of error the default values from the "config.yaml" should be used if exists.
func ReadAndMapParameterStore() func(v *viper.Viper, awsConfig *aws.Config) error {
	return func(v *viper.Viper, awsConfig *aws.Config) error {
		paramStore, _ := awsssm.NewParameterStore(awsConfig)
		params, err := paramStore.GetAllParametersByPath(v.GetString("aws.ssm.basepath"), true)
		if params != nil {
			log.Printf("Using aws parameter store values from [%v] for mongodb credentials", v.GetString("aws.ssm.basepath"))
			v.Set("mongodb.url", params.GetValueByName("mls-mongodb-url"))
			v.Set("mongodb.user", params.GetValueByName("mls-mongodb-user"))
			v.Set("mongodb.pass", params.GetValueByName("mls-mongodb-pass"))
		}
		return err
	}
}

func GenerateMongoUrl(config *MongoDBConfig) string {
	var uri = fmt.Sprintf("%s://%s@%s/", config.Prefix,
		config.User,
		config.Url,
	)

	if config.Options != "" {
		uri = uri + "?" + config.Options
	}
	return uri
}
