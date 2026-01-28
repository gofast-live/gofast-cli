/**
 * @typedef {Object} WallSection
 * @property {string} category
 * @property {string[]} items
 * @property {(state: import('../stores/state.svelte.js').State) => boolean} [showIf]
 */

/**
 * @typedef {Object} DynamicItem
 * @property {(state: import('../stores/state.svelte.js').State) => string[]} items
 * @property {(state: import('../stores/state.svelte.js').State) => boolean} [showIf]
 */

/**
 * @typedef {Object} WallSectionConfig
 * @property {string} category
 * @property {string[]} items
 * @property {DynamicItem[]} [dynamicItems]
 * @property {(state: import('../stores/state.svelte.js').State) => boolean} [showIf]
 */

/** @type {WallSectionConfig[]} */
export const wallSections = [
    {
        category: 'Backend (Go)',
        items: [
            'main.go entrypoint',
            'Graceful shutdown',
            'Structured logging (slog)',
            'CORS middleware',
            'Health + readiness endpoints',
            'Config from environment',
            'OpenTelemetry instrumentation'
        ]
    },
    {
        category: 'Auth & Security',
        items: [
            'OAuth with PKCE flow',
            'Ed25519 JWT tokens',
            'Refresh token rotation',
            'Bitwise permissions (RBAC)',
            'HttpOnly secure cookies',
            'API key generation'
        ],
        dynamicItems: [
            {
                showIf: (s) => s.has('stripe'),
                items: () => ['Plan-based access (Basic, Pro)']
            }
        ]
    },
    {
        category: 'Database (PostgreSQL)',
        items: [
            'SQLC type-safe queries',
            'pgx v5 connection pool',
            'Goose migrations',
            'UUID primary keys',
            'User schema + auth tokens'
        ],
        dynamicItems: [
            {
                showIf: (s) => s.models.length > 0,
                items: (s) => s.models.map((m) => `${m} table + queries`)
            },
            {
                showIf: (s) => s.has('stripe'),
                items: () => ['Subscriptions table']
            },
            {
                showIf: (s) => s.has('r2'),
                items: () => ['Files table']
            },
            {
                showIf: (s) => s.has('postmark'),
                items: () => ['Emails + attachments tables']
            }
        ]
    },
    {
        category: 'API (ConnectRPC)',
        items: [
            'Proto definitions (.proto)',
            'Go generated code',
            'HTTP/2 with h2c',
            'Unary + streaming RPC',
            'JSON & binary encoding',
            'Interceptor chain (auth, logging, OTEL)'
        ],
        dynamicItems: [
            {
                showIf: (s) => s.has('client'),
                items: () => ['TypeScript generated code']
            },
            {
                showIf: (s) => s.models.length > 0,
                items: (s) => s.models.map((m) => `${m} service (CRUD)`)
            }
        ]
    },
    {
        category: 'DevOps',
        items: [
            'Dockerfile (multi-stage)',
            'docker-compose.yml',
            'Makefile commands',
            '.github/workflows/build.yml',
            '.github/workflows/deploy.yml',
            '.github/workflows/pr-deploy.yml'
        ],
        dynamicItems: [
            {
                showIf: (s) => s.has('infra'),
                items: () => ['Terraform K8s configs', 'CloudNativePG setup', 'Setup scripts (rke2, gh, cloudflare)']
            },
            {
                showIf: (s) => s.has('mon'),
                items: () => ['docker-compose.monitoring.yml']
            }
        ]
    },
    {
        category: 'Frontend (SvelteKit)',
        showIf: (s) => s.has('client'),
        items: [
            'SvelteKit 2 + Vite',
            'Tailwind CSS 4 + DaisyUI',
            'ConnectRPC client',
            'Auth interceptor',
            'Toast notifications',
            'Protected layout'
        ],
        dynamicItems: [
            {
                showIf: (s) => s.models.length > 0,
                items: (s) => s.models.map((m) => `${m} CRUD pages`)
            },
            {
                showIf: (s) => s.has('stripe'),
                items: () => ['Billing UI']
            },
            {
                showIf: (s) => s.has('r2'),
                items: () => ['File manager']
            },
            {
                showIf: (s) => s.has('postmark'),
                items: () => ['Email dashboard']
            }
        ]
    },
    {
        category: 'Payments (Stripe)',
        showIf: (s) => s.has('stripe'),
        items: [
            'Checkout session creation',
            'Webhook handler (invoice.paid)',
            'Subscription sync logic',
            'Billing portal link',
            'Plan access bits'
        ]
    },
    {
        category: 'File Storage (R2)',
        showIf: (s) => s.has('r2'),
        items: [
            'S3-compatible client',
            'Presigned URLs',
            'Upload rate limiting',
            'File validation',
            'User file tracking'
        ]
    },
    {
        category: 'Email (Postmark)',
        showIf: (s) => s.has('postmark'),
        items: [
            'Postmark API client',
            'Template email support',
            'Attachment handling',
            'Send history tracking'
        ]
    },
    {
        category: 'Monitoring',
        showIf: (s) => s.has('mon'),
        items: [
            'Grafana Alloy (OTLP)',
            'Tempo (traces)',
            'Loki (logs)',
            'Prometheus (metrics)',
            'Grafana dashboards',
            'Span metrics (RED)'
        ]
    },
    {
        category: 'Infrastructure',
        showIf: (s) => s.has('infra'),
        items: [
            'Terraform K8s deployment',
            'CloudNativePG database',
            'Ingress configuration',
            '2 replicas + rolling updates',
            'Liveness + readiness probes',
            'Cron job (token cleanup)'
        ]
    }
];

/**
 * Build wall data based on current state
 * @param {import('../stores/state.svelte.js').State} state
 * @returns {WallSection[]}
 */
export function buildWallData(state) {
    return wallSections
        .filter((section) => !section.showIf || section.showIf(state))
        .map((section) => {
            const items = [...section.items];

            if (section.dynamicItems) {
                for (const dynamic of section.dynamicItems) {
                    if (!dynamic.showIf || dynamic.showIf(state)) {
                        items.push(...dynamic.items(state));
                    }
                }
            }

            return {
                category: section.category,
                items
            };
        });
}
