package config

import (
	"fmt"

	"github.com/spf13/viper"
	"time"
	"ueba-profiling/util"
)

const (
	DefConfigFilePath = "config.yml"

	// default configuration
	// extractor
	DefAsnDbFilePath        = "resources/GeoLite2-ASN_20200908/GeoLite2-ASN.mmdb"
	DefCountryDbFilePath    = "resources/GeoLite2-Country_20200908/GeoLite2-Country.mmdb"
	DefPublicSuffixFilePath = "resources/TLDs.csv"
	DefRegexFilePath        = "resources/regexes.yaml"

	// default redis
	DefRedisAddr     = "localhost:6379"
	DefRedisDb       = 0
	DefRedisPassword = ""

	// default endpoint address
	DefServerHost = "0.0.0.0"
	DefServerPort = "9199"

	DefMongoAddr       = "0.0.0.0"
	DefMongoPort       = "27027"
	DefMongoUser       = "kc"
	DefMongoPass       = "098poiA#"
	DefMongoDb         = "kian"
	DefGridfsChunkSize = 1024 // The default size of each chunk of GridFS files in bytes: 1 MB

	DefEndpointGetJobs = "http://localhost:9090/api/v2/worker/profiling/jobs"

	DefLogSourceConsumerChanSize = 1000000
	DefBehaviorConsumerChanSize  = 1000000
	DefKafkaInputChanSize        = 1000
	DefMessageChanSize           = 10000
	DefTspField                  = "local_timestamp"
	DefEventIdField              = "event_id"

	DefRangeTimeRandom  = "300000ms"
	DefKafkaNotifyHost  = "10.255.250.164:9092"
	DefKafkaNotifyTopic = "notify_dump_profile"

	DefBatchExpireTickerDuration   = "1h" // every 1 hour
	DefKeepOldBatchDuration        = "1d"
	DefBatchExpireWorkerThroughput = 24
	DefForceServerTimestamp        = false
	DefKafkaPoolLenSize            = 1000
	DefKafkaPoolCapSize            = 2000
	DefMaxEventPerSaving           = 10000
	DefWatermarkInterval           = "30s"
	DefNumConsumer                 = 2
	DefNumRouterWorker             = 10
	KafkaTypeSegmentio             = "segmentio"
	KafkaTypeSarama                = "sarama"
)

var GlobalConfig *AppConfig

// AppConfig is the application configuration
type AppConfig struct {
	// extractor configuration
	AsnDbFilePath               string
	CountryDbFilePath           string
	PublicSuffixFilePath        string
	RegexFilePath               string
	RedisAddr                   string
	RedisDb                     int
	RedisPassword               string
	ServerHost                  string
	ServerPort                  int
	MongoAddr                   string
	MongoPort                   int
	MongoUser                   string
	MongoPassword               string
	MongoDb                     string
	MongoGridfsChunkSize        int64
	MongoRangeTimeRandom        string
	EndpointGetJobs             string
	LogSourceConsumerChanSize   int
	BehaviorConsumerChanSize    int
	KafkaInputChanSize          int
	MessageChanSize             int
	DefaultTimestampField       string
	DefaultEventIdField         string
	KafkaNotifyHost             string
	KafkaNotifyTopic            string
	KafkaLogsourceGroupIDPrefix string
	BatchExpireTickerDuration   string
	KeepOldBatchDuration        string
	BatchExpireWorkerThroughput int64
	ForceServerTimeStamp        bool
	KafkaPoolLenSize            int
	KafkaPoolCapSize            int
	MaxEventPerSaving           int
	WatermarkInterval           string
	NumConsumer                 int
	NumRouterWorker             int
	ConsumerType                string
}

