package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	authStore "github.com/gabor-boros/go-oauth2-pg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/opcotech/elemo/assets/keys"
	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/repository/pg"

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

type authStoreLogger struct {
	logger log.Logger
}

func (l *authStoreLogger) Log(_ context.Context, level authStore.LogLevel, msg string, args ...any) {
	logArgs := make([]zap.Field, len(args)/2)
	for i, j := 0, 0; i < len(args)-1; i += 2 {
		logArgs[j] = zap.Any(args[i].(string), args[i+1])
		j++
	}

	switch level {
	case authStore.LogLevelDebug:
		l.logger.Debug(msg, logArgs...)
	case authStore.LogLevelInfo:
		l.logger.Info(msg, logArgs...)
	case authStore.LogLevelWarn:
		l.logger.Warn(msg, logArgs...)
	case authStore.LogLevelError:
		l.logger.Error(msg, logArgs...)
	}
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

func parseLicense(licenseConf *config.LicenseConfig) (*license.License, error) {
	if licenseConf == nil {
		return nil, license.ErrNoLicense
	}

	data, err := os.ReadFile(licenseConf.File)
	if err != nil {
		return nil, err
	}

	l, err := license.NewLicense(string(data), keys.PublicKey)
	if err != nil {
		return nil, err
	}

	logger.Info(
		"license parsed",
		zap.String("id", l.ID.String()),
		zap.String("licensee", l.Organization),
		zap.String("expires_at", l.ExpiresAt.String()),
	)

	return l, nil
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
	logger.Info("config file loaded", log.WithPath(viper.ConfigFileUsed()))
}

func initGraphDatabase() (*neo4j.Database, error) {
	driver, err := neo4j.NewDriver(&cfg.GraphDatabase)
	if err != nil {
		return nil, err
	}

	db, err := neo4j.NewDatabase(
		neo4j.WithDriver(driver),
		neo4j.WithDatabaseName(cfg.GraphDatabase.Database),
		neo4j.WithDatabaseLogger(logger.Named("neo4j")),
		neo4j.WithDatabaseTracer(tracer),
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

func initRelationalDatabase() (*pg.Database, pg.Pool, error) {
	pool, err := pg.NewPool(context.Background(), &cfg.RelationalDatabase)
	if err != nil {
		return nil, nil, err
	}

	db, err := pg.NewDatabase(
		pg.WithDatabasePool(pool),
		pg.WithDatabaseLogger(logger.Named("postgres")),
		pg.WithDatabaseTracer(tracer),
	)
	if err != nil {
		return nil, nil, err
	}

	err = db.Ping(context.Background())
	if err != nil {
		return nil, nil, err
	}

	return db, pool, nil
}
