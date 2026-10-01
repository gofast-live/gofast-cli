package cmd

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/clients"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/integrations"
	"github.com/spf13/cobra"
)

func formatEnabledClients(ctx context.Context) error {
	cfg, err := config.ParseConfig()
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	for _, client := range clients.Enabled(cfg) {
		err := formatClientProject(ctx, client.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add optional features to the project",
		Long:  "Add optional features like Stripe payments, S3 file storage, or Postmark email to an existing GoFast project.",
	}
}

func newAddStripeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stripe",
		Short: "Add Stripe payment integration",
		Long: `Add Stripe payment integration to your GoFast project.

	This command adds:
	- Payment domain service (checkout, portal, webhook handling)
	- Payment transport layer (ConnectRPC handlers)
	- Subscriptions database migration
	- Payment proto definitions
	- Full subscription-based access control in login service

	After running this command:
	1. Run 'make gen' to regenerate proto code
	2. Run 'make sql' to regenerate SQL queries
	3. Run 'make migrate' to create the subscriptions table
	4. Configure Stripe environment variables in your .env file
	`,
		Run: func(cmd *cobra.Command, _ []string) {
			email, apiKey, err := auth.CheckAuthentication()
			if err != nil {
				cmd.Printf("Authentication failed: %v.\n", err)
				return
			}

			// Ensure we are inside a valid gofast project
			_, err = config.ParseConfig()
			if err != nil {
				cmd.Printf("%v\n", err)
				return
			}

			cmd.Println("")
			cmd.Println("Adding Stripe payment integration...")

			err = integrations.StripeAdd(cmd.Context(), email, apiKey)
			if err != nil {
				cmd.Printf("Error adding Stripe: %v\n", err)
				return
			}

			// Format Go code
			gofmtCmd := exec.CommandContext(cmd.Context(), "go", "fmt", "./...")
			gofmtCmd.Dir = "app/service-core"
			output, err := gofmtCmd.CombinedOutput()
			if err != nil {
				cmd.Printf("Warning: go fmt failed: %v\nOutput: %s\n", err, output)
			}

			err = config.AddIntegration("stripe")
			if err != nil {
				cmd.Printf("Error updating config: %v\n", err)
				return
			}
			err = formatEnabledClients(cmd.Context())
			if err != nil {
				cmd.Printf("Error formatting client after Stripe add: %v\n", err)
				return
			}

			cmd.Println("")
			cmd.Println(config.SuccessStyle().Render("Stripe integration added successfully!"))
			cmd.Println("")
			cfg, err := config.ParseConfig()
			if err == nil && clients.HasAny(cfg) {
				cmd.Println("Add this route to your client navigation:")
				cmd.Printf("  %s\n", config.SuccessStyle().Render("/payments"))
				cmd.Println("")
			}
			cmd.Println("Next steps:")
			cmd.Printf("  1. Run %s to regenerate proto code\n", config.SuccessStyle().Render("'make gen'"))
			cmd.Printf("  2. Run %s to regenerate SQL queries\n", config.SuccessStyle().Render("'make sql'"))
			cmd.Printf("  3. Run %s to format generated code\n", config.SuccessStyle().Render("'make format'"))
			cmd.Printf("  4. Run %s to apply migrations\n", config.SuccessStyle().Render("'make migrate'"))
			cmd.Println("  5. Add environment variables to docker-compose.yml:")
			cmd.Println("     - STRIPE_API_KEY")
			cmd.Println("     - STRIPE_WEBHOOK_SECRET")
			cmd.Println("     - STRIPE_PRICE_ID_BASIC")
			cmd.Println("     - STRIPE_PRICE_ID_PRO")
			cmd.Println("  6. Add to GitHub secrets/variables:")
			cmd.Println("     Secrets: STRIPE_API_KEY, STRIPE_WEBHOOK_SECRET")
			cmd.Println("     Variables: STRIPE_PRICE_ID_BASIC, STRIPE_PRICE_ID_PRO")
			cmd.Println("")
		},
	}
}

