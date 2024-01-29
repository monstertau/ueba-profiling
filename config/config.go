package config

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"time"
	"ueba-profiling/util"
)

const (
	DefaultConfigFilePath = "./config.yml"

	// DefRedisAddr default redis
	DefRedisAddr     = "localhost:8379"
	DefRedisDb       = 0
	DefRedisPassword = ""

	// DefServiceHost default endpoint address
	DefServiceHost = "0.0.0.0"
	DefServicePort = 9999

	DefMongoAddr = "0.0.0.0"
	DefMongoPort = "27027"
	DefMongoUser = "kc"
	DefMongoPass = "098poiA#"
	DefMongoDb   = "kian"

	DefAsnDbFilePath        = "resources/GeoLite2-ASN_20200908/GeoLite2-ASN.mmdb"
	DefCityDbFilePath       = "resources/GeoLite2-City_20200908/GeoLite2-City.mmdb"
	DefCountryDbFilePath    = "resources/GeoLite2-Country_20200908/GeoLite2-Country.mmdb"
	DefPublicSuffixFilePath = "resources/TLDs.csv"
	DefEnWords              = "resources/ML_Model/english_words.txt"
	DefViWords              = "resources/ML_Model/Viet74K.txt"
	DefRegexFilePath        = "resources/regexes.yaml"

	DefEndpointGetJobs   = "http://localhost:9090/api/v2/worker/stateless/jobs"
	DefEndpointGetJobsV3 = "http://localhost:9090/api/v3/worker/stateless/jobs"

	DefLogSourceConsumerChanSize = 1000000
	DefKafkaInputChanSize        = 1000
	DefDestinationChannelSize    = 100000
	DefMessageChanSize           = 10000
	DefKafkaBatchSize            = 100

	DefKafkaServer = "10.255.250.190:9092"

	DefTimestampField = "local_timestamp"
	DefEventIDField   = "event_id"
	DefTaggingField   = "id"

	DefNumSnapshotWorker = 8

	FieldSfilter = "sfilter"
	TypeFilter   = "filter"
)

var AppConfig *Config

func init() {
	AppConfig = DefaultConfig()
}

