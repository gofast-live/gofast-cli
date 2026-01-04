<script>
	import { onMount } from 'svelte';
	import { state as appState } from '$lib/stores/state.svelte.js';
	import { getCommand } from '$lib/data/commands.js';
	import { gsap } from '$lib/animations/gsap.js';
	import { fade, fly } from 'svelte/transition';

	/** @type {{ commandId: string, variant?: any, onComplete?: () => void, children?: import('svelte').Snippet }} */
	const { commandId, variant, onComplete, children } = $props();

	let container = $state();
	let commandTextElement = $state();
	let lineElement = $state();
	let itemsContainer = $state();

	let isCompleted = $state(false);
	
	// State for the hovered item details
	/** @type {import('$lib/data/commands.js').OutputDetails | null} */
	let hoveredDetails = $state(null);

	let commandData = $derived(
		getCommand(commandId) || {
			id: 'error',
			label: 'Error',
			command: 'echo "Error: Command not found"',
			description: 'Error',
			baseOutputs: [{ text: 'Error: Command not found' }],
			contextOutputs: []
		}
	);

	// Determine the display command string
	let displayCommand = $derived.by(() => {
		if (variant) return variant.command;
		if (typeof commandData.command === 'function') return commandData.command();
		return commandData.command || `gof ${commandId}`;
	});

	// Compute outputs based on current state
	let outputs = $derived.by(() => {
		const base = commandData.baseOutputs || [];
		const context = (commandData.contextOutputs || []).filter(
			(o) => !o.showIf || o.showIf(appState)
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
		// Calculate duration based on length: ~30 chars/sec
		// longer commands (like model) will take ~1.5s, shorter ones ~0.5s
		const typeDuration = Math.max(0.5, displayCommand.length * 0.03);

		tl.to(commandTextElement, {
			duration: typeDuration,
			text: displayCommand,
			ease: 'none'
		});

		// 2. Draw line down
		// We want the line to grow as items appear

		// Check if itemsContainer has children before animating
		const items = itemsContainer?.children;

		tl.to(
			lineElement,
			{
				height: '100%',
				duration: outputs.length * 0.3 + 0.5,
				ease: 'power1.inOut'
			},
			'+=0.2'
		);

		// 3. Reveal items
		if (items && items.length > 0) {
			tl.fromTo(
				items,
				{ opacity: 0, x: -20 },
				{ opacity: 1, x: 0, duration: 0.4, stagger: 0.3 },
				'<+=0.2' // Start 0.2s after line starts
			);
		}

		return () => tl.kill();
	});

	function handleMouseEnter(details) {
		hoveredDetails = details;
	}

	function handleMouseLeave() {
		hoveredDetails = null;
	}
</script>

<section bind:this={container} class="flex flex-col items-center pt-12 pb-12 px-6 relative w-full">
	<!-- Fixed Detail Panel (Desktop) -->
	<div
		class="fixed top-1/2 -translate-y-1/2 right-8 w-80 bg-surface/90 backdrop-blur border border-border rounded-xl p-6 hidden xl:block shadow-2xl transition-opacity duration-300 pointer-events-none z-50"
		class:opacity-0={!hoveredDetails}
		class:opacity-100={!!hoveredDetails}
	>
		{#if hoveredDetails}
			<div class="space-y-6">
				{#if hoveredDetails.files && hoveredDetails.files.length > 0}
					<div>
						<h4 class="text-primary font-mono text-sm mb-3 border-b border-border pb-2">
							[Generated Files]
						</h4>
						<ul class="space-y-1.5">
							{#each hoveredDetails.files as file}
								<li class="text-xs text-gray-300 font-mono flex items-start gap-2">
									<span class="text-border">└</span>
									{file}
								</li>
							{/each}
						</ul>
					</div>
				{/if}

				{#if hoveredDetails.features && hoveredDetails.features.length > 0}
					<div>
						<h4 class="text-primary font-mono text-sm mb-3 border-b border-border pb-2">
							[Logic Wired]
						</h4>
						<ul class="space-y-1.5">
							{#each hoveredDetails.features as feature}
								<li class="text-xs text-gray-300 font-sans flex items-start gap-2">
									<span class="text-primary text-[10px] mt-0.5">●</span>
									{feature}
								</li>
							{/each}
						</ul>
					</div>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Command Box -->
	<div
		class="relative z-10 bg-surface border border-border rounded-xl p-4 md:p-6 w-full max-w-2xl shadow-lg mb-8"
	>
		<div class="flex items-center gap-2 mb-2 text-xs text-muted font-mono">
			<div class="w-3 h-3 rounded-full bg-red-500/20"></div>
			<div class="w-3 h-3 rounded-full bg-yellow-500/20"></div>
			<div class="w-3 h-3 rounded-full bg-green-500/20"></div>
			<span class="ml-auto opacity-50">bash</span>
		</div>
		<div class="font-mono text-lg md:text-xl text-white min-h-[1.5em]">
			<span class="text-primary mr-2">$</span><span bind:this={commandTextElement}></span><span
				class="animate-pulse bg-primary/50 inline-block w-2 h-5 ml-1 align-middle"
				class:hidden={isCompleted}
			></span>
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
					<div
						class="absolute left-1/2 -translate-x-1/2 w-3 h-3 bg-bg border-2 border-primary rounded-full z-20 shadow-[0_0_8px_rgba(16,185,129,0.4)] transition-transform group-hover:scale-125 duration-300"
					></div>

					<!-- Content Card -->
					<div
						class={`flex-1 flex ${i % 2 === 0 ? 'justify-end pr-4 md:pr-12' : 'justify-start pl-4 md:pl-12 order-last'}`}
					>
						<!-- Interactive Item -->
						<button
							class="text-left bg-surface/80 backdrop-blur border border-border/50 px-4 py-3 rounded-lg text-sm md:text-base text-gray-300 font-mono shadow-sm hover:border-primary/50 hover:bg-surface-hover hover:text-white hover:shadow-[0_0_15px_rgba(16,185,129,0.1)] transition-all duration-300 cursor-help"
							onmouseenter={() => handleMouseEnter(output.details)}
							onmouseleave={handleMouseLeave}
							onfocus={() => handleMouseEnter(output.details)}
							onblur={handleMouseLeave}
						>
							<span class="text-success mr-2">✓</span>
							{typeof output.text === 'function' ? output.text(appState) : output.text}
						</button>
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
			<div in:fade={{ duration: 500 }} class="w-full flex justify-center">
				{@render children?.()}
			</div>
		{/if}
	</div>
</section>
