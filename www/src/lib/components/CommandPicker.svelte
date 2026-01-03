<script>
	/**
	 * @typedef {Object} CommandOption
	 * @property {string} id
	 * @property {string} label
	 * @property {boolean} [disabled]
	 * @property {boolean} [hasSubPicker]
	 * @property {Array<{id: string, name: string, tagline: string}>} [variants]
	 */

	/**
	 * @type {{
	 *   commands: CommandOption[],
	 *   onSelect: (id: string, variant?: string) => void,
	 *   modelVariants?: Array<{id: string, name: string, tagline: string}>
	 * }}
	 */
	let {
		commands = [],
		onSelect = () => {},
		modelVariants = []
	} = $props();

	let showModelPicker = $state(false);

	/**
	 * @param {CommandOption} cmd
	 */
	function handleClick(cmd) {
		if (cmd.disabled) return;

		if (cmd.hasSubPicker && modelVariants.length > 0) {
			showModelPicker = true;
		} else {
			onSelect(cmd.id);
		}
	}

	/**
	 * @param {{id: string, name: string, tagline: string}} variant
	 */
	function handleModelSelect(variant) {
		showModelPicker = false;
		onSelect('model', variant.name);
	}

	function handleBackdropClick() {
		showModelPicker = false;
	}
</script>

<div class="command-picker">
	{#if !showModelPicker}
		<p class="text-[var(--text-muted)] text-sm mb-4">What's next?</p>
		<div class="grid grid-cols-3 gap-2 max-w-md">
			{#each commands as cmd}
				<button
					onclick={() => handleClick(cmd)}
					disabled={cmd.disabled}
					class="px-4 py-2 rounded-lg font-mono text-sm transition-all
						{cmd.disabled
							? 'bg-[var(--surface)] text-[var(--text-subtle)] cursor-not-allowed opacity-50'
							: 'bg-[var(--surface)] text-[var(--text)] border border-[var(--border)] hover:border-[var(--primary)] hover:bg-[var(--surface-hover)]'
						}"
				>
					{cmd.label}
				</button>
			{/each}
		</div>

		<button
			onclick={() => onSelect('finish')}
			class="mt-6 text-sm text-[var(--text-muted)] hover:text-[var(--primary)] transition-colors"
		>
			Finish building →
		</button>
	{:else}
		<!-- Model sub-picker -->
		<button
			onclick={handleBackdropClick}
			class="fixed inset-0 bg-black/50 z-40"
			aria-label="Close picker"
		></button>

		<div class="relative z-50">
			<p class="text-[var(--text-muted)] text-sm mb-4">Pick a model:</p>
			<div class="grid grid-cols-3 gap-2 max-w-md">
				{#each modelVariants as variant}
					<button
						onclick={() => handleModelSelect(variant)}
						class="px-4 py-3 rounded-lg text-left transition-all bg-[var(--surface)] border border-[var(--border)] hover:border-[var(--primary)] hover:bg-[var(--surface-hover)]"
					>
						<div class="font-mono text-sm text-[var(--text)]">{variant.name}</div>
						<div class="text-xs text-[var(--text-muted)] mt-1">{variant.tagline}</div>
					</button>
				{/each}
			</div>
			<button
				onclick={handleBackdropClick}
				class="mt-4 text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors"
			>
				← Back
			</button>
		</div>
	{/if}
</div>
