package config

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/spf13/viper"
	"os"
	"slices"
)

type Config struct {
	DBHost                     string `mapstructure:"DB_HOST"`
	DBPort                     string `mapstructure:"DB_PORT"`
	DBUser                     string `mapstructure:"DB_USER"`
	DBPassword                 string `mapstructure:"DB_PASSWORD"`
	DBName                     string `mapstructure:"DB_NAME"`
	DBSslMode                  string `mapstructure:"DB_SSL_MODE"`
	ServerAddress              string `mapstructure:"SERVER_ADDRESS"`
	BackdoorOtp                string `mapstructure:"BACKDOOR_OTP"`
	GoogleMapsKey              string `mapstructure:"GOOGLE_MAPS_KEY"`
	ServerPort                 string `mapstructure:"SERVER_PORT"`
	ServerHost                 string `mapstructure:"SERVER_HOST"`
	UseRedis                   bool   `mapstructure:"USE_REDIS"`
	RedisUri                   string `mapstructure:"REDIS_URI"`
	JWTSecretKey               string `mapstructure:"JWT_SECRET"`
	JWTAccessExpirationMinutes int    `mapstructure:"JWT_ACCESS_EXPIRATION_MINUTES"`
	JWTRefreshExpirationDays   int    `mapstructure:"JWT_REFRESH_EXPIRATION_DAYS"`
	Mode                       string `mapstructure:"GIN_MODE"`
	AwsAccessKeyId             string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AwsSecretAccessKey         string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	SqsRegion                  string `mapstructure:"SQS_REGION"`
	SqsLocationQueue           string `mapstructure:"SQS_LOCATION_QUEUE"`
	H3MinResolution            int    `mapstructure:"MINIMUM_RESOLUTION"`
	H3MaxResolution            int    `mapstructure:"MAXIMUM_RESOLUTION"`
	H3DefaultResolution        int    `mapstructure:"DEFAULT_RESOLUTION"`
	AwsBucket                  string `mapstructure:"AWS_BUCKET"`
	AwsBucketRegion            string `mapstructure:"AWS_BUCKET_REGION"`
}

func LoadConfig() (Config, error) {
	v := viper.New()

	env := os.Getenv("APP_ENV")
	envsWithEnvVars := []string{"preview", "staging", "prod"}
	if slices.Contains(envsWithEnvVars, env) {
		// Read in environment variables that match
		v.BindEnv("DB_HOST")
		v.BindEnv("DB_PORT")
		v.BindEnv("DB_USER")
		v.BindEnv("DB_PASSWORD")
		v.BindEnv("DB_NAME")
		v.BindEnv("DB_SSL_MODE")
		v.BindEnv("SERVER_ADDRESS")
		v.BindEnv("BACKDOOR_OTP")
		v.BindEnv("GOOGLE_MAPS_KEY")
		v.BindEnv("SERVER_PORT")
		v.BindEnv("SERVER_HOST")
		v.BindEnv("USE_REDIS")
		v.BindEnv("REDIS_URI")
		v.BindEnv("JWT_SECRET")
		v.BindEnv("JWT_ACCESS_EXPIRATION_MINUTES")
		v.BindEnv("JWT_REFRESH_EXPIRATION_DAYS")
		v.BindEnv("GIN_MODE")
		v.BindEnv("AWS_ACCESS_KEY_ID")
		v.BindEnv("AWS_SECRET_ACCESS_KEY")
		v.BindEnv("SQS_REGION")
		v.BindEnv("SQS_LOCATION_QUEUE")
		v.BindEnv("MINIMUM_RESOLUTION")
		v.BindEnv("MAXIMUM_RESOLUTION")
		v.BindEnv("DEFAULT_RESOLUTION")
		v.BindEnv("AWS_BUCKET")
		v.BindEnv("AWS_BUCKET_REGION")
		v.AutomaticEnv()
	} else {
		v.SetDefault("SERVER_PORT", "8080")
		v.SetDefault("MODE", "debug")

		v.SetConfigName("config") // name of config file (without extension)
		v.SetConfigType("env")    // or viper.SetConfigType("yaml") for YAML files
		v.AddConfigPath(".")      // path to look for the config file in

		// Read the config file
		if err := v.ReadInConfig(); err != nil {
			return Config{}, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	return cfg, nil
}

func (config *Config) Validate() error {
	return validation.ValidateStruct(config,
		validation.Field(&config.ServerPort, is.Port),
		validation.Field(&config.ServerHost, validation.Required),

		validation.Field(&config.UseRedis, validation.In(true, false)),
		validation.Field(&config.RedisUri),

		//validation.Field(&config.JWTSecretKey, validation.Required),
		//validation.Field(&config.JWTAccessExpirationMinutes, validation.Required),
		//validation.Field(&config.JWTRefreshExpirationDays, validation.Required),

		validation.Field(&config.Mode, validation.In("debug", "release")),
	)
}
