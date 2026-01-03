import { commands } from '$lib/data/commands.js';

/**
 * @typedef {Object} State
 * @property {Set<string>} completed
 * @property {string[]} models
 * @property {boolean} initialized
 * @property {(id: string) => boolean} has
 * @property {(name: string) => boolean} hasModel
 */

/** @type {Set<string>} */
let completed = $state(new Set());

/** @type {string[]} */
let models = $state([]);

/** @type {boolean} */
let initialized = $state(false);

/** @type {boolean} */
let finished = $state(false);

export const state = {
	get completed() {
		return completed;
	},
	get models() {
		return models;
	},
	get initialized() {
		return initialized;
	},
	get finished() {
		return finished;
	},

	/**
	 * Check if a feature is completed
	 * @param {string} id
	 * @returns {boolean}
	 */
	has(id) {
		return completed.has(id);
	},

	/**
	 * Check if a model variant was added
	 * @param {string} name
	 * @returns {boolean}
	 */
	hasModel(name) {
		return models.includes(name);
	},

	/** Mark init as complete */
	init() {
		initialized = true;
	},

	/**
	 * Add a completed feature
	 * @param {string} id
	 */
	add(id) {
		completed = new Set([...completed, id]);
	},

	/**
	 * Add a model
	 * @param {string} name
	 */
	addModel(name) {
		if (!models.includes(name)) {
			models = [...models, name];
		}
	},

	/** Mark as finished */
	finish() {
		finished = true;
	},

	/** Reset state */
	reset() {
		completed = new Set();
		models = [];
		initialized = false;
		finished = false;
	},

	/** Get commands still available */
	get availableCommands() {
		return commands.filter((c) => {
			if (c.id === 'init') return false;
			if (c.id === 'model') {
				// Show model if any variants remain
				const usedCount = c.variants?.filter((v) => models.includes(v.name)).length ?? 0;
				return usedCount < (c.variants?.length ?? 0);
			}
			return !completed.has(c.id);
		});
	},

	/** Get remaining model variants */
	get availableModelVariants() {
		const modelCmd = commands.find((c) => c.id === 'model');
		return modelCmd?.variants?.filter((v) => !models.includes(v.name)) ?? [];
	},

	/** Get summary of what was built */
	get stackSummary() {
		const parts = ['Go', 'ConnectRPC'];
		if (completed.has('client')) parts.push('Svelte');
		if (completed.has('stripe')) parts.push('Stripe');
		if (completed.has('r2')) parts.push('R2');
		if (completed.has('postmark')) parts.push('Postmark');
		if (completed.has('infra')) parts.push('K8s');
		return parts.join(' + ');
	}
};
