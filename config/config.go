package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	BuildTime string
	Version   string

	ServerAddress int    `mapstructure:"SERVER_ADDRESS"`
	Environment   string `mapstructure:"ENVIRONMENT"`

	PostgreDSN   string        `mapstructure:"POSTGRE_DSN"`
	MaxOpenConns int           `mapstructure:"POSTGRE_MAX_OPEN_CONNS"`
	MaxIdleConns int           `mapstructure:"POSTGRE_MAX_IDLE_CONNS"`
	MaxIdleTime  time.Duration `mapstructure:"POSTGRE_MAX_IDLE_TIME"`

	LimitedEnable bool    `mapstructure:"LIMITED_ENABLE"`
	LimitedRPS    float64 `mapstructure:"LIMITED_RPS"`
	LimitedBurst  int     `mapstructure:"LIMITED_BURST"`

	SMTPHost string `mapstructure:"SMTP_HOST"`
	SMTPPort int    `mapstructure:"SMTP_PORT"`
	Username string `mapstructure:"SMTP_USERNAME"`
	Password string `mapstructure:"SMTP_PASSWORD"`
	Sender   string `mapstructure:"SMTP_SENDER"`

	TrustedOrigins []string
}

func LoadConfig(path string) (cfg Config) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	cfg.Version = "1.0.0"
	cfg.BuildTime = time.Now().Format(time.RFC850)
	return
}