type (
	Config struct {
		Service            *NodeConfig         `mapstructure:"service" json:"service"`
		Redis              *Redis              `mapstructure:"redis" json:"redis"`
		RedisAsset         *Redis              `mapstructure:"redis_asset" json:"redis_asset"`
		MongoDB            *MongoDBConfig      `mapstructure:"mongodb" json:"mongodb"`
		Resource           *Resource           `mapstructure:"resource" json:"resource"`
		Endpoint           *Endpoint           `mapstructure:"endpoint" json:"endpoint"`
		Channel            *Channel            `mapstructure:"channel_config" json:"channel_config"`
		TimestampField     string              `mapstructure:"timestamp_field" json:"timestamp_field"`
		EventIdField       string              `mapstructure:"eventid_field" json:"eventid_field"`
		TaggingField       string              `mapstructure:"tagging_field" json:"tagging_field"`
		KafkaLogSource     *KafkaConfig        `mapstructure:"kafka" json:"kafka"`
		Elastic            *ElasticConfig      `mapstructure:"elastic" json:"elastic"`
		ProfileModel       *ProfileModelConfig `mapstructure:"profile_model" json:"profile_model"`
		TracingConfig      *TracingConfig      `mapstructure:"tracing" json:"tracing"`
		NumConsumer        int                 `mapstructure:"num_consumer" json:"num_consumer"`
		KafkaNotify        *KafkaConfig        `mapstructure:"kafka_notify" json:"kafka_notify"`
		InfluxDB           *InfluxDBConfig     `mapstructure:"influxdb" json:"influxdb"`
		Destination        *DestinationConfig  `mapstructure:"destination" json:"destination"`
		DumpSnapshot       *DumpSnapshotConfig `mapstructure:"dump_snapshot" json:"dump_snapshot"`
		Tenant             *TenantConfig       `mapstructure:"tenant" json:"tenant"`
		RedisClientTimeout int                 `mapstructure:"redis_client_timeout" json:"redis_client_timeout"`
	}

	TracingConfig struct {
		ExpirationTime        string                 `mapstructure:"expiration_time" json:"expiration_time"`
		Type                  string                 `mapstructure:"type" json:"type"`
		Config                map[string]interface{} `mapstructure:"config" json:"config"`
		Workers               int                    `mapstructure:"workers" json:"workers"`
		NotifyChannel         string                 `mapstructure:"notify_channel" json:"notify_channel"`
		LimitSession          int                    `mapstructure:"limit_session" json:"limit_session"`
		SessionExpirationTime string                 `mapstructure:"session_expiration_time" json:"session_expiration_time"`
	}

	SessionRiskScoreFilter struct {
		MongoCol string `mapstructure:"mongo_col" json:"mongo_col"`
	}
	SessionFilterConfig struct {
		MongoCol string `mapstructure:"mongo_col" json:"mongo_col"`
	}

	Redis struct {
		Addr     string `mapstructure:"addr" json:"addr"`
		Db       int    `mapstructure:"db" json:"db"`
		Password string `mapstructure:"password" json:"password"`
	}

	NodeConfig struct {
		Host                  string                 `mapstructure:"host" json:"host"`
		Port                  int                    `mapstructure:"port" json:"port"`
		ServiceRegistryConfig *ServiceRegistryConfig `mapstructure:"service_registry" json:"service_registry"`
	}

	ServiceRegistryConfig struct {
		Type        string                      `mapstructure:"type" json:"type"`
		Addr        string                      `mapstructure:"addr" json:"addr"`
		Scheme      string                      `mapstructure:"scheme" json:"scheme"`
		Enable      bool                        `mapstructure:"enable" json:"enable"`
		Name        string                      `mapstructure:"name" json:"name"`
		Tags        []string                    `mapstructure:"tags" json:"tags"`
		Metadata    *ServiceRegistryMetadata    `mapstructure:"metadata" json:"metadata"`
		HealthCheck *ServiceRegistryHealthCheck `mapstructure:"health_check" json:"health_check"`
	}

	ServiceRegistryHealthCheck struct {
		URL      string `mapstructure:"url" json:"url"`
		Interval string `mapstructure:"interval" json:"interval"`
		Timeout  string `mapstructure:"timeout" json:"timeout"`
	}

	ServiceRegistryMetadata struct {
		ClusterID     string `mapstructure:"cluster_id" json:"cluster_id"`
		ConsumerGroup string `mapstructure:"kafka_consumer_group" json:"kafka_consumer_group"`
	}

	MongoDBConfig struct {
		Addr        string `mapstructure:"addr" json:"addr"`
		Port        string `mapstructure:"port" json:"port"`
		Username    string `mapstructure:"user" json:"user"`
		Password    string `mapstructure:"pass" json:"pass"`
		DBName      string `mapstructure:"db" json:"db"`
		ReadTimeOut int    `mapstructure:"read_time_out" json:"read_time_out"`
	}

	Resource struct {
		MaxMindAsn     string `mapstructure:"maxmind_asn" json:"maxmind_asn"`
		MaxMindCountry string `mapstructure:"maxmind_country" json:"maxmind_country"`
		IPLocationPath string `mapstructure:"ip_location" json:"ip_location"`
		Tld            string `mapstructure:"tld" json:"tld"`
		EnglishWords   string `mapstructure:"english_words" json:"english_words"`
		VietWords      string `mapstructure:"viet_words" json:"viet_words"`
		RegexFilePath  string `mapstructure:"regex_file" json:"regex_file"`
	}

	Endpoint struct {
		GetJobs string `mapstructure:"get_job" json:"get_job"`
	}

	Channel struct {
		LogSourceConsumerSize  int64 `mapstructure:"log_source_consumer_size" json:"log_source_consumer_size"`
		KafkaConsumerSize      int   `mapstructure:"kafka_consumer_size" json:"kafka_consumer_size"`
		MessageSize            int64 `mapstructure:"message_size" json:"message_size"`
		DestinationChannelSize int64 `mapstructure:"destination_channel_size" json:"destination_channel_size"`
	}
	KafkaConfig struct {
		BootstrapServers        string `mapstructure:"bootstrap_servers" json:"bootstrap_servers"`
		Topics                  string `mapstructure:"topics" json:"topics"`
		GroupID                 string `mapstructure:"group_id" json:"group_id"`
		PoolCapSize             int    `mapstructure:"pool_cap_size" json:"pool_cap_size"`
		PoolLenSize             int    `mapstructure:"pool_len_size" json:"pool_len_size"`
		RecvChanSize            int    `mapstructure:"recv_chan_size" json:"recv_chan_size"`
		TmpChanSize             int    `mapstructure:"tmp_chan_size" json:"tmp_chan_size"`
		ProducerBatchSize       int    `mapstructure:"producer_batch_size" json:"producer_batch_size"`
		ProducerCompressionType int    `mapstructure:"producer_compression_type" json:"producer_compression_type"`
		ProducerAcks            int    `mapstructure:"producer_acks" json:"producer_acks"`
		ProducerAsync           bool   `mapstructure:"producer_async" json:"producer_async"`
	}
	ElasticConfig struct {
		Addr string `mapstructure:"addr" json:"addr"`
	}
	ProfileModelConfig struct {
		PersistDuration  string              `mapstructure:"persist_duration" json:"persist_duration"`
		ReloadDuration   string              `mapstructure:"reload_duration" json:"reload_duration"`
		PersistDirectory string              `mapstructure:"persist_directory" json:"persist_directory"`
		ProfileLayer     *ProfileLayerConfig `mapstructure:"profile_layer" json:"profile_layer"`
	}

	ProfileLayerConfig struct {
		Enable         bool   `mapstructure:"enable" json:"enable"`
		CacheDirectory string `mapstructure:"cache_directory" json:"cache_directory"`
		CacheMaxSize   int    `mapstructure:"cache_max_size" json:"cache_max_size"`
		SyncDuration   string `mapstructure:"sync_duration" json:"sync_duration"`
	}

	InfluxDBConfig struct {
		Addr                    string `mapstructure:"addr" json:"addr"`
		AuthToken               string `mapstructure:"authtoken" json:"authtoken"`
		BucketName              string `mapstructure:"bucketname" json:"bucketname"`
		UsecaseStatsMeasurement string `mapstructure:"usecase_stats_measurement" json:"usecase_stats_measurement"`
	}

	DestinationConfig struct {
		NumSnapshotWorker int   `mapstructure:"num_snapshot_worker" json:"num_snapshot_worker"`
		QueueDuration     int64 `mapstructure:"queue_duration" json:"queue_duration"` // in milliseconds
		QueueElementLimit int   `mapstructure:"queue_element_limit" json:"queue_element_limit"`
		PoolSliceCap      int   `mapstructure:"pool_slice_cap" json:"pool_slice_cap"`
	}
	DumpSnapshotConfig struct {
		Enable                 bool   `mapstructure:"enable" json:"enable"`
		NumSnapshotWorker      int    `mapstructure:"num_snapshot_worker" json:"num_snapshot_worker"`
		SnapshotInputChanSize  int    `mapstructure:"snapshot_input_chan_size" json:"snapshot_input_chan_size"`
		DumpSnapshotThroughput int    `mapstructure:"dump_snapshot_throughput" json:"dump_snapshot_throughput"`
		DumpDuration           string `mapstructure:"dump_duration" json:"dump_duration"`
		ESBulkSizeLimit        int64  `mapstructure:"esbulk_size_limit" json:"esbulk_size_limit"`
		ESBulkElementLimit     int64  `mapstructure:"esbulk_element_limit" json:"esbulk_element_limit"`
		ESDocSizeLimit         int64  `mapstructure:"esdoc_size_limit" json:"esdoc_size_limit"`
		SamplingCount          int64  `mapstructure:"sampling_count" json:"sampling_count"`
	}

	TenantConfig struct {
		LogField string `mapstructure:"log_field" json:"log_field"`
	}
)

