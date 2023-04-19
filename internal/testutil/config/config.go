package config

import (
	"path/filepath"
	"runtime"

	"github.com/opcotech/elemo/internal/config"
	"github.com/spf13/viper"
)

var (
	_, f, _, _ = runtime.Caller(0)
	RootDir    = filepath.Join(filepath.Dir(f), "..", "..", "..")
	Conf       = LoadConfig(filepath.Join(RootDir, "configs", "test", "config.yml"))
)

// LoadConfig loads the configuration for integration tests.
func LoadConfig(file string) *config.Config {
	conf := new(config.Config)

	viper.SetConfigFile(file)

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(conf); err != nil {
		panic(err)
	}

	return conf
}
