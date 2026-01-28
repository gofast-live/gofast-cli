<script>
	import { state as appState } from '$lib/stores/state.svelte.js';
	import { fade, fly } from 'svelte/transition';
    import { ArrowRight, ArrowLeft } from '@lucide/svelte';

	/** @type {{ onSelect: (cmd: { id: string, variant?: any }) => void }} */
	let { onSelect } = $props();

	let showModelVariants = $state(false);

	/** @param {{ id: string }} cmd */
	function handleCommandClick(cmd) {
		if (cmd.id === 'model') {
			showModelVariants = true;
		} else {
			onSelect({ id: cmd.id });
		}
	}

	/** @param {any} variant */
	function handleVariantClick(variant) {
		onSelect({ id: 'model', variant });
		showModelVariants = false;
	}

    function handleFinish() {
        appState.finish();
        onSelect({ id: 'finish' });
    }
</script>

<div class="w-full max-w-3xl mx-auto mt-8 bg-surface/50 border border-border rounded-xl p-6 backdrop-blur-sm" in:fade={{ duration: 200 }}>
	<div class="text-sm text-muted mb-4 font-mono">
		{#if showModelVariants}
			Select a data model:
		{:else}
			$ What's next? (More options coming soon!)
		{/if}
	</div>

	<div class="flex flex-wrap gap-3">
		{#if showModelVariants}
			{#each appState.availableModelVariants as variant}
				<button
					class="flex flex-col items-start p-4 bg-surface border border-border hover:border-primary/50 hover:bg-surface-hover rounded-lg transition-all text-left min-w-[200px]"
					onclick={() => handleVariantClick(variant)}
					in:fly={{ y: 10, duration: 300 }}
				>
					<span class="font-mono text-primary font-bold">{variant.name}</span>
					<span class="text-xs text-muted mt-1">{variant.tagline}</span>
				</button>
			{/each}
            <button
                class="px-4 py-2 text-sm text-muted hover:text-white transition-colors flex items-center gap-2"
                onclick={() => showModelVariants = false}
            >
                <ArrowLeft size={16} /> Back
            </button>
		{:else}
			{#each appState.availableCommands as cmd}
				<button
					class="px-5 py-3 bg-surface border border-border hover:border-primary/50 hover:bg-surface-hover rounded-lg font-mono text-sm text-white transition-all shadow-sm hover:shadow-md hover:-translate-y-0.5"
					onclick={() => handleCommandClick(cmd)}
				>
					{cmd.label}
				</button>
			{/each}

            <!-- Finish Button -->
            <div class="flex-grow"></div>
             <button
                class="px-5 py-3 bg-primary/10 border border-primary/20 hover:bg-primary/20 text-primary rounded-lg font-mono text-sm transition-all ml-auto flex items-center gap-2"
                onclick={handleFinish}
            >
                Deploy <ArrowRight size={16} />
            </button>
		{/if}
	</div>
</div>
