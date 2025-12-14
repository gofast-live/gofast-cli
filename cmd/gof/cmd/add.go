package cmd

import (
	"os/exec"

	"github.com/gofast-live/gofast-cli/v2/cmd/gof/auth"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/config"
	"github.com/gofast-live/gofast-cli/v2/cmd/gof/stripe"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addStripeCmd)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add optional features to the project",
	Long:  "Add optional features like Stripe payments to an existing GoFast project.",
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
1. Run 'scripts/run_proto.sh' to regenerate proto code
2. Run 'scripts/run_queries.sh' to regenerate SQL queries
3. Run 'scripts/run_migrations.sh' to create the subscriptions table
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

		if err := stripe.Add(email, apiKey); err != nil {
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
		cmd.Printf("  1. Run %s to regenerate proto code\n", config.SuccessStyle.Render("'scripts/run_proto.sh'"))
		cmd.Printf("  2. Run %s to regenerate SQL queries\n", config.SuccessStyle.Render("'scripts/run_queries.sh'"))
		cmd.Printf("  3. Run %s to apply migrations\n", config.SuccessStyle.Render("'scripts/run_migrations.sh'"))
		cmd.Println("  4. Add Stripe environment variables to your .env file:")
		cmd.Println("     - STRIPE_API_KEY")
		cmd.Println("     - STRIPE_WEBHOOK_SECRET")
		cmd.Println("     - STRIPE_PRICE_ID_BASIC")
		cmd.Println("     - STRIPE_PRICE_ID_PRO")
		cmd.Println("")
	},
}
