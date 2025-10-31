package cli

import (
	"context"

	"log/slog"

	authStore "github.com/gabor-boros/go-oauth2-pg"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"github.com/opcotech/elemo/internal/model"
)

// authAddClient represents the addClient command
var authAddClient = &cobra.Command{
	Use:   "add-client",
	Short: "Add new OAuth 2.0 client",
	Long: `Add a new OAuth 2.0 client to the database. The client ID and secret will be
generated automatically. The client ID and secret will be printed to the
standard output.

Examples:

# Create a new client
elemo auth add-client --callback-url https://example.com/callback

# Create a new public client for the domain example.com
elemo auth add-client --domain example.com --public`,
	Run: func(cmd *cobra.Command, _ []string) {
		callbackURL, err := cmd.Flags().GetString("callback-url")
		if err != nil {
			logger.Fatal(context.Background(), "failed to get callback-url flag value", slog.Any("error", err))
		}

		public, err := cmd.Flags().GetBool("public")
		if err != nil {
			logger.Fatal(context.Background(), "failed to get public flag value", slog.Any("error", err))
		}

		userID, err := cmd.Flags().GetString("user-id")
		if err != nil {
			logger.Fatal(context.Background(), "failed to get user-id flag value", slog.Any("error", err))
		}

		if callbackURL == "" {
			logger.Fatal(context.Background(), "callback-url is required")
		}

		initTracer("cli-auth-add-client")

		_, relDBPool, err := initRelationalDatabase()
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize relational database", slog.Any("error", err))
		}

		clientStore, err := authStore.NewClientStore(
			authStore.WithClientStoreConnPool(relDBPool.(*pgxpool.Pool)),
			authStore.WithClientStoreTable(authStore.DefaultClientStoreTable),
			authStore.WithClientStoreLogger(&authStoreLogger{
				logger: logger.Named("auth_store"),
			}),
		)
		if err != nil {
			logger.Fatal(context.Background(), "failed to initialize client store", slog.Any("error", err))
		}

		if err := clientStore.InitTable(context.Background()); err != nil {
			logger.Fatal(context.Background(), "failed to initialize client store", slog.Any("error", err))
		}

		client := &models.Client{
			ID:     model.NewRawID(),
			Secret: uuid.New().String(),
			Domain: callbackURL,
			Public: public,
			UserID: userID,
		}

		if err = clientStore.Create(client); err != nil {
			logger.Fatal(context.Background(), "failed to create client", slog.Any("error", err))
		}

		logger.Info(context.Background(), "client created successfully",
			slog.String("callback-url", client.GetDomain()),
			slog.Bool("public", client.IsPublic()),
			slog.String("user-id", client.GetUserID()),
			slog.String("client-id", client.GetID()),
			slog.String("client-secret", client.GetSecret()),
		)
	},
}

func init() {
	authCmd.AddCommand(authAddClient)

	authAddClient.Flags().StringP("user-id", "u", "", "User ID of the client")
	authAddClient.Flags().StringP("callback-url", "c", "", "Callback URL of the client")
	authAddClient.Flags().BoolP("public", "p", false, "Set the client as public")
}
