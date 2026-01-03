package integrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type infraInsert struct {
	path   string
	anchor string
	check  string
	block  string
}

type infraIntegration struct {
	name    string
	inserts []infraInsert
}

type InfraRequirement struct {
	Name    string
	Secrets []string
	Vars    []string
}

func ApplyInfraIntegrations(projectRoot string, enabled []string) error {
	integrations := infraIntegrations()
	enabledSet := make(map[string]bool, len(enabled))
	for _, name := range enabled {
		enabledSet[name] = true
	}

	for _, integ := range integrations {
		if !enabledSet[integ.name] {
			continue
		}
		for _, ins := range integ.inserts {
			path := filepath.Join(projectRoot, ins.path)
			if _, err := os.Stat(path); err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return fmt.Errorf("stat %s: %w", path, err)
			}
			if err := insertAtAnchor(path, ins.anchor, ins.check, ins.block); err != nil {
				return err
			}
		}
	}

	return nil
}

func InfraRequirements(enabled []string) []InfraRequirement {
	all := infraRequirementMap()
	var out []InfraRequirement
	for _, name := range enabled {
		if req, ok := all[name]; ok {
			out = append(out, req)
		}
	}
	return out
}

func InfraRequirementFor(name string) (InfraRequirement, bool) {
	req, ok := infraRequirementMap()[name]
	return req, ok
}

