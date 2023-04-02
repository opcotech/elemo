package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository/neo4j"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

const (
	envPrefix = "ELEMO"
)

var (
	versionInfo *model.VersionInfo

	cfgFile        string
	cfg            *config.Config
	logger         log.Logger
	tracer         trace.Tracer
	tracerProvider trace.TracerProvider
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "elemo",
	Short: "The next-generation project management platform",
	Long:  `Elemo is a project management platform that is designed to be flexible and easy to use.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version, commit, date, goVersion string) {
	versionInfo = &model.VersionInfo{
		Version:   version,
		Commit:    commit,
		Date:      date,
		GoVersion: goVersion,
	}

	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initHook)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.elemo.yml)")
}

func initHook() {
	initConfig()
	initTracer()
	initLogger()
}

func initConfig() {
	var err error

	if cfgFile == "" {
		cfgFile = os.Getenv(envPrefix + "_CONFIG")
	}

	if cfgFile == "" {
		cobra.CheckErr(fmt.Errorf("config file not specified and %s_CONFIG not set", envPrefix))
	}

	viper.SetConfigFile(cfgFile)

	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viper.AutomaticEnv() // read in environment variables that match

	if err = viper.ReadInConfig(); err != nil {
		if ok := errors.As(err, &viper.ConfigFileNotFoundError{}); !ok {
			cobra.CheckErr(err)
		}
	}

	// Bind flags to config value
	cobra.CheckErr(viper.BindPFlags(rootCmd.Flags()))

	if err = viper.Unmarshal(&cfg); err != nil {
		cobra.CheckErr(err)
	}
}

func initTracer() {
	var err error

	tracerProvider, err = tracing.NewTracerProvider(context.Background(), versionInfo, &cfg.Tracing)
	cobra.CheckErr(err)

	tracer = tracerProvider.Tracer("github.com/opcotech/elemo")
}

func initLogger() {
	var err error
	logger, err = log.ConfigureLogger(cfg.Log.Level)
	cobra.CheckErr(err)

	logger.Debug("root logger configured", zap.String("level", cfg.Log.Level))
	logger.Debug("config file loaded", log.WithPath(viper.ConfigFileUsed()))
}

func initDatabase() (*neo4j.Database, error) {
	driver, err := neo4j.NewDriver(&cfg.Database)
	if err != nil {
		return nil, err
	}

	db, err := neo4j.NewDatabase(
		neo4j.WithDriver(driver),
		neo4j.WithDatabaseName(cfg.Database.Name),
	)
	if err != nil {
		return nil, err
	}

	err = db.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return db, nil
}