// Command workload-prefill wipes local Elemo stores and seeds a mature-company
// demo world through the internal services.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/opcotech/elemo/internal/pkg/log"
)

func main() {
	if err := parseAndRun(); err != nil {
		fmt.Fprintf(os.Stderr, "workload-prefill: %v\n", err)
		var usage usageError
		if errors.As(err, &usage) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func parseAndRun() error {
	opts, err := parseOptions()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return run(ctx, opts)
}

type usageError struct {
	msg string
}

func (e usageError) Error() string {
	return e.msg
}

func run(ctx context.Context, opts options) error {
	cfg, err := loadConfig(opts.configPath)
	if err != nil {
		return err
	}

	logger, err := log.ConfigureLogger(cfg.Log.Level)
	if err != nil {
		return fmt.Errorf("configure logger: %w", err)
	}

	deps, err := wire(ctx, cfg, logger)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := deps.close(ctx); closeErr != nil {
			logger.Error(ctx, "failed to close dependencies", log.WithError(closeErr))
		}
	}()

	if err := wipe(ctx, deps, opts.queriesDir); err != nil {
		return err
	}

	summary, err := seed(ctx, deps, opts)
	if err != nil {
		return err
	}

	if !opts.skipReindex {
		logger.Info(ctx, "reindexing search")
		if err := reindexSearch(ctx, deps); err != nil {
			return err
		}
	}

	if err := deps.cacheDB.Client().FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("flush redis: %w", err)
	}

	logger.Info(ctx, "prefill complete",
		log.WithEmail(summary.adminEmail),
		log.WithValue(summary),
	)
	fmt.Fprintf(os.Stdout, "Prefill complete (%s profile)\n", opts.profile)
	fmt.Fprintf(os.Stdout, "Login: %s / %s\n", summary.adminEmail, opts.password)
	fmt.Fprintf(os.Stdout, "Organizations: %d  Users: %d  Projects: %d  Issues: %d  Documents: %d\n",
		summary.organizations,
		summary.users,
		summary.projects,
		summary.issues,
		summary.documents,
	)
	return nil
}

func parseOptions() (options, error) {
	opts := options{
		configPath:  os.Getenv("ELEMO_CONFIG"),
		profile:     profileFull,
		concurrency: 8,
		seed:        42,
		password:    defaultPassword,
		queriesDir:  "assets/queries",
	}

	flag.StringVar(&opts.configPath, "config", opts.configPath, "path to Elemo config file (or ELEMO_CONFIG)")
	flag.BoolVar(&opts.yes, "yes", false, "confirm destructive wipe of Neo4j, Meilisearch, Redis, and Postgres tokens")
	flag.StringVar(&opts.profile, "profile", opts.profile, "workload size: full or smoke")
	flag.IntVar(&opts.concurrency, "concurrency", opts.concurrency, "max parallel project issue seeds")
	flag.Int64Var(&opts.seed, "seed", opts.seed, "RNG seed for generated names and issue mixes")
	flag.StringVar(&opts.password, "password", opts.password, "password assigned to every seeded user")
	flag.StringVar(&opts.queriesDir, "queries-dir", opts.queriesDir, "directory containing bootstrap.cypher and bootstrap.sql")
	flag.BoolVar(&opts.skipReindex, "skip-reindex", false, "skip Meilisearch reindex after seed")
	flag.Parse()

	if opts.configPath == "" {
		return options{}, usageError{msg: "config file not specified and ELEMO_CONFIG not set"}
	}
	if !opts.yes {
		return options{}, usageError{msg: "refusing to wipe stores without -yes"}
	}
	if opts.profile != profileFull && opts.profile != profileSmoke {
		return options{}, usageError{msg: fmt.Sprintf("invalid profile %q (want full or smoke)", opts.profile)}
	}
	if opts.concurrency < 1 {
		return options{}, usageError{msg: "concurrency must be at least 1"}
	}
	if len(opts.password) < 8 || len(opts.password) > 64 {
		return options{}, usageError{msg: "password must be between 8 and 64 characters"}
	}
	return opts, nil
}

type options struct {
	configPath  string
	yes         bool
	profile     string
	concurrency int
	seed        int64
	password    string
	queriesDir  string
	skipReindex bool
}

type seedSummary struct {
	adminEmail    string
	organizations int
	users         int
	projects      int
	issues        int
	documents     int
}
