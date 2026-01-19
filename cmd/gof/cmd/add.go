package cmd

import (
	"os/exec"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/integrations"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addStripeCmd)
	addCmd.AddCommand(addR2Cmd)
	addCmd.AddCommand(addPostmarkCmd)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add optional features to the project",
	Long:  "Add optional features like Stripe payments, R2 file storage, or Postmark email to an existing GoFast project.",
}

var addStripeCmd = &cobra.Command{
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
	Run: func(cmd *cobra.Command, args []string) {
		email, apiKey, err := auth.CheckAuthentication()
		if err != nil {
			cmd.Printf("Authentication failed: %v.\n", err)
			return
		}

		// Ensure we are inside a valid gofast project
		if _, err := config.ParseConfig(); err != nil {
			cmd.Printf("%v\n", err)
			return
		}

		cmd.Println("")
		cmd.Println("Adding Stripe payment integration...")

		if err := integrations.StripeAdd(email, apiKey); err != nil {
			cmd.Printf("Error adding Stripe: %v\n", err)
			return
		}

		// Format Go code
		gofmtCmd := exec.Command("go", "fmt", "./...")
		gofmtCmd.Dir = "app/service-core"
		if output, err := gofmtCmd.CombinedOutput(); err != nil {
			cmd.Printf("Warning: go fmt failed: %v\nOutput: %s\n", err, output)
		}

		if err := config.AddIntegration("stripe"); err != nil {
			cmd.Printf("Error updating config: %v\n", err)
			return
		}

		cmd.Println("")
		cmd.Println(config.SuccessStyle.Render("Stripe integration added successfully!"))
		cmd.Println("")
		cmd.Println("Next steps:")
		cmd.Printf("  1. Run %s to regenerate proto code\n", config.SuccessStyle.Render("'make gen'"))
		cmd.Printf("  2. Run %s to regenerate SQL queries\n", config.SuccessStyle.Render("'make sql'"))
		cmd.Printf("  3. Run %s to format generated code\n", config.SuccessStyle.Render("'make format'"))
		cmd.Printf("  4. Run %s to apply migrations\n", config.SuccessStyle.Render("'make migrate'"))
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

var addR2Cmd = &cobra.Command{
	Use:   "r2",
	Short: "Add Cloudflare R2 file storage integration",
	Long: `Add Cloudflare R2 file storage integration to your GoFast project.

This command adds:
- File domain service (upload, download, delete)
- File transport layer (ConnectRPC handlers)
- Files database migration
- File proto definitions

After running this command:
1. Run 'make gen' to regenerate proto code
2. Run 'make sql' to regenerate SQL queries
3. Run 'make migrate' to create the files table
4. Configure R2 environment variables in your .env file
`,
	Run: func(cmd *cobra.Command, args []string) {
		email, apiKey, err := auth.CheckAuthentication()
		if err != nil {
			cmd.Printf("Authentication failed: %v.\n", err)
			return
		}

		// Ensure we are inside a valid gofast project
		if _, err := config.ParseConfig(); err != nil {
			cmd.Printf("%v\n", err)
			return
		}

		cmd.Println("")
		cmd.Println("Adding Cloudflare R2 file storage integration...")

		if err := integrations.R2Add(email, apiKey); err != nil {
			cmd.Printf("Error adding R2: %v\n", err)
			return
		}

		// Format Go code
		gofmtCmd := exec.Command("go", "fmt", "./...")
		gofmtCmd.Dir = "app/service-core"
		if output, err := gofmtCmd.CombinedOutput(); err != nil {
			cmd.Printf("Warning: go fmt failed: %v\nOutput: %s\n", err, output)
		}

		if err := config.AddIntegration("r2"); err != nil {
			cmd.Printf("Error updating config: %v\n", err)
			return
		}

		cmd.Println("")
		cmd.Println(config.SuccessStyle.Render("R2 integration added successfully!"))
		cmd.Println("")
		cmd.Println("Next steps:")
		cmd.Printf("  1. Run %s to regenerate proto code\n", config.SuccessStyle.Render("'make gen'"))
		cmd.Printf("  2. Run %s to regenerate SQL queries\n", config.SuccessStyle.Render("'make sql'"))
		cmd.Printf("  3. Run %s to format generated code\n", config.SuccessStyle.Render("'make format'"))
		cmd.Printf("  4. Run %s to apply migrations\n", config.SuccessStyle.Render("'make migrate'"))
		cmd.Println("  5. Add environment variables to docker-compose.yml:")
		cmd.Println("     - R2_ACCESS_KEY_ID")
		cmd.Println("     - R2_SECRET_ACCESS_KEY")
		cmd.Println("     - R2_ENDPOINT")
		cmd.Println("     - BUCKET_NAME")
		cmd.Println("  6. Add to GitHub secrets/variables:")
		cmd.Println("     Secrets: R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY")
		cmd.Println("     Variables: R2_ENDPOINT, BUCKET_NAME")
		cmd.Println("")
	},
}

var addPostmarkCmd = &cobra.Command{
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
	Run: func(cmd *cobra.Command, args []string) {
		email, apiKey, err := auth.CheckAuthentication()
		if err != nil {
			cmd.Printf("Authentication failed: %v.\n", err)
			return
		}

		// Ensure we are inside a valid gofast project
		if _, err := config.ParseConfig(); err != nil {
			cmd.Printf("%v\n", err)
			return
		}

		cmd.Println("")
		cmd.Println("Adding Postmark email integration...")

		if err := integrations.PostmarkAdd(email, apiKey); err != nil {
			cmd.Printf("Error adding Postmark: %v\n", err)
			return
		}

		// Format Go code
		gofmtCmd := exec.Command("go", "fmt", "./...")
		gofmtCmd.Dir = "app/service-core"
		if output, err := gofmtCmd.CombinedOutput(); err != nil {
			cmd.Printf("Warning: go fmt failed: %v\nOutput: %s\n", err, output)
		}

		if err := config.AddIntegration("postmark"); err != nil {
			cmd.Printf("Error updating config: %v\n", err)
			return
		}

		cmd.Println("")
		cmd.Println(config.SuccessStyle.Render("Postmark integration added successfully!"))
		cmd.Println("")
		cmd.Println("Next steps:")
		cmd.Printf("  1. Run %s to regenerate proto code\n", config.SuccessStyle.Render("'make gen'"))
		cmd.Printf("  2. Run %s to regenerate SQL queries\n", config.SuccessStyle.Render("'make sql'"))
		cmd.Printf("  3. Run %s to format generated code\n", config.SuccessStyle.Render("'make format'"))
		cmd.Printf("  4. Run %s to apply migrations\n", config.SuccessStyle.Render("'make migrate'"))
		cmd.Println("  5. Add environment variables to docker-compose.yml:")
		cmd.Println("     - POSTMARK_API_KEY")
		cmd.Println("     - EMAIL_FROM")
		cmd.Println("  6. Add to GitHub secrets/variables:")
		cmd.Println("     Secrets: POSTMARK_API_KEY")
		cmd.Println("     Variables: EMAIL_FROM")
		cmd.Println("")
	},
}

