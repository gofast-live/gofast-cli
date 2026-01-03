<script>
	/**
	 * @type {{
	 *   initialized: boolean,
	 *   models: string[],
	 *   completed: Set<string>
	 * }}
	 */
	let {
		initialized = false,
		models = [],
		completed = new Set()
	} = $props();

	const allFeatures = ['client', 'stripe', 'r2', 'postmark', 'infra'];
</script>

<div class="state-tracker bg-[var(--surface)] border border-[var(--border)] rounded-lg p-4 font-mono text-sm">
	<div class="text-[var(--text)] font-medium mb-3">myproject</div>

	<div class="space-y-2 text-xs">
		<!-- Init status -->
		<div class="flex items-center gap-2">
			{#if initialized}
				<span class="text-[var(--success)]">✓</span>
				<span class="text-[var(--text)]">init</span>
			{:else}
				<span class="text-[var(--text-subtle)]">○</span>
				<span class="text-[var(--text-muted)]">init</span>
			{/if}
		</div>

		<!-- Models -->
		{#if models.length > 0}
			<div class="flex items-center gap-2">
				<span class="text-[var(--success)]">✓</span>
				<span class="text-[var(--text)]">models: {models.join(', ')}</span>
			</div>
		{/if}

		<!-- Other features -->
		<div class="flex flex-wrap gap-x-4 gap-y-1">
			{#each allFeatures as feature}
				<div class="flex items-center gap-1">
					{#if completed.has(feature)}
						<span class="text-[var(--success)]">✓</span>
						<span class="text-[var(--text)]">{feature}</span>
					{:else}
						<span class="text-[var(--text-subtle)]">○</span>
						<span class="text-[var(--text-muted)]">{feature}</span>
					{/if}
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	.state-tracker {
		min-width: 280px;
	}
</style>