func NewAppConfigFrom(configFilePath string) (*AppConfig, error) {
	appConfig := &AppConfig{}

	// reading configuration
	viper.SetConfigName(configFilePath)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("resource.maxmind_asn", DefAsnDbFilePath)
	viper.SetDefault("resource.maxmind_country", DefCountryDbFilePath)
	viper.SetDefault("resource.tld", DefPublicSuffixFilePath)
	viper.SetDefault("resource.regex_file", DefRegexFilePath)
	viper.SetDefault("redis.addr", DefRedisAddr)
	viper.SetDefault("redis.db", DefRedisDb)
	viper.SetDefault("redis.password", DefRedisPassword)
	viper.SetDefault("server.host", DefServerHost)
	viper.SetDefault("server.port", DefServerPort)
	viper.SetDefault("mongo.addr", DefMongoAddr)
	viper.SetDefault("mongo.port", DefMongoPort)
	viper.SetDefault("mongo.user", DefMongoUser)
	viper.SetDefault("mongo.password", DefMongoPass)
	viper.SetDefault("mongo.db", DefMongoDb)
	viper.SetDefault("mongo.gridfs_chunk_size", DefGridfsChunkSize)
	viper.SetDefault("mongo.range_time_random", DefRangeTimeRandom)

	viper.SetDefault("endpoint.get_jobs", DefEndpointGetJobs)

	viper.SetDefault("channel_config.log_source_consumer_size", DefLogSourceConsumerChanSize)
	viper.SetDefault("channel_config.behavior_source_consumer_size", DefBehaviorConsumerChanSize)
	viper.SetDefault("channel_config.kafka_input_size", DefKafkaInputChanSize)
	viper.SetDefault("channel_config.message_size", DefMessageChanSize)
	viper.SetDefault("event_config.timestamp_field", DefTspField)
	viper.SetDefault("event_config.event_id", DefEventIdField)
	viper.SetDefault("notify.host", DefKafkaNotifyHost)
	viper.SetDefault("notify.topic", DefKafkaNotifyTopic)
	viper.SetDefault("batch_expire_ticker_duration", DefBatchExpireTickerDuration)
	viper.SetDefault("keep_old_batch_duration", DefKeepOldBatchDuration)
	viper.SetDefault("batch_expire_worker_throughput", DefBatchExpireWorkerThroughput)
	viper.SetDefault("force_server_timestamp", DefForceServerTimestamp)
	viper.SetDefault("kafka_logsource.pool_cap_size", DefKafkaPoolCapSize)
	viper.SetDefault("kafka_logsource.pool_len_size", DefKafkaPoolLenSize)
	viper.SetDefault("max_event_per_saving", DefMaxEventPerSaving)
	viper.SetDefault("watermark_interval", DefWatermarkInterval)
	viper.SetDefault("num_consumer", DefNumConsumer)
	viper.SetDefault("num_router_worker", DefNumRouterWorker)
	viper.SetDefault("consumer_type", KafkaTypeSegmentio)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	appConfig.AsnDbFilePath = viper.GetString("resource.maxmind_asn")
	appConfig.CountryDbFilePath = viper.GetString("resource.maxmind_country")
	appConfig.PublicSuffixFilePath = viper.GetString("resource.tld")
	appConfig.RegexFilePath = viper.GetString("resource.regex_file")

	appConfig.RedisAddr = viper.GetString("redis.addr")
	appConfig.RedisDb = viper.GetInt("redis.db")
	appConfig.RedisPassword = viper.GetString("redis.password")
	appConfig.ServerHost = viper.GetString("server.host")
	appConfig.ServerPort = viper.GetInt("server.port")

	appConfig.MongoAddr = viper.GetString("mongo.addr")
	appConfig.MongoPort = viper.GetInt("mongo.port")
	appConfig.MongoUser = viper.GetString("mongo.user")
	appConfig.MongoPassword = viper.GetString("mongo.password")
	appConfig.MongoDb = viper.GetString("mongo.db")
	appConfig.MongoGridfsChunkSize = viper.GetInt64("mongo.gridfs_chunk_size")
	appConfig.MongoRangeTimeRandom = viper.GetString("mongo.range_time_random")

	appConfig.EndpointGetJobs = viper.GetString("endpoint.get_jobs")

	appConfig.LogSourceConsumerChanSize = viper.GetInt("channel_config.log_source_consumer_size")
	appConfig.BehaviorConsumerChanSize = viper.GetInt("channel_config.behavior_source_consumer_size")
	appConfig.KafkaInputChanSize = viper.GetInt("channel_config.kafka_input_size")
	appConfig.MessageChanSize = viper.GetInt("channel_config.message_size")
	appConfig.DefaultTimestampField = viper.GetString("event_config.timestamp_field")
	appConfig.DefaultEventIdField = viper.GetString("event_config.event_id")
	appConfig.KafkaNotifyHost = viper.GetString("notify.host")
	appConfig.KafkaNotifyTopic = viper.GetString("notify.topic")
	appConfig.KafkaLogsourceGroupIDPrefix = viper.GetString("kafka_logsource.group_id_prefix")
	appConfig.BatchExpireTickerDuration = viper.GetString("batch_expire_ticker_duration")
	appConfig.KeepOldBatchDuration = viper.GetString("keep_old_batch_duration")
	appConfig.BatchExpireWorkerThroughput = viper.GetInt64("batch_expire_worker_throughput")
	appConfig.ForceServerTimeStamp = viper.GetBool("force_server_timestamp")
	appConfig.KafkaPoolCapSize = viper.GetInt("kafka_logsource.pool_cap_size")
	appConfig.KafkaPoolLenSize = viper.GetInt("kafka_logsource.pool_len_size")
	appConfig.MaxEventPerSaving = viper.GetInt("max_event_per_saving")
	appConfig.WatermarkInterval = viper.GetString("watermark_interval")
	appConfig.NumConsumer = viper.GetInt("num_consumer")
	appConfig.NumRouterWorker = viper.GetInt("num_router_worker")
	appConfig.ConsumerType = viper.GetString("consumer_type")
	GlobalConfig = appConfig

	return appConfig, nil
}

func (config *AppConfig) ParseWatermarkInterval() time.Duration {
	d, err := util.ParseDurationExtended(config.WatermarkInterval)
	if err != nil {
		panic(fmt.Sprintf("invalid batch_expire_ticker_duration: %v", err))
	}
	return d
}

func (config *AppConfig) ParseBatchExpireTickerDuration() time.Duration {
	d, err := util.ParseDurationExtended(config.BatchExpireTickerDuration)
	if err != nil {
		panic(fmt.Sprintf("invalid batch_expire_ticker_duration: %v", err))
	}
	return d
}

func (config *AppConfig) ParseKeepOldBatchDuration() time.Duration {
	d, err := util.ParseDurationExtended(config.KeepOldBatchDuration)
	if err != nil {
		panic(fmt.Sprintf("invalid keep_old_batch_duration: %v", err))
	}
	return d
}

func (config *AppConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort)
}
