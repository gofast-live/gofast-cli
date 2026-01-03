/**
 * @typedef {Object} Output
 * @property {string | ((state: import('../stores/state.svelte.js').State) => string)} text
 * @property {(state: import('../stores/state.svelte.js').State) => boolean} [showIf]
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
			{ text: 'OAuth (GitHub + Google)' },
			{ text: 'Bitwise role authorization' },
			{ text: 'Docker Compose setup' },
			{ text: 'GitHub Actions CI/CD' },
			{ text: 'PR preview deployments' },
			{ text: 'PostgreSQL + SQLC' },
			{ text: 'ConnectRPC transport' }
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
			{ text: 'SQL migration' },
			{ text: 'SQLC queries' },
			{ text: 'Proto definitions' },
			{ text: 'Domain service layer' },
			{ text: 'Transport handlers' },
			{ text: 'Validation + tests' }
		],
		contextOutputs: [
			{ text: 'Svelte pages generated', showIf: (s) => s.has('client') },
			{ text: 'Subscription checks wired', showIf: (s) => s.has('stripe') }
		]
	},
	{
		id: 'client',
		label: 'client',
		command: () => 'gof client svelte',
		description: 'Adding Svelte frontend...',
		baseOutputs: [
			{ text: 'SvelteKit scaffold' },
			{ text: 'Auth integration' },
			{ text: 'Type-safe API client' }
		],
		contextOutputs: [
			{ text: (s) => `Generated pages for: ${s.models.join(', ')}`, showIf: (s) => s.models.length > 0 },
			{ text: 'Stripe billing UI', showIf: (s) => s.has('stripe') },
			{ text: 'File management UI', showIf: (s) => s.has('r2') },
			{ text: 'Email dashboard', showIf: (s) => s.has('postmark') }
		]
	},
	{
		id: 'stripe',
		label: 'stripe',
		command: () => 'gof add stripe',
		description: 'Adding Stripe payments...',
		baseOutputs: [
			{ text: 'Payment domain service' },
			{ text: 'Subscriptions migration' },
			{ text: 'Webhook handlers' },
			{ text: 'Access control integration' }
		],
		contextOutputs: [
			{ text: 'Billing UI components', showIf: (s) => s.has('client') }
		]
	},
	{
		id: 'r2',
		label: 'r2',
		command: () => 'gof add r2',
		description: 'Adding file storage...',
		baseOutputs: [
			{ text: 'File domain service' },
			{ text: 'Files migration' },
			{ text: 'S3-compatible uploads' }
		],
		contextOutputs: [
			{ text: 'File manager UI', showIf: (s) => s.has('client') }
		]
	},
	{
		id: 'postmark',
		label: 'postmark',
		command: () => 'gof add postmark',
		description: 'Adding email service...',
		baseOutputs: [
			{ text: 'Email domain service' },
			{ text: 'Emails migration' },
			{ text: 'Template support' }
		],
		contextOutputs: [
			{ text: 'Email dashboard UI', showIf: (s) => s.has('client') }
		]
	},
	{
		id: 'infra',
		label: 'infra',
		command: () => 'gof infra',
		description: 'Adding production infrastructure...',
		baseOutputs: [
			{ text: 'Kubernetes manifests' },
			{ text: 'Terraform configs' },
			{ text: 'OpenTelemetry setup' },
			{ text: 'GitHub Actions deploy' }
		],
		contextOutputs: [
			{ text: 'Cloudflare Workers (client)', showIf: (s) => s.has('client') },
			{ text: 'Stripe secrets configured', showIf: (s) => s.has('stripe') },
			{ text: 'R2 bucket configured', showIf: (s) => s.has('r2') },
			{ text: 'Postmark configured', showIf: (s) => s.has('postmark') }
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
