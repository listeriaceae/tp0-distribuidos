package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/common"
)

// InitConfig Function that uses viper library to parse configuration parameters.
// Viper is configured to read variables from both environment variables and the
// config file ./config.yaml. Environment variables takes precedence over parameters
// defined in the configuration file. If some of the variables cannot be parsed,
// an error is returned
func InitConfig() (*viper.Viper, error) {
	v := viper.New()

	// Configure viper to read env variables with the CLI_ prefix
	v.AutomaticEnv()
	v.SetEnvPrefix("cli")
	// Use a replacer to replace env variables underscores with points. This let us
	// use nested configurations in the config file and at the same time define
	// env variables for the nested configurations
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	env_vars := map[string]string{
		"first_name": "NOMBRE",
		"last_name":  "APELLIDO",
		"document":   "DOCUMENTO",
		"birthdate":  "NACIMIENTO",
		"number":     "NUMERO",
	}

	for key, replace := range env_vars {
		v.BindEnv(key, replace)
	}

	// Try to read configuration from config file. If config file
	// does not exists then ReadInConfig will fail but configuration
	// can be loaded from the environment variables so we shouldn't
	// return an error in that case
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("client")
	if err := v.ReadInConfig(); err != nil {
		fmt.Println("Configuration could not be read from config file. Using env variables instead")
	}

	for key, env := range env_vars {
		if !v.IsSet(key) {
			return nil, fmt.Errorf("Missing environment variable: %s", env)
		}
	}

	if _, err := strconv.Atoi(v.GetString("agency")); err != nil {
		return nil, errors.Wrap(err, "Could not parse CLI_AGENCY env var as int.")
	}

	if _, err := strconv.Atoi(v.GetString("number")); err != nil {
		return nil, errors.Wrap(err, "Could not parse NUMERO env var as int.")
	}

	return v, nil
}

// InitLogger Receives the log level to be set in logrus as a string. This method
// parses the string and set the level to the logger. If the level string is not
// valid an error is returned
func InitLogger(logLevel string) error {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return err
	}

	customFormatter := &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   false,
	}
	logrus.SetFormatter(customFormatter)
	logrus.SetLevel(level)
	return nil
}

// PrintConfig Print all the configuration parameters of the program.
// For debugging purposes only
func PrintConfig(v *viper.Viper) {
	logrus.Debugf("action: config | result: success | agency: %d | server_address: %s | log_level: %s | first_name: %s | last_name: %s | document: %s | birthdate: %s | number: %d",
		v.GetInt("agency"),
		v.GetString("server.address"),
		v.GetString("log.level"),
		v.GetString("first_name"),
		v.GetString("last_name"),
		v.GetString("document"),
		v.GetString("birthdate"),
		v.GetInt("number"),
	)
}

func main() {
	v, err := InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err := InitLogger(v.GetString("log.level")); err != nil {
		log.Fatal(err)
	}

	// Print program config with debugging purposes
	PrintConfig(v)

	clientConfig := common.ClientConfig{
		Agency:        v.GetInt("agency"),
		ServerAddress: v.GetString("server.address"),
	}

	bet := common.Bet{
		FirstName: v.GetString("first_name"),
		LastName:  v.GetString("last_name"),
		Document:  v.GetString("document"),
		Birthdate: v.GetString("birthdate"),
		Number:    v.GetInt("number"),
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	client := common.NewClient(clientConfig)
	client.Start(sig, bet)
}