func (c *DumpSnapshotConfig) ParseDumpDuration() time.Duration {
	d, err := util.ParseDurationExtended(c.DumpDuration)
	if err != nil {
		panic(fmt.Sprintf("invalid dump_duration: %v", err))
	}
	return d
}

func (c *ProfileModelConfig) ParsePersistDuration() time.Duration {
	d, err := util.ParseDurationExtended(c.PersistDuration)
	if err != nil {
		panic(fmt.Sprintf("invalid persist_duration: %v", err))
	}
	return d
}

func (c *ProfileModelConfig) ParseReloadDuration() time.Duration {
	d, err := util.ParseDurationExtended(c.ReloadDuration)
	if err != nil {
		panic(fmt.Sprintf("invalid reload_duration: %v", err))
	}
	return d
}

func (c *ProfileLayerConfig) ParseSyncDuration() time.Duration {
	d, err := util.ParseDurationExtended(c.SyncDuration)
	if err != nil {
		panic(fmt.Sprintf("invalid sync_duration: %v", err))
	}
	return d
}

func LoadFile(filepath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(filepath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, errors.New("configuration file not found")
		} else {
			return nil, errors.Wrap(err, "ReadInConfig")
		}
	}
	config := DefaultConfig()
	if err := v.Unmarshal(&config); err != nil {
		return nil, errors.Wrap(err, "viper.Unmarshal")
	}
	AppConfig = config
	return config, nil
}

