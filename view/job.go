package view

type (
	ProfileConfig struct {
		ID          string    `json:"id"`
		Name        string    `json:"name"`
		Status      int64     `json:"status"`
		ProfileType string    `json:"profile_type"`
		Entities    []*Object `json:"entity"`
		Attributes  []*Object `json:"attribute"`
		Categories  []*Object `json:"category"`
		ProfileTime string    `json:"profile_time"`
		Threshold   float64   `json:"threshold"`
	}
	Object struct {
		Name      string                 `json:"field_name" example:"server_id"`
		Type      string                 `json:"type" enums:"original,mapping,reference" example:"original"`
		ExtraData map[string]interface{} `json:"extra_data"`
	}
	kafkaConfig struct {
		BootstrapServer string `json:"bootstrap.servers" binding:"required"`
		Topic           string `json:"topic" binding:"required"`
	}
	LogSourceConfig struct {
		Config kafkaConfig `json:"config" binding:"required"`
	}
	JobConfig struct {
		ID                       string         `json:"id"`
		ProfileConfig            *ProfileConfig `json:"profile_config" binding:"required"`
		LogSourceBuilderConfig   *kafkaConfig   `json:"builder_source_config" binding:"required"`
		LogSourcePredictorConfig *kafkaConfig   `json:"predictor_source_config" binding:"required"`
		OutputConfig             *OutputConfig  `json:"output_config" binding:"required"`
	}
	OutputConfig struct {
		ExpirationTime string                 `mapstructure:"expiration_time" json:"expiration_time"`
		Type           string                 `mapstructure:"type" json:"type"`
		Config         map[string]interface{} `mapstructure:"config" json:"config"`
		Workers        int                    `mapstructure:"workers" json:"workers"`
	}
)