func insertAtAnchor(path, anchor, check, block string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	s := string(content)
	if check != "" && strings.Contains(s, check) {
		return nil
	}
	if !strings.Contains(s, anchor) {
		return fmt.Errorf("anchor %q not found in %s", anchor, path)
	}
	if !strings.HasSuffix(block, "\n") {
		block += "\n"
	}
	updated := strings.Replace(s, anchor, block+anchor, 1)

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(updated), info.Mode()); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func infraIntegrations() []infraIntegration {
	return []infraIntegration{
		{
			name: "stripe",
			inserts: []infraInsert{
				{
					path:   "infra/variables.tf",
					anchor: "# GF_INFRA_VARS_INSERT",
					check:  "STRIPE_API_KEY",
					block: `variable "STRIPE_API_KEY" {
  description = "Stripe API key."
  type        = string
  sensitive   = true
}

variable "STRIPE_WEBHOOK_SECRET" {
  description = "Stripe webhook secret."
  type        = string
  sensitive   = true
}

variable "STRIPE_PRICE_ID_BASIC" {
  description = "Stripe basic price ID."
  type        = string
}

variable "STRIPE_PRICE_ID_PRO" {
  description = "Stripe pro price ID."
  type        = string
}
`,
				},
				{
					path:   "infra/secrets.tf",
					anchor: "# GF_INFRA_SECRETS_INSERT",
					check:  "stripe-secrets",
					block: `resource "kubernetes_secret" "stripe_secrets" {
  metadata {
    name = "stripe-secrets"
  }

  data = {
    api_key        = var.STRIPE_API_KEY
    webhook_secret = var.STRIPE_WEBHOOK_SECRET
  }

  type = "Opaque"
}
`,
				},
				{
					path:   "infra/service-core.tf",
					anchor: "# GF_INFRA_ENV_INSERT",
					check:  "STRIPE_API_KEY",
					block: `          env {
            name = "STRIPE_API_KEY"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.stripe_secrets.metadata[0].name
                key  = "api_key"
              }
            }
          }
          env {
            name = "STRIPE_WEBHOOK_SECRET"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.stripe_secrets.metadata[0].name
                key  = "webhook_secret"
              }
            }
          }
          env {
            name  = "STRIPE_PRICE_ID_BASIC"
            value = var.STRIPE_PRICE_ID_BASIC
          }
          env {
            name  = "STRIPE_PRICE_ID_PRO"
            value = var.STRIPE_PRICE_ID_PRO
          }
`,
				},
				{
					path:   ".github/workflows/terraform.yml",
					anchor: "# GF_INFRA_TF_ENV_INSERT",
					check:  "TF_VAR_STRIPE_API_KEY",
					block: `      TF_VAR_STRIPE_API_KEY: ${{ secrets.STRIPE_API_KEY }}
      TF_VAR_STRIPE_WEBHOOK_SECRET: ${{ secrets.STRIPE_WEBHOOK_SECRET }}
      TF_VAR_STRIPE_PRICE_ID_BASIC: ${{ vars.STRIPE_PRICE_ID_BASIC }}
      TF_VAR_STRIPE_PRICE_ID_PRO: ${{ vars.STRIPE_PRICE_ID_PRO }}
`,
				},
				{
					path:   ".github/workflows/pr-deploy.yml",
					anchor: "# GF_INFRA_PR_EXPORTS_INSERT",
					check:  "STRIPE_API_KEY",
					block: `          export STRIPE_API_KEY='${{ secrets.STRIPE_API_KEY }}'
          export STRIPE_WEBHOOK_SECRET='${{ secrets.STRIPE_WEBHOOK_SECRET }}'
          export STRIPE_PRICE_ID_BASIC='${{ vars.STRIPE_PRICE_ID_BASIC }}'
          export STRIPE_PRICE_ID_PRO='${{ vars.STRIPE_PRICE_ID_PRO }}'
`,
				},
				{
					path:   "infra/pr-environment/secrets.yaml",
					anchor: "# GF_INFRA_PR_SECRETS_INSERT",
					check:  "stripe-secrets",
					block: `---
apiVersion: v1
kind: Secret
metadata:
  name: stripe-secrets
  namespace: pr-${PR_NUMBER}
type: Opaque
stringData:
  api_key: ${STRIPE_API_KEY}
  webhook_secret: ${STRIPE_WEBHOOK_SECRET}
`,
				},
				{
					path:   "infra/pr-environment/service-core.yaml",
					anchor: "# GF_INFRA_PR_ENV_INSERT",
					check:  "STRIPE_API_KEY",
					block: `            - name: STRIPE_API_KEY
              valueFrom:
                secretKeyRef:
                  name: stripe-secrets
                  key: api_key
            - name: STRIPE_WEBHOOK_SECRET
              valueFrom:
                secretKeyRef:
                  name: stripe-secrets
                  key: webhook_secret
            - name: STRIPE_PRICE_ID_BASIC
              value: ${STRIPE_PRICE_ID_BASIC}
            - name: STRIPE_PRICE_ID_PRO
              value: ${STRIPE_PRICE_ID_PRO}
`,
				},
			},
		},
		{
			name: "r2",
			inserts: []infraInsert{
				{
					path:   "infra/variables.tf",
					anchor: "# GF_INFRA_VARS_INSERT",
					check:  "R2_ACCESS_KEY",
					block: `variable "R2_ACCESS_KEY" {
  description = "Cloudflare R2 access key."
  type        = string
  sensitive   = true
}

variable "R2_SECRET_KEY" {
  description = "Cloudflare R2 secret key."
  type        = string
  sensitive   = true
}

variable "R2_ENDPOINT" {
  description = "Cloudflare R2 endpoint."
  type        = string
}

variable "BUCKET_NAME" {
  description = "R2 bucket name."
  type        = string
}
`,
				},
				{
					path:   "infra/secrets.tf",
					anchor: "# GF_INFRA_SECRETS_INSERT",
					check:  "r2-secrets",
					block: `resource "kubernetes_secret" "r2_secrets" {
  metadata {
    name = "r2-secrets"
  }

  data = {
    access_key = var.R2_ACCESS_KEY
    secret_key = var.R2_SECRET_KEY
  }

  type = "Opaque"
}
`,
				},
				{
					path:   "infra/service-core.tf",
					anchor: "# GF_INFRA_ENV_INSERT",
					check:  "R2_ACCESS_KEY",
					block: `          env {
            name = "R2_ACCESS_KEY"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.r2_secrets.metadata[0].name
                key  = "access_key"
              }
            }
          }
          env {
            name = "R2_SECRET_KEY"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.r2_secrets.metadata[0].name
                key  = "secret_key"
              }
            }
          }
          env {
            name  = "R2_ENDPOINT"
            value = var.R2_ENDPOINT
          }
          env {
            name  = "BUCKET_NAME"
            value = var.BUCKET_NAME
          }
`,
				},
				{
					path:   ".github/workflows/terraform.yml",
					anchor: "# GF_INFRA_TF_ENV_INSERT",
					check:  "TF_VAR_R2_ACCESS_KEY",
					block: `      TF_VAR_R2_ACCESS_KEY: ${{ secrets.R2_ACCESS_KEY }}
      TF_VAR_R2_SECRET_KEY: ${{ secrets.R2_SECRET_KEY }}
      TF_VAR_R2_ENDPOINT: ${{ vars.R2_ENDPOINT }}
      TF_VAR_BUCKET_NAME: ${{ vars.BUCKET_NAME }}
`,
				},
				{
					path:   ".github/workflows/pr-deploy.yml",
					anchor: "# GF_INFRA_PR_EXPORTS_INSERT",
					check:  "R2_ACCESS_KEY",
					block: `          export R2_ACCESS_KEY='${{ secrets.R2_ACCESS_KEY }}'
          export R2_SECRET_KEY='${{ secrets.R2_SECRET_KEY }}'
          export R2_ENDPOINT='${{ vars.R2_ENDPOINT }}'
          export BUCKET_NAME='${{ vars.BUCKET_NAME }}'
`,
				},
				{
					path:   "infra/pr-environment/secrets.yaml",
					anchor: "# GF_INFRA_PR_SECRETS_INSERT",
					check:  "r2-secrets",
					block: `---
apiVersion: v1
kind: Secret
metadata:
  name: r2-secrets
  namespace: pr-${PR_NUMBER}
type: Opaque
stringData:
  access_key: ${R2_ACCESS_KEY}
  secret_key: ${R2_SECRET_KEY}
`,
				},
				{
					path:   "infra/pr-environment/service-core.yaml",
					anchor: "# GF_INFRA_PR_ENV_INSERT",
					check:  "R2_ACCESS_KEY",
					block: `            - name: R2_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: r2-secrets
                  key: access_key
            - name: R2_SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: r2-secrets
                  key: secret_key
            - name: R2_ENDPOINT
              value: ${R2_ENDPOINT}
            - name: BUCKET_NAME
              value: ${BUCKET_NAME}
`,
				},
			},
		},
		{
			name: "postmark",
			inserts: []infraInsert{
				{
					path:   "infra/variables.tf",
					anchor: "# GF_INFRA_VARS_INSERT",
					check:  "POSTMARK_API_KEY",
					block: `variable "POSTMARK_API_KEY" {
  description = "Postmark API key."
  type        = string
  sensitive   = true
}

variable "EMAIL_FROM" {
  description = "Default from address."
  type        = string
}
`,
				},
				{
					path:   "infra/secrets.tf",
					anchor: "# GF_INFRA_SECRETS_INSERT",
					check:  "postmark-secrets",
					block: `resource "kubernetes_secret" "postmark_secrets" {
  metadata {
    name = "postmark-secrets"
  }

  data = {
    api_key = var.POSTMARK_API_KEY
  }

  type = "Opaque"
}
`,
				},
				{
					path:   "infra/service-core.tf",
					anchor: "# GF_INFRA_ENV_INSERT",
					check:  "POSTMARK_API_KEY",
					block: `          env {
            name = "POSTMARK_API_KEY"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.postmark_secrets.metadata[0].name
                key  = "api_key"
              }
            }
          }
          env {
            name  = "EMAIL_FROM"
            value = var.EMAIL_FROM
          }
`,
				},
				{
					path:   ".github/workflows/terraform.yml",
					anchor: "# GF_INFRA_TF_ENV_INSERT",
					check:  "TF_VAR_POSTMARK_API_KEY",
					block: `      TF_VAR_POSTMARK_API_KEY: ${{ secrets.POSTMARK_API_KEY }}
      TF_VAR_EMAIL_FROM: ${{ vars.EMAIL_FROM }}
`,
				},
				{
					path:   ".github/workflows/pr-deploy.yml",
					anchor: "# GF_INFRA_PR_EXPORTS_INSERT",
					check:  "POSTMARK_API_KEY",
					block: `          export POSTMARK_API_KEY='${{ secrets.POSTMARK_API_KEY }}'
          export EMAIL_FROM='${{ vars.EMAIL_FROM }}'
`,
				},
				{
					path:   "infra/pr-environment/secrets.yaml",
					anchor: "# GF_INFRA_PR_SECRETS_INSERT",
					check:  "postmark-secrets",
					block: `---
apiVersion: v1
kind: Secret
metadata:
  name: postmark-secrets
  namespace: pr-${PR_NUMBER}
type: Opaque
stringData:
  api_key: ${POSTMARK_API_KEY}
`,
				},
				{
					path:   "infra/pr-environment/service-core.yaml",
					anchor: "# GF_INFRA_PR_ENV_INSERT",
					check:  "POSTMARK_API_KEY",
					block: `            - name: POSTMARK_API_KEY
              valueFrom:
                secretKeyRef:
                  name: postmark-secrets
                  key: api_key
            - name: EMAIL_FROM
              value: ${EMAIL_FROM}
`,
				},
			},
		},
	}
}

func infraRequirementMap() map[string]InfraRequirement {
	return map[string]InfraRequirement{
		"stripe": {
			Name:    "stripe",
			Secrets: []string{"STRIPE_API_KEY", "STRIPE_WEBHOOK_SECRET"},
			Vars:    []string{"STRIPE_PRICE_ID_BASIC", "STRIPE_PRICE_ID_PRO"},
		},
		"r2": {
			Name:    "r2",
			Secrets: []string{"R2_ACCESS_KEY", "R2_SECRET_KEY"},
			Vars:    []string{"R2_ENDPOINT", "BUCKET_NAME"},
		},
		"postmark": {
			Name:    "postmark",
			Secrets: []string{"POSTMARK_API_KEY"},
			Vars:    []string{"EMAIL_FROM"},
		},
	}
}
