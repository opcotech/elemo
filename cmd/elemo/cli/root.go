package cli

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"

	"log/slog"

	"github.com/Shopify/gomail"
	authStore "github.com/gabor-boros/go-oauth2-pg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/assets/keys"
	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	elemoSMTP "github.com/opcotech/elemo/internal/pkg/smtp"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
)

const (
	envPrefix = "ELEMO"
)

var (
	versionInfo *model.VersionInfo

	cfgFile        string
	cfg            *config.Config
	logger         log.Logger
	tracer         tracing.Tracer
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

func (l *authStoreLogger) Log(ctx context.Context, level authStore.LogLevel, msg string, args ...any) {
	logArgs := make([]log.Attr, len(args)/2)
	for i, j := 0, 0; i < len(args)-1; i += 2 {
		logArgs[j] = slog.Any(args[i].(string), args[i+1])
		j++
	}

	switch level {
	case authStore.LogLevelDebug:
		l.logger.Debug(ctx, msg, logArgs...)
	case authStore.LogLevelInfo:
		l.logger.Info(ctx, msg, logArgs...)
	case authStore.LogLevelWarn:
		l.logger.Warn(ctx, msg, logArgs...)
	case authStore.LogLevelError:
		l.logger.Error(ctx, msg, logArgs...)
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

func initTracer(service string) {
	var err error

	tracerProvider, err = tracing.NewTracerProvider(context.Background(), versionInfo, service, &cfg.Tracing)
	cobra.CheckErr(err)

	otelTracer := tracerProvider.Tracer("github.com/opcotech/elemo")
	tracer = tracing.WrapTracer(otelTracer)
}

func initLogger() {
	var err error
	logger, err = log.ConfigureLogger(cfg.Log.Level)
	cobra.CheckErr(err)

	logger.Debug(context.Background(), "root logger configured", slog.String("level", cfg.Log.Level))
	logger.Info(context.Background(), "config file loaded", log.WithPath(viper.ConfigFileUsed()))
}

func initCacheDatabase() (*repository.RedisDatabase, error) {
	client, err := repository.NewRedisClient(&cfg.CacheDatabase)
	if err != nil {
		return nil, err
	}

	db, err := repository.NewRedisDatabase(
		repository.WithRedisClient(client),
		repository.WithRedisDatabaseLogger(logger.Named("redis")),
		repository.WithRedisDatabaseTracer(tracer),
	)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, err
	}

	return db, nil
}

func initGraphDatabase() (*repository.Neo4jDatabase, error) {
	driver, err := repository.NewNeo4jDriver(&cfg.GraphDatabase)
	if err != nil {
		return nil, err
	}

	db, err := repository.NewNeo4jDatabase(
		repository.WithNeo4jDriver(driver),
		repository.WithNeo4jDatabaseName(cfg.GraphDatabase.Database),
		repository.WithNeo4jDatabaseLogger(logger.Named("neo4j")),
		repository.WithNeo4jDatabaseTracer(tracer),
	)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, err
	}

	return db, nil
}

func initRelationalDatabase() (*repository.PGDatabase, repository.PGPool, error) {
	pool, err := repository.NewPool(context.Background(), &cfg.RelationalDatabase)
	if err != nil {
		return nil, nil, err
	}

	db, err := repository.NewPGDatabase(
		repository.WithDatabasePool(pool),
		repository.WithPGDatabaseLogger(logger.Named("postgres")),
		repository.WithPGDatabaseTracer(tracer),
	)
	if err != nil {
		return nil, nil, err
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, nil, err
	}

	return db, pool, nil
}

// initSMTPClient initializes the SMTP client.
func initSMTPClient(smtpConf *config.SMTPConfig) (*elemoSMTP.Client, error) {
	dialer := &gomail.Dialer{Host: smtpConf.Host, Port: smtpConf.Port}
	dialer.Timeout = smtpConf.ConnectionTimeout * time.Second

	if smtpConf.EnableAuth {
		dialer.Auth = smtp.PlainAuth("", smtpConf.Username, smtpConf.Password, smtpConf.Host)
	}

	if smtpConf.SecurityProtocol == "TLS" {
		dialer.TLSConfig = &tls.Config{
			InsecureSkipVerify: smtpConf.SkipTLSVerify, //nolint:gosec
			ServerName:         smtpConf.Host,
		}
	}

	if smtpConf.SecurityProtocol == "STARTTLS" {
		dialer.StartTLSPolicy = gomail.MandatoryStartTLS
	}

	client, err := elemoSMTP.NewClient(
		elemoSMTP.WithWrappedClient(dialer),
		elemoSMTP.WithConfig(smtpConf),
		elemoSMTP.WithLogger(logger.Named("smtp")),
		elemoSMTP.WithTracer(tracer),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
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
		context.Background(),
		"license parsed",
		slog.String("id", l.ID.String()),
		slog.String("licensee", l.Organization),
		slog.String("expires_at", l.ExpiresAt.String()),
	)

	return l, nil
}