func DefaultNodeConfig() *NodeConfig {
	return &NodeConfig{
		Host:                  DefServiceHost,
		Port:                  DefServicePort,
		ServiceRegistryConfig: DefaultServiceRegistryConfig(),
	}
}

func DefaultServiceRegistryConfig() *ServiceRegistryConfig {
	return &ServiceRegistryConfig{
		Addr:     "localhost:8500",
		Type:     "consul",
		Scheme:   "http",
		Enable:   false,
		Name:     "kc_stateless",
		Tags:     []string{"stateless"},
		Metadata: DefaultServiceRegistryMetadata(),
	}
}

func DefaultServiceRegistryMetadata() *ServiceRegistryMetadata {
	return &ServiceRegistryMetadata{
		ClusterID:     "stateless_cluster_1",
		ConsumerGroup: "kc_stateless_csg_1",
	}
}

func DefaultMongoDBConfig() *MongoDBConfig {
	return &MongoDBConfig{
		Addr:        DefMongoAddr,
		Port:        DefMongoPort,
		Username:    DefMongoUser,
		Password:    DefMongoPass,
		DBName:      DefMongoDb,
		ReadTimeOut: 60, // 1 minute
	}
}

func DefaultRedisConfig() *Redis {
	return &Redis{
		Addr:     DefRedisAddr,
		Db:       DefRedisDb,
		Password: DefRedisPassword,
	}
}

func DefaultResourceConfig() *Resource {
	return &Resource{
		MaxMindAsn:     DefAsnDbFilePath,
		MaxMindCountry: DefCountryDbFilePath,
		Tld:            DefPublicSuffixFilePath,
		EnglishWords:   DefEnWords,
		VietWords:      DefViWords,
		RegexFilePath:  DefRegexFilePath,
	}
}

func DefaultChannelConfig() *Channel {
	return &Channel{
		LogSourceConsumerSize:  DefLogSourceConsumerChanSize,
		KafkaConsumerSize:      DefKafkaInputChanSize,
		MessageSize:            DefMessageChanSize,
		DestinationChannelSize: DefDestinationChannelSize,
	}
}

func DefaultEndpoint() *Endpoint {
	return &Endpoint{
		GetJobs: DefEndpointGetJobs,
	}
}

//func DefaultOutputConfig() *OutputConfig {
//	return &OutputConfig{
//		Type: "kafka",
//		Config: map[string]interface{}{
//			"bootstrap_servers": "localhost:9092",
//			"topics":            "KIAN-stateless",
//			"producer_timeout":  "60s",
//		},
//	}
//}

func DefaultKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		GroupID:                 "KIAN-stateless",
		PoolCapSize:             2000,
		PoolLenSize:             1000,
		RecvChanSize:            1000,
		TmpChanSize:             1000,
		ProducerBatchSize:       200000,
		ProducerCompressionType: 3, // 3 for Lz4
		ProducerAcks:            1,
		ProducerAsync:           false,
	}
}

