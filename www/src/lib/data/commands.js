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
                text: 'OAuth with PKCE (Google, GitHub, Facebook, Microsoft)',
                details: {
                    files: [
                        'domain/login/oauth.go',
                        'domain/login/service.go',
                        'transport/login/route.go'
                    ],
                    features: [
                        'PKCE flow with S256 challenge',
                        'Provider-specific user info extraction',
                        'Secure state/verifier exchange'
                    ]
                }
            },
            {
                text: 'Ed25519 JWT tokens',
                details: {
                    files: ['pkg/auth/auth.go', 'transport/login/route.go'],
                    features: [
                        'EdDSA signatures (modern, fast)',
                        'Access + refresh token pair',
                        'HttpOnly secure cookies'
                    ]
                }
            },
            {
                text: '2FA with Twilio Verify',
                details: {
                    files: ['domain/login/twilio.go', 'transport/login/route.go'],
                    features: [
                        'SMS verification flow',
                        'Session tokens for 2FA state',
                        'Phone number storage'
                    ]
                }
            },
            {
                text: 'Bitwise permissions (Discord-style)',
                details: {
                    files: ['pkg/auth/auth.go', 'pkg/auth/middleware.go'],
                    features: [
                        'Efficient bitwise AND permission checks',
                        'Per-model CRUD flags auto-generated',
                        'Plan-based access control (Basic, Pro)'
                    ]
                }
            },
            {
                text: 'PostgreSQL + SQLC',
                details: {
                    files: [
                        'storage/sqlc.yaml',
                        'storage/query.sql',
                        'storage/migrations/'
                    ],
                    features: [
                        'Type-safe query generation',
                        'pgx v5 with connection pooling',
                        'Goose migrations with up/down'
                    ]
                }
            },
            {
                text: 'ConnectRPC transport',
                details: {
                    files: ['transport/server.go', 'proto/v1/'],
                    features: [
                        'HTTP/2 with h2c multiplexing',
                        'Unary + streaming RPC support',
                        'JSON & binary encoding'
                    ]
                }
            },
            {
                text: 'Interceptor chain',
                details: {
                    files: ['transport/server.go', 'pkg/auth/middleware.go'],
                    features: [
                        'OpenTelemetry tracing',
                        'Auth validation + token refresh',
                        'Request/response logging'
                    ]
                }
            },
            {
                text: 'Docker Compose setup',
                details: {
                    files: ['docker-compose.yml', 'Dockerfile', 'Makefile'],
                    features: [
                        'service-core (Go backend, port 4000)',
                        'oauth-proxy (callback handler)',
                        'PostgreSQL 18 with health checks'
                    ]
                }
            },
            {
                text: 'GitHub Actions CI/CD',
                details: {
                    files: [
                        '.github/workflows/build.yml',
                        '.github/workflows/deploy.yml',
                        '.github/workflows/pr-deploy.yml'
                    ],
                    features: [
                        'Docker build & push to GHCR',
                        'PR preview environments',
                        'Migration pipeline'
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
                    files: ['storage/migrations/NNNN_create_models.sql'],
                    features: [
                        'Timestamptz for dates',
                        'UUID primary keys',
                        'Foreign key to users'
                    ]
                }
            },
            {
                text: 'SQLC queries',
                details: {
                    files: ['storage/query.sql', 'storage/query/'],
                    features: [
                        'Insert, Get, List, Update, Delete',
                        'User-scoped queries (WHERE user_id)',
                        'Type-safe Go structs'
                    ]
                }
            },
            {
                text: 'Proto + ConnectRPC service',
                details: {
                    files: ['proto/v1/model.proto', 'gen/proto/'],
                    features: [
                        'Message with all column types',
                        '5 RPC methods (CRUD + List)',
                        'Go + TypeScript codegen'
                    ]
                }
            },
            {
                text: 'Domain service layer',
                details: {
                    files: ['domain/model/service.go', 'domain/model/validation.go'],
                    features: [
                        'Business logic + auth checks',
                        'Input validation per field type',
                        'Error mapping to ConnectRPC codes'
                    ]
                }
            },
            {
                text: 'Transport handlers',
                details: {
                    files: ['transport/model/route.go', 'transport/model/route_test.go'],
                    features: [
                        'ConnectRPC unary handlers',
                        'Request/response mapping',
                        'Integration tests included'
                    ]
                }
            },
            {
                text: 'Permission flags',
                details: {
                    files: ['pkg/auth/auth.go'],
                    features: [
                        'Get, Create, Edit, Remove flags',
                        'Auto-wired to UserAccess bitmap',
                        'Bitwise permission checks'
                    ]
                }
            },
            {
                text: 'Generated tests',
                details: {
                    files: ['domain/model/service_test.go', 'transport/model/route_test.go'],
                    features: [
                        'Factory functions (makeQuery, makeInsert)',
                        'Table-driven validation tests',
                        'Integration tests with real DB'
                    ]
                }
            }
        ],
        contextOutputs: [
            {
                text: 'Svelte CRUD pages',
                showIf: (s) => s.has('client'),
                dependency: 'Frontend',
                details: {
                    files: [
                        'routes/(app)/models/+page.svelte',
                        'routes/(app)/models/[id]/+page.svelte'
                    ],
                    features: ['List with pagination', 'Create/Edit forms', 'Delete with confirmation']
                }
            },
            {
                text: 'Subscription checks wired',
                showIf: (s) => s.has('stripe'),
                dependency: 'Stripe',
                details: {
                    files: ['domain/model/service.go'],
                    features: [
                        'Plan-based access control',
                        'Feature gating per tier',
                        'Graceful upgrade prompts'
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
                text: 'SvelteKit 2 + Vite',
                details: {
                    files: ['svelte.config.js', 'vite.config.js', 'src/app.html'],
                    features: ['SSR/CSR hydration', 'File-based routing', 'TypeScript 5.9']
                }
            },
            {
                text: 'Tailwind CSS 4 + DaisyUI',
                details: {
                    files: ['src/app.css', 'tailwind.config.js'],
                    features: [
                        'Utility-first styling',
                        'Pre-built component library',
                        'Dark mode support'
                    ]
                }
            },
            {
                text: 'ConnectRPC client',
                details: {
                    files: ['src/lib/connect.ts', 'src/lib/gen/proto/'],
                    features: [
                        'Type-safe API calls',
                        'Auto-generated from proto',
                        'Credentials included (cookies)'
                    ]
                }
            },
            {
                text: 'Auth interceptor',
                details: {
                    files: ['src/lib/connect.ts', 'src/routes/login/'],
                    features: [
                        'Catches Unauthenticated errors',
                        'Auto-redirect to login',
                        'Stream error handling'
                    ]
                }
            },
            {
                text: 'Cloudflare Workers ready',
                details: {
                    files: ['wrangler.toml', 'svelte.config.js'],
                    features: [
                        'Edge deployment adapter',
                        'Node.js adapter fallback',
                        'PR preview environments'
                    ]
                }
            }
        ],
        contextOutputs: [
            {
                text: (s) => `CRUD pages for: ${s.models.join(', ')}`,
                showIf: (s) => s.models.length > 0,
                dependency: 'Models',
                details: {
                    files: ['src/routes/(app)/[model]/'],
                    features: ['Auto-generated list/detail views', 'Form validation', 'Toast notifications']
                }
            },
            {
                text: 'Billing UI',
                showIf: (s) => s.has('stripe'),
                dependency: 'Stripe',
                details: {
                    files: ['src/routes/(app)/billing/+page.svelte'],
                    features: ['Plan selection', 'Checkout redirect', 'Portal link']
                }
            },
            {
                text: 'File manager',
                showIf: (s) => s.has('r2'),
                dependency: 'R2 Storage',
                details: {
                    files: ['src/routes/(app)/files/+page.svelte'],
                    features: ['Upload with progress', 'File listing', 'Download/delete actions']
                }
            },
            {
                text: 'Email dashboard',
                showIf: (s) => s.has('postmark'),
                dependency: 'Postmark',
                details: {
                    files: ['src/routes/(app)/emails/+page.svelte'],
                    features: ['Send history', 'Compose form', 'Attachment support']
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
                    files: ['domain/payment/service.go', 'transport/payment/route.go'],
                    features: [
                        'Stripe customer creation',
                        'Checkout session generation',
                        'Billing portal links'
                    ]
                }
            },
            {
                text: 'Subscriptions table',
                details: {
                    files: ['storage/migrations/NNNN_create_subscriptions.sql'],
                    features: [
                        'stripe_customer_id + subscription_id',
                        'Status + period tracking',
                        'Cancellation handling'
                    ]
                }
            },
            {
                text: 'Webhook handler',
                details: {
                    files: ['transport/payment/route.go'],
                    features: [
                        'Signature verification',
                        'invoice.paid event handling',
                        'Subscription status sync'
                    ]
                }
            },
            {
                text: 'Plan-based access bits',
                details: {
                    files: ['pkg/auth/auth.go', 'domain/login/service.go'],
                    features: [
                        'BasicPlan + ProPlan flags',
                        'Added to user access bitmap',
                        'Safe period buffer for lapses'
                    ]
                }
            }
        ],
        contextOutputs: [
            {
                text: 'Billing UI components',
                showIf: (s) => s.has('client'),
                dependency: 'Frontend',
                details: {
                    files: ['src/routes/(app)/billing/'],
                    features: ['Plan comparison', 'Upgrade/downgrade flow']
                }
            },
            {
                text: 'Infra secrets configured',
                showIf: (s) => s.has('infra'),
                dependency: 'Infrastructure',
                details: {
                    files: ['infra/integrations.tf', 'infra/secrets.tf'],
                    features: ['K8s secret injection', 'Webhook signing key']
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
                    files: ['domain/file/service.go', 'domain/file/r2.go'],
                    features: [
                        'S3-compatible API client',
                        'Presigned URL generation',
                        'Upload rate limiting'
                    ]
                }
            },
            {
                text: 'Files table',
                details: {
                    files: ['storage/migrations/NNNN_create_files.sql'],
                    features: [
                        'file_key (user_id/file_id format)',
                        'file_name, file_size, content_type',
                        'User ownership tracking'
                    ]
                }
            },
            {
                text: 'File validation',
                details: {
                    files: ['domain/file/service.go'],
                    features: [
                        'Size limit enforcement',
                        'Content-type validation',
                        'Configurable rate limits'
                    ]
                }
            },
            {
                text: 'Permission flags',
                details: {
                    files: ['pkg/auth/auth.go'],
                    features: [
                        'GetFiles, UploadFiles, DownloadFile, RemoveFile',
                        'Per-operation access control'
                    ]
                }
            }
        ],
        contextOutputs: [
            {
                text: 'File manager UI',
                showIf: (s) => s.has('client'),
                dependency: 'Frontend',
                details: {
                    files: ['src/routes/(app)/files/'],
                    features: ['Drag & drop upload', 'Progress indicator', 'File listing']
                }
            },
            {
                text: 'Infra secrets configured',
                showIf: (s) => s.has('infra'),
                dependency: 'Infrastructure',
                details: {
                    files: ['infra/integrations.tf'],
                    features: ['R2 credentials injection', 'Bucket name config']
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
                    files: ['domain/email/service.go', 'domain/email/postmark.go'],
                    features: [
                        'Postmark API integration',
                        'Template email support',
                        'Attachment handling (base64)'
                    ]
                }
            },
            {
                text: 'Email tables',
                details: {
                    files: ['storage/migrations/NNNN_create_emails.sql'],
                    features: [
                        'emails: to, from, subject, body',
                        'email_attachments: file_name, content_type',
                        'Send history tracking'
                    ]
                }
            },
            {
                text: 'Email validation',
                details: {
                    files: ['domain/email/service.go'],
                    features: [
                        'Recipient validation',
                        'Subject + body required',
                        'Attachment size limits'
                    ]
                }
            },
            {
                text: 'Permission flags',
                details: {
                    files: ['pkg/auth/auth.go'],
                    features: ['GetEmails, SendEmail flags', 'Admin-only send capability']
                }
            }
        ],
        contextOutputs: [
            {
                text: 'Email dashboard UI',
                showIf: (s) => s.has('client'),
                dependency: 'Frontend',
                details: {
                    files: ['src/routes/(app)/emails/'],
                    features: ['Compose form', 'Send history', 'Attachment upload']
                }
            },
            {
                text: 'Infra secrets configured',
                showIf: (s) => s.has('infra'),
                dependency: 'Infrastructure',
                details: {
                    files: ['infra/integrations.tf'],
                    features: ['Postmark API key injection', 'From address config']
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
                text: 'Terraform K8s deployment',
                details: {
                    files: ['infra/service-core.tf', 'infra/service-oauth-proxy.tf'],
                    features: [
                        '2 replicas with rolling updates',
                        'Resource limits (250m CPU, 256Mi)',
                        'Liveness + readiness probes'
                    ]
                }
            },
            {
                text: 'CloudNativePG database',
                details: {
                    files: ['infra/cloudnativepg.tf'],
                    features: [
                        'Managed PostgreSQL on K8s',
                        'Automatic backups',
                        'Connection pooling'
                    ]
                }
            },
            {
                text: 'Ingress configuration',
                details: {
                    files: ['infra/ingress.tf', 'infra/variables.tf'],
                    features: [
                        'Domain routing',
                        'TLS termination',
                        'Path-based routing'
                    ]
                }
            },
            {
                text: 'Setup scripts',
                details: {
                    files: [
                        'infra/setup_rke2.sh',
                        'infra/setup_gh.sh',
                        'infra/setup_cloudflare.sh'
                    ],
                    features: [
                        'RKE2 cluster bootstrap',
                        'GitHub secrets/variables sync',
                        'Cloudflare DNS + Workers'
                    ]
                }
            },
            {
                text: 'Cron jobs',
                details: {
                    files: ['infra/cron-delete-tokens.tf'],
                    features: ['Expired token cleanup', 'Authenticated cron endpoint']
                }
            }
        ],
        contextOutputs: [
            {
                text: 'Client Workers deployment',
                showIf: (s) => s.has('client'),
                dependency: 'Frontend',
                details: {
                    files: ['wrangler.toml'],
                    features: ['Cloudflare Workers edge hosting', 'PR preview environments']
                }
            },
            {
                text: 'Stripe secrets',
                showIf: (s) => s.has('stripe'),
                dependency: 'Stripe',
                details: {
                    files: ['infra/integrations.tf'],
                    features: ['API key + webhook secret', 'Price ID configuration']
                }
            },
            {
                text: 'R2 credentials',
                showIf: (s) => s.has('r2'),
                dependency: 'R2 Storage',
                details: {
                    files: ['infra/integrations.tf'],
                    features: ['Access key + secret', 'Endpoint + bucket config']
                }
            },
            {
                text: 'Postmark credentials',
                showIf: (s) => s.has('postmark'),
                dependency: 'Postmark',
                details: {
                    files: ['infra/integrations.tf'],
                    features: ['API key injection', 'From address config']
                }
            },
            {
                text: 'Monitoring Terraform',
                showIf: (s) => s.has('mon'),
                dependency: 'Monitoring',
                details: {
                    files: ['infra/monitoring.tf'],
                    features: [
                        'Alloy, Loki, Tempo, Prometheus',
                        'Grafana ConfigMaps',
                        'Helm chart deployments'
                    ]
                }
            }
        ]
    },
    {
        id: 'mon',
        label: 'monitoring',
        command: () => 'gof mon',
        description: 'Adding monitoring stack...',
        baseOutputs: [
            {
                text: 'Grafana Alloy collector',
                details: {
                    files: ['monitoring/alloy-config.alloy'],
                    features: [
                        'OTLP receivers (gRPC + HTTP)',
                        'Span metrics generation (RED)',
                        'Service graph mapping'
                    ]
                }
            },
            {
                text: 'Tempo distributed tracing',
                details: {
                    files: ['monitoring/tempo.yaml'],
                    features: [
                        'Trace storage + querying',
                        'OTLP + Zipkin receivers',
                        'Trace-to-logs correlation'
                    ]
                }
            },
            {
                text: 'Loki log aggregation',
                details: {
                    files: ['monitoring/loki-config.yaml'],
                    features: [
                        'Structured log collection',
                        'Label-based filtering',
                        'LogQL query language'
                    ]
                }
            },
            {
                text: 'Prometheus metrics',
                details: {
                    files: ['monitoring/prometheus.yml'],
                    features: [
                        'Metrics scraping from Alloy',
                        'Exemplar storage (trace links)',
                        'Remote write receiver'
                    ]
                }
            },
            {
                text: 'Grafana dashboards',
                details: {
                    files: [
                        'monitoring/grafana-datasources.yaml',
                        'monitoring/grafana-dashboards.yaml',
                        'monitoring/dashboards/'
                    ],
                    features: [
                        'Pre-configured datasources',
                        'Service dashboard included',
                        'Anonymous access (dev mode)'
                    ]
                }
            },
            {
                text: 'Docker Compose monitoring',
                details: {
                    files: ['docker-compose.monitoring.yml'],
                    features: [
                        '`make startm` one-command stack',
                        'Grafana at localhost:3001',
                        'All services with health checks'
                    ]
                }
            }
        ],
        contextOutputs: [
            {
                text: 'Kubernetes monitoring',
                showIf: (s) => s.has('infra'),
                dependency: 'Infrastructure',
                details: {
                    files: ['infra/monitoring.tf'],
                    features: [
                        'Helm chart deployments',
                        'ConfigMap injection',
                        'Service discovery'
                    ]
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
