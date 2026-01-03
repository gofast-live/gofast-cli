package cmd

import (
	"os/exec"
	"strings"

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
		showInfraReqs := false
		if cfg, err := config.ParseConfig(); err == nil && cfg.InfraPopulated {
			if err := integrations.ApplyInfraIntegrations(".", cfg.Integrations); err != nil {
				cmd.Printf("Error updating infra integrations: %v\n", err)
				return
			}
			showInfraReqs = true
		}

		cmd.Println("")
		cmd.Println(config.SuccessStyle.Render("Stripe integration added successfully!"))
		cmd.Println("")
		cmd.Println("Next steps:")
		cmd.Printf("  1. Run %s to regenerate proto code\n", config.SuccessStyle.Render("'make gen'"))
		cmd.Printf("  2. Run %s to regenerate SQL queries\n", config.SuccessStyle.Render("'make sql'"))
		cmd.Printf("  3. Run %s to apply migrations\n", config.SuccessStyle.Render("'make migrate'"))
		cmd.Println("  4. Add Stripe environment variables to your .env file:")
		cmd.Println("     - STRIPE_API_KEY")
		cmd.Println("     - STRIPE_WEBHOOK_SECRET")
		cmd.Println("     - STRIPE_PRICE_ID_BASIC")
		cmd.Println("     - STRIPE_PRICE_ID_PRO")
		if showInfraReqs {
			printInfraRequirements(cmd, "stripe")
		}
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
		showInfraReqs := false
		if cfg, err := config.ParseConfig(); err == nil && cfg.InfraPopulated {
			if err := integrations.ApplyInfraIntegrations(".", cfg.Integrations); err != nil {
				cmd.Printf("Error updating infra integrations: %v\n", err)
				return
			}
			showInfraReqs = true
		}

		cmd.Println("")
		cmd.Println(config.SuccessStyle.Render("R2 integration added successfully!"))
		cmd.Println("")
		cmd.Println("Next steps:")
		cmd.Printf("  1. Run %s to regenerate proto code\n", config.SuccessStyle.Render("'make gen'"))
		cmd.Printf("  2. Run %s to regenerate SQL queries\n", config.SuccessStyle.Render("'make sql'"))
		cmd.Printf("  3. Run %s to apply migrations\n", config.SuccessStyle.Render("'make migrate'"))
		cmd.Println("  4. Add R2 environment variables to your .env file:")
		cmd.Println("     - R2_ACCESS_KEY")
		cmd.Println("     - R2_SECRET_KEY")
		cmd.Println("     - R2_ENDPOINT")
		if showInfraReqs {
			printInfraRequirements(cmd, "r2")
		}
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
		showInfraReqs := false
		if cfg, err := config.ParseConfig(); err == nil && cfg.InfraPopulated {
			if err := integrations.ApplyInfraIntegrations(".", cfg.Integrations); err != nil {
				cmd.Printf("Error updating infra integrations: %v\n", err)
				return
			}
			showInfraReqs = true
		}

		cmd.Println("")
		cmd.Println(config.SuccessStyle.Render("Postmark integration added successfully!"))
		cmd.Println("")
		cmd.Println("Next steps:")
		cmd.Printf("  1. Run %s to regenerate proto code\n", config.SuccessStyle.Render("'make gen'"))
		cmd.Printf("  2. Run %s to regenerate SQL queries\n", config.SuccessStyle.Render("'make sql'"))
		cmd.Printf("  3. Run %s to apply migrations\n", config.SuccessStyle.Render("'make migrate'"))
		cmd.Println("  4. Add Postmark environment variables to your .env file:")
		cmd.Println("     - POSTMARK_API_KEY")
		if showInfraReqs {
			printInfraRequirements(cmd, "postmark")
		}
		cmd.Println("")
	},
}

func printInfraRequirements(cmd *cobra.Command, name string) {
	req, ok := integrations.InfraRequirementFor(name)
	if !ok || (len(req.Secrets) == 0 && len(req.Vars) == 0) {
		return
	}

	cmd.Println("")
	cmd.Println("GitHub integration secrets/vars to add:")
	if len(req.Secrets) > 0 {
		cmd.Printf("  %s secrets:\n", strings.ToUpper(req.Name))
		for _, value := range req.Secrets {
			cmd.Printf("    - %s\n", value)
		}
	}
	if len(req.Vars) > 0 {
		cmd.Printf("  %s vars:\n", strings.ToUpper(req.Name))
		for _, value := range req.Vars {
			cmd.Printf("    - %s\n", value)
		}
	}
}