func newAddS3Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "s3",
		Short: "Add S3 file storage integration",
		Long: `Add S3 file storage integration to your GoFast project.
	Works with any S3-compatible provider (AWS S3, Cloudflare R2, RustFS, MinIO, etc.).

	This command adds:
	- File domain service (upload, download, delete)
	- File transport layer (ConnectRPC handlers)
	- Files database migration
	- File proto definitions
	- Local RustFS (S3-compatible) services in docker-compose.yml

	After running this command:
	1. Run 'make gen' to regenerate proto code
	2. Run 'make sql' to regenerate SQL queries
	3. Run 'make migrate' to create the files table
	4. Local development works as is: 'make start' runs RustFS and the S3_* defaults point at it.
	   Set S3_* to use R2/AWS instead.
	`,
		Run: func(cmd *cobra.Command, _ []string) {
			email, apiKey, err := auth.CheckAuthentication()
			if err != nil {
				cmd.Printf("Authentication failed: %v.\n", err)
				return
			}

			// Ensure we are inside a valid gofast project
			_, err = config.ParseConfig()
			if err != nil {
				cmd.Printf("%v\n", err)
				return
			}

			cmd.Println("")
			cmd.Println("Adding S3 file storage integration...")

			err = integrations.S3Add(cmd.Context(), email, apiKey)
			if err != nil {
				cmd.Printf("Error adding S3: %v\n", err)
				return
			}

			// Format Go code
			gofmtCmd := exec.CommandContext(cmd.Context(), "go", "fmt", "./...")
			gofmtCmd.Dir = "app/service-core"
			output, err := gofmtCmd.CombinedOutput()
			if err != nil {
				cmd.Printf("Warning: go fmt failed: %v\nOutput: %s\n", err, output)
			}

			err = config.AddIntegration("s3")
			if err != nil {
				cmd.Printf("Error updating config: %v\n", err)
				return
			}
			err = formatEnabledClients(cmd.Context())
			if err != nil {
				cmd.Printf("Error formatting client after S3 add: %v\n", err)
				return
			}

			cmd.Println("")
			cmd.Println(config.SuccessStyle().Render("S3 integration added successfully!"))
			cmd.Println("")
			cfg, err := config.ParseConfig()
			if err == nil && clients.HasAny(cfg) {
				cmd.Println("Add this route to your client navigation:")
				cmd.Printf("  %s\n", config.SuccessStyle().Render("/files"))
				cmd.Println("")
			}
			cmd.Println("Next steps:")
			cmd.Printf("  1. Run %s to regenerate proto code\n", config.SuccessStyle().Render("'make gen'"))
			cmd.Printf("  2. Run %s to regenerate SQL queries\n", config.SuccessStyle().Render("'make sql'"))
			cmd.Printf("  3. Run %s to format generated code\n", config.SuccessStyle().Render("'make format'"))
			cmd.Printf("  4. Run %s to apply migrations\n", config.SuccessStyle().Render("'make migrate'"))
			cmd.Println("  5. Local development uses RustFS (started by 'make start', console at http://localhost:9001/rustfs/console/).")
			cmd.Println("     To use R2/AWS instead, set in docker-compose.yml or your shell:")
			cmd.Println("     - S3_ACCESS_KEY_ID")
			cmd.Println("     - S3_SECRET_ACCESS_KEY")
			cmd.Println("     - S3_ENDPOINT")
			cmd.Println("     - BUCKET_NAME")
			cmd.Println("  6. Add to GitHub secrets/variables:")
			cmd.Println("     Secrets: S3_ACCESS_KEY_ID, S3_SECRET_ACCESS_KEY")
			cmd.Println("     Variables: S3_ENDPOINT, BUCKET_NAME")
			cmd.Println("")
		},
	}
}

func newAddPostmarkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "postmark",
		Short: "Add Postmark email integration",
		Long: `Add Postmark email integration to your GoFast project.

	This command adds:
	- Email domain service (send emails with attachments)
	- Email transport layer (ConnectRPC handlers)
	- Emails database migration
	- Email proto definitions

	After running this command:
	1. Run 'make gen' to regenerate proto code
	2. Run 'make sql' to regenerate SQL queries
	3. Run 'make migrate' to create the emails table
	4. Configure Postmark environment variables in your .env file
	`,
		Run: func(cmd *cobra.Command, _ []string) {
			email, apiKey, err := auth.CheckAuthentication()
			if err != nil {
				cmd.Printf("Authentication failed: %v.\n", err)
				return
			}

			// Ensure we are inside a valid gofast project
			_, err = config.ParseConfig()
			if err != nil {
				cmd.Printf("%v\n", err)
				return
			}

			cmd.Println("")
			cmd.Println("Adding Postmark email integration...")

			err = integrations.PostmarkAdd(cmd.Context(), email, apiKey)
			if err != nil {
				cmd.Printf("Error adding Postmark: %v\n", err)
				return
			}

			// Format Go code
			gofmtCmd := exec.CommandContext(cmd.Context(), "go", "fmt", "./...")
			gofmtCmd.Dir = "app/service-core"
			output, err := gofmtCmd.CombinedOutput()
			if err != nil {
				cmd.Printf("Warning: go fmt failed: %v\nOutput: %s\n", err, output)
			}

			err = config.AddIntegration("postmark")
			if err != nil {
				cmd.Printf("Error updating config: %v\n", err)
				return
			}
			err = formatEnabledClients(cmd.Context())
			if err != nil {
				cmd.Printf("Error formatting client after Postmark add: %v\n", err)
				return
			}

			cmd.Println("")
			cmd.Println(config.SuccessStyle().Render("Postmark integration added successfully!"))
			cmd.Println("")
			cfg, err := config.ParseConfig()
			if err == nil && clients.HasAny(cfg) {
				cmd.Println("Add this route to your client navigation:")
				cmd.Printf("  %s\n", config.SuccessStyle().Render("/emails"))
				cmd.Println("")
			}
			cmd.Println("Next steps:")
			cmd.Printf("  1. Run %s to regenerate proto code\n", config.SuccessStyle().Render("'make gen'"))
			cmd.Printf("  2. Run %s to regenerate SQL queries\n", config.SuccessStyle().Render("'make sql'"))
			cmd.Printf("  3. Run %s to format generated code\n", config.SuccessStyle().Render("'make format'"))
			cmd.Printf("  4. Run %s to apply migrations\n", config.SuccessStyle().Render("'make migrate'"))
			cmd.Println("  5. Add environment variables to docker-compose.yml:")
			cmd.Println("     - POSTMARK_API_KEY")
			cmd.Println("     - EMAIL_FROM")
			cmd.Println("  6. Add to GitHub secrets/variables:")
			cmd.Println("     Secrets: POSTMARK_API_KEY")
			cmd.Println("     Variables: EMAIL_FROM")
			cmd.Println("")
		},
	}
}
