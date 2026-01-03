<script>
	import { onMount } from 'svelte';
	import { state as appState } from '$lib/stores/state.svelte.js';
	import { getCommand } from '$lib/data/commands.js';
	import { gsap } from '$lib/animations/gsap.js';

	/** @type {{ commandId: string, variant?: any, onComplete?: () => void, children?: import('svelte').Snippet }} */
	let { commandId, variant, onComplete, children } = $props();

	let container = $state();
	let commandTextElement = $state();
	let lineElement = $state();
	let itemsContainer = $state();
	
	let isCompleted = $state(false);

	const commandData = getCommand(commandId) || {
        id: 'error',
        label: 'Error',
        command: 'echo "Error: Command not found"',
        description: 'Error',
        baseOutputs: [{text: 'Error: Command not found'}],
        contextOutputs: []
    };
	
	// Determine the display command string
	let displayCommand = $derived.by(() => {
		if (variant) return variant.command;
		if (typeof commandData.command === 'function') return commandData.command();
		return commandData.command || `gof ${commandId}`;
	});

	// Compute outputs based on current state
	let outputs = $derived.by(() => {
		const base = commandData.baseOutputs || [];
		const context = (commandData.contextOutputs || []).filter(o => 
			!o.showIf || o.showIf(appState)
		);
		return [...base, ...context];
	});

	onMount(() => {
		const tl = gsap.timeline({
			onComplete: () => {
				isCompleted = true;
				if (onComplete) onComplete();
			}
		});

		// 1. Type the command
		tl.to(commandTextElement, {
			duration: 1.5,
			text: displayCommand,
			ease: "none"
		});

		// 2. Draw line down
		// We want the line to grow as items appear
		
        // Check if itemsContainer has children before animating
		const items = itemsContainer?.children;
		
		tl.to(lineElement, {
			height: '100%',
			duration: outputs.length * 0.3 + 0.5,
			ease: "power1.inOut"
		}, "+=0.2");

		// 3. Reveal items
		if (items && items.length > 0) {
            tl.fromTo(items, 
                { opacity: 0, x: -20 },
                { opacity: 1, x: 0, duration: 0.4, stagger: 0.3 },
                "<+=0.2" // Start 0.2s after line starts
            );
        }

		return () => tl.kill();
	});
</script>

<section 
	bind:this={container} 
	class="min-h-screen flex flex-col items-center pt-24 pb-12 px-6 snap-start relative"
>
	<!-- Command Box -->
	<div class="relative z-10 bg-surface border border-border rounded-xl p-4 md:p-6 w-full max-w-2xl shadow-lg mb-8">
		<div class="flex items-center gap-2 mb-2 text-xs text-muted font-mono">
			<div class="w-3 h-3 rounded-full bg-red-500/20"></div>
			<div class="w-3 h-3 rounded-full bg-yellow-500/20"></div>
			<div class="w-3 h-3 rounded-full bg-green-500/20"></div>
			<span class="ml-auto opacity-50">bash</span>
		</div>
		<div class="font-mono text-lg md:text-xl text-white min-h-[1.5em]">
			<span class="text-primary mr-2">$</span><span bind:this={commandTextElement}></span><span class="animate-pulse bg-primary/50 inline-block w-2 h-5 ml-1 align-middle"></span>
		</div>
	</div>

	<!-- Flow Container -->
	<div class="relative flex-grow w-full max-w-2xl flex flex-col items-center">
		<!-- The Line -->
		<div 
			bind:this={lineElement}
			class="absolute top-0 left-1/2 -translate-x-1/2 w-0.5 bg-gradient-to-b from-primary via-primary/50 to-transparent h-0 shadow-[0_0_10px_rgba(16,185,129,0.5)]"
		></div>

		<!-- Output Items -->
		<div bind:this={itemsContainer} class="w-full relative z-10 pt-4 pb-12 space-y-6">
			{#each outputs as output, i}
				<div class="relative flex items-center w-full group">
					<!-- Connector Dot -->
					<div class="absolute left-1/2 -translate-x-1/2 w-3 h-3 bg-bg border-2 border-primary rounded-full z-20 shadow-[0_0_8px_rgba(16,185,129,0.4)]"></div>
					
					<!-- Content Card -->
					<div class={`flex-1 flex ${i % 2 === 0 ? 'justify-end pr-12' : 'justify-start pl-12 order-last'}`}>
						<div class="bg-surface/80 backdrop-blur border border-border/50 px-4 py-3 rounded-lg text-sm md:text-base text-gray-300 font-mono shadow-sm hover:border-primary/30 transition-colors">
							<span class="text-success mr-2">✓</span>
							{typeof output.text === 'function' ? output.text(appState) : output.text}
						</div>
					</div>
					
					<!-- Spacer for the other side -->
					<div class="flex-1"></div>
				</div>
			{/each}
		</div>
	</div>

	<!-- Next Step Picker (Slot) -->
	<div class="w-full flex justify-center mt-auto min-h-[120px]">
		{#if isCompleted}
			{@render children?.()}
		{/if}
	</div>

</section>