func DefaultKafkaNotifyConfig() *KafkaConfig {
	return &KafkaConfig{
		BootstrapServers:        "localhost:9092",
		Topics:                  "notify_dump_profile",
		PoolCapSize:             2000,
		PoolLenSize:             1000,
		RecvChanSize:            1000,
		TmpChanSize:             1000,
		ProducerBatchSize:       200000,
		ProducerCompressionType: 3, // 3 for Lz4
		ProducerAcks:            1,
		ProducerAsync:           false,
	}
}

func DefaultElasticConfig() *ElasticConfig {
	return &ElasticConfig{
		Addr: "localhost:9200",
	}
}

func DefaultProfileModel() *ProfileModelConfig {
	return &ProfileModelConfig{
		PersistDuration:  "1d",
		ReloadDuration:   "1d",
		PersistDirectory: "KIAN-profile-models",
		ProfileLayer: &ProfileLayerConfig{
			Enable:         false,
			CacheDirectory: "KIAN-profile-models/KIAN-profile-layer",
			CacheMaxSize:   1000,
			SyncDuration:   "30s",
		},
	}
}

//func DefaultDebugConfig() *OutputConfig {
//	return &OutputConfig{
//		Type: "redis",
//		Config: map[string]interface{}{
//			"addr":     "localhost:8379",
//			"password": "",
//			"db":       "0",
//		},
//		Workers: 1,
//	}
//}

func DefaultTracingConfig() *TracingConfig {
	return &TracingConfig{
		ExpirationTime: "1d",
		Type:           "redis",
		Config: map[string]interface{}{
			"addr":     "localhost:8379",
			"password": "",
			"db":       "0",
		},
		Workers:               4,
		NotifyChannel:         "tracing_notify",
		LimitSession:          50,
		SessionExpirationTime: "1d",
	}
}

func DefaultInfluxDBConfig() *InfluxDBConfig {
	return &InfluxDBConfig{
		Addr:                    "127.0.0.1:8086",
		AuthToken:               "kian:098poiA#",
		BucketName:              "kian",
		UsecaseStatsMeasurement: "kian_stats",
	}
}
func DefaultConfig() *Config {
	return &Config{
		Service:        DefaultNodeConfig(),
		MongoDB:        DefaultMongoDBConfig(),
		Redis:          DefaultRedisConfig(),
		RedisAsset:     DefaultRedisConfig(),
		Channel:        DefaultChannelConfig(),
		Endpoint:       DefaultEndpoint(),
		Resource:       DefaultResourceConfig(),
		TimestampField: DefTimestampField,
		EventIdField:   DefEventIDField,
		TaggingField:   DefTaggingField,
		KafkaLogSource: DefaultKafkaConfig(),
		Elastic:        DefaultElasticConfig(),
		ProfileModel:   DefaultProfileModel(),
		TracingConfig:  DefaultTracingConfig(),
		NumConsumer:    DefNumSnapshotWorker,
		KafkaNotify:    DefaultKafkaNotifyConfig(),
		InfluxDB:       DefaultInfluxDBConfig(),
		DumpSnapshot: &DumpSnapshotConfig{
			Enable:                 true,
			NumSnapshotWorker:      5,
			SnapshotInputChanSize:  16,
			DumpSnapshotThroughput: 16, // should == DumpSnapshotThroughput
			DumpDuration:           "1m",
			ESBulkSizeLimit:        1024 * 1024 * 20,  // 20MB
			ESBulkElementLimit:     10000,             // 10k
			ESDocSizeLimit:         1024 * 1024 * 100, // 100MB
			SamplingCount:          1000,              // sample 1k elements
		},
		Destination: &DestinationConfig{
			NumSnapshotWorker: 1,
			QueueDuration:     500, // 0.1s
			QueueElementLimit: 1000,
			PoolSliceCap:      2048,
		},
		Tenant: &TenantConfig{
			LogField: "tenant",
		},
		RedisClientTimeout: 30,
	}
}

func GetConfig() *Config {
	if AppConfig == nil {
		panic("init config first")
	}
	return AppConfig
}
