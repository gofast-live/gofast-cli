/**
 * @typedef {Object} WallItem
 * @property {string} text
 * @property {string} [source] - Command that added this item (e.g., 'stripe', 'model')
 */

/**
 * @typedef {Object} WallSection
 * @property {string} category
 * @property {WallItem[]} items
 * @property {(state: import('../stores/state.svelte.js').State) => boolean} [showIf]
 */

/**
 * @typedef {Object} DynamicItem
 * @property {(state: import('../stores/state.svelte.js').State) => WallItem[]} items
 * @property {(state: import('../stores/state.svelte.js').State) => boolean} [showIf]
 */

/**
 * @typedef {Object} WallSectionConfig
 * @property {string} category
 * @property {(string | WallItem)[]} items
 * @property {DynamicItem[]} [dynamicItems]
 * @property {(state: import('../stores/state.svelte.js').State) => boolean} [showIf]
 * @property {string} [source] - Default source for all items in this section
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
                items: () => [{ text: 'Plan-based access (Basic, Pro)', source: 'stripe' }]
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
                items: (s) => s.models.map((m) => ({ text: `${m} table + queries`, source: 'model' }))
            },
            {
                showIf: (s) => s.has('stripe'),
                items: () => [{ text: 'Subscriptions table', source: 'stripe' }]
            },
            {
                showIf: (s) => s.has('r2'),
                items: () => [{ text: 'Files table', source: 'r2' }]
            },
            {
                showIf: (s) => s.has('postmark'),
                items: () => [{ text: 'Emails + attachments tables', source: 'postmark' }]
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
                items: () => [{ text: 'TypeScript generated code', source: 'client' }]
            },
            {
                showIf: (s) => s.models.length > 0,
                items: (s) => s.models.map((m) => ({ text: `${m} service (CRUD)`, source: 'model' }))
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
                items: () => [
                    { text: 'Terraform K8s configs', source: 'infra' },
                    { text: 'CloudNativePG setup', source: 'infra' },
                    { text: 'Setup scripts (rke2, gh, cloudflare)', source: 'infra' }
                ]
            },
            {
                showIf: (s) => s.has('mon'),
                items: () => [{ text: 'docker-compose.monitoring.yml', source: 'mon' }]
            }
        ]
    },
    {
        category: 'Frontend (SvelteKit)',
        showIf: (s) => s.has('client'),
        source: 'client',
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
                items: (s) => s.models.map((m) => ({ text: `${m} CRUD pages`, source: 'model' }))
            },
            {
                showIf: (s) => s.has('stripe'),
                items: () => [{ text: 'Billing UI', source: 'stripe' }]
            },
            {
                showIf: (s) => s.has('r2'),
                items: () => [{ text: 'File manager', source: 'r2' }]
            },
            {
                showIf: (s) => s.has('postmark'),
                items: () => [{ text: 'Email dashboard', source: 'postmark' }]
            }
        ]
    },
    {
        category: 'Payments (Stripe)',
        showIf: (s) => s.has('stripe'),
        source: 'stripe',
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
        source: 'r2',
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
        source: 'postmark',
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
        source: 'mon',
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
        source: 'infra',
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

/** @type {Record<string, string>} */
export const sourceLabels = {
    init: 'init',
    model: 'model',
    client: 'client',
    stripe: 'stripe',
    r2: 'r2',
    postmark: 'postmark',
    infra: 'infra',
    mon: 'mon'
};

/** @type {Record<string, string>} */
export const sourceColors = {
    init: 'bg-blue-500/20 text-blue-400',
    model: 'bg-purple-500/20 text-purple-400',
    client: 'bg-orange-500/20 text-orange-400',
    stripe: 'bg-violet-500/20 text-violet-400',
    r2: 'bg-amber-500/20 text-amber-400',
    postmark: 'bg-rose-500/20 text-rose-400',
    infra: 'bg-emerald-500/20 text-emerald-400',
    mon: 'bg-cyan-500/20 text-cyan-400'
};

/**
 * Build wall data based on current state
 * @param {import('../stores/state.svelte.js').State} state
 * @returns {WallSection[]}
 */
export function buildWallData(state) {
    return wallSections
        .filter((section) => !section.showIf || section.showIf(state))
        .map((section) => {
            /** @type {WallItem[]} */
            const items = section.items.map((item) => {
                if (typeof item === 'string') {
                    return { text: item, source: section.source };
                }
                return item;
            });

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
