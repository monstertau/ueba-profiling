package view

type (
	ProfileConfig struct {
		Name        string                 `json:"name"`
		Status      int64                  `json:"status"`
		ProfileType string                 `json:"profile_type"`
		Entities    []*Object              `json:"entity"`
		Attributes  []*Object              `json:"attribute"`
		Categories  []*Object              `json:"category"`
		ProfileTime string                 `json:"profile_time"`
		Params      map[string]interface{} `json:"params"`
		Threshold   float64                `json:"threshold"`
	}
	Object struct {
		Name      string                 `json:"field_name" example:"server_id"`
		Type      string                 `json:"type" enums:"original,mapping,reference" example:"original"`
		ExtraData map[string]interface{} `json:"extra_data"`
	}
	kafkaConfig struct {
		BootstrapServer string        `json:"bootstrap.servers" binding:"required"`
		Topic           []interface{} `json:"topics" binding:"required"`
		AuthenType      string        `json:"authen_type"`
		Keytab          string        `json:"keytab"`
		Principal       string        `json:"principal"`
	}
	LogSourceConfig struct {
		Config kafkaConfig `json:"config" binding:"required"`
	}
	JobConfig struct {
		ID                       string           `json:"id"`
		ProfileConfig            *ProfileConfig   `json:"profile_config" binding:"required"`
		LogSourceBuilderConfig   *LogSourceConfig `json:"builder_source_config" binding:"required"`
		LogSourcePredictorConfig *LogSourceConfig `json:"predictor_source_config" binding:"required"`
	}
)
