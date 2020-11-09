package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

var envMap = map[string]string{
	port:           defaultPort,
	logLevel:       defaultlogLevel,
	userDatabase:   defaultUserDatabase,
	userCollection: defaultUserCollection,
	roleCollection: defaultRoleCollection,
}

//Config is the general struct for app configuration
type Config struct {
	Port           string       `json:"port"`
	UserDatabase   string       `json:"characterDatabase"`
	UserCollection string       `json:"characterCollection"`
	RoleCollection string       `json:"roleCollection"`
	LogLevel       logrus.Level `json:"log-level"`
}

//Accessor is the interface setup for any configuration accessor
type Accessor interface {
	BindEnv(input ...string) error
	IsSet(key string) bool
	GetString(key string) string
}

//New sets up a new config based on the interface passed
func New(accessor Accessor) (c *Config, err error) {
	error := loadEnvVars(accessor)
	if error != nil {
		return nil, error
	}

	currentLogLevel, err := logrus.ParseLevel(envMap[logLevel])
	if err != nil {
		logrus.Warnf("Cannot load log-level: %v", err)
	}

	config := Config{
		Port:           envMap[port],
		LogLevel:       currentLogLevel,
		UserDatabase:   envMap[userDatabase],
		UserCollection: envMap[userCollection],
		RoleCollection: envMap[roleCollection],
	}
	return &config, nil
}

func loadEnvVars(accessor Accessor) error {
	for envKey := range envMap {
		err := accessor.BindEnv(envKey)
		if err != nil {
			return fmt.Errorf("error loading environment variable %s: %v", envKey, err)
		}

		if accessor.IsSet(envKey) {
			envMap[envKey] = accessor.GetString(envKey)
		}
	}

	return nil
}
