/**
 * @typedef {Object} OutputDetails
 * @property {string[]} [files]
 * @property {string[]} [features]
 */

/**
 * @typedef {Object} Output
 * @property {string | ((state: import('../stores/state.svelte.js').State) => string)} text
 * @property {(state: import('../stores/state.svelte.js').State) => boolean} [showIf]
 * @property {OutputDetails} [details]
 * @property {string} [dependency] - The feature that triggered this output (e.g., "Stripe", "Models")
 */

/**
 * @typedef {Object} ModelVariant
 * @property {string} id
 * @property {string} name
 * @property {string} command
 * @property {string} tagline
 */

/**
 * @typedef {Object} Command
 * @property {string} id
 * @property {string} label
 * @property {() => string} [command]
 * @property {string} description
 * @property {Output[]} baseOutputs
 * @property {Output[]} contextOutputs
 * @property {boolean} [hasSubPicker]
 * @property {ModelVariant[]} [variants]
 */

/** @type {Command[]} */
export const commands = [
    {
        id: 'init',
        label: 'init',
        command: () => 'gof init myproject',
        description: 'Creating project structure...',
        baseOutputs: [
            {
                text: 'OAuth (GitHub + Google)',
                details: {
                    files: [
                        'internal/auth/handler.go',
                        'internal/auth/provider.go',
                        'schema/001_users.sql'
                    ],
                    features: [
                        'OIDC Callback handling',
                        'Secure HTTP-only cookies',
                        'User persistence & session lookups'
                    ]
                }
            },
            {
                text: 'Bitwise role authorization',
                details: {
                    files: ['internal/auth/roles.go', 'internal/middleware/auth.go'],
                    features: [
                        'Efficient bitwise permission checks',
                        'Middleware for role enforcement',
                        'Custom role definitions'
                    ]
                }
            },
            {
                text: 'Docker Compose setup',
                details: {
                    files: ['docker-compose.yml', 'Dockerfile', 'scripts/init-db.sh'],
                    features: [
                        'PostgreSQL service',
                        'Redis service (optional)',
                        'Local development environment'
                    ]
                }
            },
            {
                text: 'GitHub Actions CI/CD',
                details: {
                    files: [
                        '.github/workflows/ci.yml',
                        '.github/workflows/deploy.yml'
                    ],
                    features: [
                        'Automated testing',
                        'Linting & formatting checks',
                        'Build & push Docker images'
                    ]
                }
            },
            {
                text: 'PR preview deployments',
                details: {
                    files: ['.github/workflows/preview.yml'],
                    features: [
                        'Ephemeral environments per PR',
                        'Automatic cleanup',
                        'URL comment on PR'
                    ]
                }
            },
            {
                text: 'PostgreSQL + SQLC',
                details: {
                    files: [
                        'internal/db/models.go',
                        'internal/db/query.sql.go',
                        'db/schema.sql'
                    ],
                    features: [
                        'Type-safe struct generation',
                        'Connection pooling (pgxpool)',
                        'Database migration runner'
                    ]
                }
            },
            {
                text: 'ConnectRPC transport',
                details: {
                    files: ['internal/api/connect.go', 'gen/proto/go/...'],
                    features: [
                        'gRPC-compatible HTTP APIs',
                        'Type-safe handlers',
                        'Automatic JSON encoding support'
                    ]
                }
            }
        ],
        contextOutputs: []
    },
    {
        id: 'model',
        label: 'model',
        hasSubPicker: true,
        variants: [
            {
                id: 'model:task',
                name: 'task',
                command: 'gof model task title:string due:date done:bool',
                tagline: 'mixed types'
            },
            {
                id: 'model:post',
                name: 'post',
                command: 'gof model post title:string body:string views:number',
                tagline: 'strings + number'
            },
            {
                id: 'model:event',
                name: 'event',
                command: 'gof model event name:string start:date end:date',
                tagline: 'multiple dates'
            }
        ],
        description: 'Generating full CRUD stack...',
        baseOutputs: [
            {
                text: 'SQL migration',
                details: {
                    files: ['db/migrations/002_create_model.sql'],
                    features: [
                        'Up/Down migration scripts',
                        'Index creation',
                        'Foreign key constraints'
                    ]
                }
            },
            {
                text: 'SQLC queries',
                details: {
                    files: ['db/queries/model.sql', 'internal/db/model.sql.go'],
                    features: [
                        'CRUD operations (Create, Get, List, Update, Delete)',
                        'Efficient partial updates',
                        'Context-aware execution'
                    ]
                }
            },
            {
                text: 'Proto definitions',
                details: {
                    files: ['proto/v1/model.proto', 'gen/proto/go/v1/model.pb.go'],
                    features: [
                        'Service definition',
                        'Message types',
                        'gRPC service descriptors'
                    ]
                }
            },
            {
                text: 'Domain service layer',
                details: {
                    files: ['internal/service/model.go'],
                    features: [
                        'Business logic encapsulation',
                        'Transaction management',
                        'Error handling & mapping'
                    ]
                }
            },
            {
                text: 'Transport handlers',
                details: {
                    files: ['internal/api/v1/model.go'],
                    features: [
                        'ConnectRPC handler implementation',
                        'Request/Response mapping',
                        'Validation middleware hooks'
                    ]
                }
            },
            {
                text: 'Validation + tests',
                details: {
                    files: ['internal/service/model_test.go', 'internal/validation/model.go'],
                    features: [
                        'Unit tests for logic',
                        'Integration tests for DB',
                        'Input validation rules'
                    ]
                }
            }
        ],
        contextOutputs: [
            {
                text: 'Svelte pages generated',
                showIf: (s) => s.has('client'),
                dependency: 'Frontend',
                details: {
                    files: [
                        'src/routes/app/[model]/+page.svelte',
                        'src/routes/app/[model]/[id]/+page.svelte'
                    ],
                    features: ['List view with pagination', 'Create/Edit forms', 'Delete actions']
                }
            },
            {
                text: 'Subscription checks wired',
                showIf: (s) => s.has('stripe'),
                dependency: 'Stripe',
                details: {
                    files: ['internal/service/policy.go'],
                    features: [
                        'Limit checks based on plan',
                        'Feature gating',
                        'Usage tracking'
                    ]
                }
            }
        ]
    },
    {
        id: 'client',
        label: 'client',
        command: () => 'gof client svelte',
        description: 'Adding Svelte frontend...',
        baseOutputs: [
            {
                text: 'SvelteKit scaffold',
                details: {
                    files: ['svelte.config.js', 'vite.config.js', 'src/app.html'],
                    features: ['SSR/CSR hydration', 'Routing setup', 'Layout system']
                }
            },
            {
                text: 'Auth integration',
                details: {
                    files: ['src/lib/auth.ts', 'src/hooks.server.ts'],
                    features: [
                        'Session cookie handling',
                        'Protected routes middleware',
                        'User state store'
                    ]
                }
            },
            {
                text: 'Type-safe API client',
                details: {
                    files: ['src/lib/api.ts', 'gen/proto/ts/...'],
                    features: [
                        'Auto-generated TypeScript types',
                        'ConnectRPC client instance',
                        'Request hooks'
                    ]
                }
            }
        ],
        contextOutputs: [
            {
                text: (s) => `Generated pages for: ${s.models.join(', ')}`,
                showIf: (s) => s.models.length > 0,
                dependency: 'Data Models',
                details: {
                    files: ['src/routes/app/[model]/...'],
                    features: ['Auto-generated CRUD UIs', 'Form validation', 'Data loading']
                }
            },
            {
                text: 'Stripe billing UI',
                showIf: (s) => s.has('stripe'),
                dependency: 'Stripe',
                details: {
                    files: ['src/routes/app/billing/+page.svelte'],
                    features: ['Subscription management', 'Invoice history', 'Plan switching']
                }
            },
            {
                text: 'File management UI',
                showIf: (s) => s.has('r2'),
                dependency: 'R2 Storage',
                details: {
                    files: ['src/lib/components/FileManager.svelte'],
                    features: ['Drag & drop uploads', 'File preview', 'Progress bars']
                }
            },
            {
                text: 'Email dashboard',
                showIf: (s) => s.has('postmark'),
                dependency: 'Postmark',
                details: {
                    files: ['src/routes/admin/emails/+page.svelte'],
                    features: ['Template preview', 'Send history', 'Bounce logs']
                }
            }
        ]
    },
    {
        id: 'stripe',
        label: 'stripe',
        command: () => 'gof add stripe',
        description: 'Adding Stripe payments...',
        baseOutputs: [
            {
                text: 'Payment domain service',
                details: {
                    files: ['internal/payment/service.go'],
                    features: ['Customer creation', 'Checkout sessions', 'Portal links']
                }
            },
            {
                text: 'Subscriptions migration',
                details: {
                    files: ['db/migrations/003_subscriptions.sql'],
                    features: [
                        'Customer ID mapping',
                        'Subscription status tracking',
                        'Period tracking'
                    ]
                }
            },
            {
                text: 'Webhook handlers',
                details: {
                    files: ['internal/api/webhooks/stripe.go'],
                    features: [
                        'Signature verification',
                        'Event dispatching',
                        'Status synchronization'
                    ]
                }
            },
            {
                text: 'Access control integration',
                details: {
                    files: ['internal/auth/permissions.go'],
                    features: ['Plan-based access gates', 'Trial expiration logic']
                }
            }
        ],
        contextOutputs: [
            {
                text: 'Billing UI components',
                showIf: (s) => s.has('client'),
                dependency: 'Frontend',
                details: {
                    files: ['src/lib/components/PricingTable.svelte'],
                    features: ['Plan comparison', 'Upgrade/Downgrade flows']
                }
            },
            {
                text: 'Infra secrets configured',
                showIf: (s) => s.has('infra'),
                dependency: 'Infrastructure',
                details: {
                    files: ['k8s/secrets.yaml'],
                    features: ['Secure key injection', 'Webhook signing secrets']
                }
            }
        ]
    },
    {
        id: 'r2',
        label: 'cloudflare r2',
        command: () => 'gof add r2',
        description: 'Adding file storage...',
        baseOutputs: [
            {
                text: 'File domain service',
                details: {
                    files: ['internal/files/service.go'],
                    features: ['Presigned URL generation', 'Upload verification', 'Deletions']
                }
            },
            {
                text: 'Files migration',
                details: {
                    files: ['db/migrations/004_files.sql'],
                    features: ['Metadata storage', 'Owner mapping', 'Mime types']
                }
            },
            {
                text: 'S3-compatible uploads',
                details: {
                    files: ['internal/files/s3.go'],
                    features: ['R2/S3 client setup', 'Bucket policy helpers']
                }
            }
        ],
        contextOutputs: [
            {
                text: 'File manager UI',
                showIf: (s) => s.has('client'),
                dependency: 'Frontend',
                details: {
                    files: ['src/lib/components/FileUploader.svelte'],
                    features: ['Client-side validation', 'Direct-to-cloud upload']
                }
            },
            {
                text: 'Infra secrets configured',
                showIf: (s) => s.has('infra'),
                dependency: 'Infrastructure',
                details: {
                    files: ['terraform/r2.tf'],
                    features: ['Bucket creation', 'CORS policy']
                }
            }
        ]
    },
    {
        id: 'postmark',
        label: 'postmark',
        command: () => 'gof add postmark',
        description: 'Adding email service...',
        baseOutputs: [
            {
                text: 'Email domain service',
                details: {
                    files: ['internal/email/service.go'],
                    features: ['Sender signature verification', 'Delivery tracking']
                }
            },
            {
                text: 'Emails migration',
                details: {
                    files: ['db/migrations/005_emails.sql'],
                    features: ['Outbox pattern support', 'Log storage']
                }
            },
            {
                text: 'Template support',
                details: {
                    files: ['internal/email/templates/'],
                    features: ['HTML/Text variations', 'Variable substitution']
                }
            }
        ],
        contextOutputs: [
            {
                text: 'Email dashboard UI',
                showIf: (s) => s.has('client'),
                dependency: 'Frontend',
                details: {
                    files: ['src/routes/admin/emails/logs/+page.svelte'],
                    features: ['Search & filter', 'Resend capabilities']
                }
            },
            {
                text: 'Infra secrets configured',
                showIf: (s) => s.has('infra'),
                dependency: 'Infrastructure',
                details: {
                    files: ['k8s/configmap.yaml'],
                    features: ['API token management', 'Sender signature ID']
                }
            }
        ]
    },
    {
        id: 'infra',
        label: 'infra',
        command: () => 'gof infra',
        description: 'Adding production infrastructure...',
        baseOutputs: [
            {
                text: 'Kubernetes manifests',
                details: {
                    files: ['k8s/deployment.yaml', 'k8s/service.yaml', 'k8s/ingress.yaml'],
                    features: ['Replica sets', 'Rolling updates', 'Health probes']
                }
            },
            {
                text: 'Terraform configs',
                details: {
                    files: ['terraform/main.tf', 'terraform/variables.tf'],
                    features: ['Managed DB provisioning', 'VPC setup', 'IAM roles']
                }
            },
            {
                text: 'OpenTelemetry setup',
                details: {
                    files: ['internal/telemetry/tracing.go'],
                    features: ['Distributed tracing', 'Metrics export', 'Log correlation']
                }
            },
            {
                text: 'GitHub Actions deploy',
                details: {
                    files: ['.github/workflows/deploy-prod.yml'],
                    features: ['Manual approval gate', 'K8s rollout status', 'Slack notification']
                }
            }
        ],
        contextOutputs: [
            {
                text: 'Cloudflare Workers (client)',
                showIf: (s) => s.has('client'),
                dependency: 'Frontend',
                details: {
                    files: ['wrangler.toml'],
                    features: ['Edge hosting configuration', 'Route binding']
                }
            },
            {
                text: 'Stripe secrets configured',
                showIf: (s) => s.has('stripe'),
                dependency: 'Stripe',
                details: {
                    files: ['k8s/secrets.yaml'],
                    features: ['Secure key injection', 'Webhook signing secrets']
                }
            },
            {
                text: 'R2 bucket configured',
                showIf: (s) => s.has('r2'),
                dependency: 'R2 Storage',
                details: {
                    files: ['terraform/r2.tf'],
                    features: ['Bucket creation', 'CORS policy']
                }
            },
            {
                text: 'Postmark configured',
                showIf: (s) => s.has('postmark'),
                dependency: 'Postmark',
                details: {
                    files: ['k8s/configmap.yaml'],
                    features: ['API token management', 'Sender signature ID']
                }
            }
        ]
    }
];

/**
 * Get command by id
 * @param {string} id
 * @returns {Command | undefined}
 */
export function getCommand(id) {
    return commands.find((c) => c.id === id);
}
