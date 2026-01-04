<script>
	import { onMount } from 'svelte';
	import { state as appState } from '$lib/stores/state.svelte.js';
	import { getCommand } from '$lib/data/commands.js';
	import { gsap } from '$lib/animations/gsap.js';
	import { fade, fly } from 'svelte/transition';
	import { Info, X, Zap } from '@lucide/svelte';

	/** @type {{ commandId: string, variant?: any, onComplete?: () => void, children?: import('svelte').Snippet }} */
	const { commandId, variant, onComplete, children } = $props();

	let container = $state();
	let commandTextElement = $state();
	let lineElement = $state();
	let itemsContainer = $state();

	let isCompleted = $state(false);
	
	// State for the hovered/active item details
	/** @type {import('$lib/data/commands.js').OutputDetails | null} */
	let activeDetails = $state(null);

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
		const typeDuration = Math.max(0.5, displayCommand.length * 0.03);

		tl.to(commandTextElement, {
			duration: typeDuration,
			text: displayCommand,
			ease: 'none'
		});

		// 2. Draw line down
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
				'<+=0.2'
			);
		}

		return () => tl.kill();
	});

	// Desktop Hover
	function handleMouseEnter(details) {
		if (window.matchMedia('(min-width: 1280px)').matches) {
			activeDetails = details;
		}
	}

	function handleMouseLeave() {
		if (window.matchMedia('(min-width: 1280px)').matches) {
			activeDetails = null;
		}
	}

	// Mobile Tap
	function handleTap(details) {
		if (activeDetails === details) {
			activeDetails = null; // Toggle off
		} else {
			activeDetails = details;
		}
	}
</script>

<section bind:this={container} class="flex flex-col items-center pt-12 pb-12 px-4 md:px-6 relative w-full">
	
	<!-- FIXED PANELS -->

	<!-- Desktop: Fixed Right Panel -->
	{#if activeDetails}
		<div
			transition:fade={{ duration: 200 }}
			class="fixed top-1/2 -translate-y-1/2 right-8 w-80 bg-surface/90 backdrop-blur border border-border rounded-xl p-6 hidden xl:block shadow-2xl pointer-events-none z-50"
		>
			<div class="space-y-6">
				{#if activeDetails.files && activeDetails.files.length > 0}
					<div>
						<h4 class="text-primary font-mono text-sm mb-3 border-b border-border pb-2">
							[Generated Files]
						</h4>
						<ul class="space-y-1.5">
							{#each activeDetails.files as file}
								<li class="text-xs text-gray-300 font-mono flex items-start gap-2">
									<span class="text-border">└</span>
									{file}
								</li>
							{/each}
						</ul>
					</div>
				{/if}

				{#if activeDetails.features && activeDetails.features.length > 0}
					<div>
						<h4 class="text-primary font-mono text-sm mb-3 border-b border-border pb-2">
							[Logic Wired]
						</h4>
						<ul class="space-y-1.5">
							{#each activeDetails.features as feature}
								<li class="text-xs text-gray-300 font-sans flex items-start gap-2">
									<span class="text-primary text-[10px] mt-0.5">●</span>
									{feature}
								</li>
							{/each}
						</ul>
					</div>
				{/if}
			</div>
		</div>
	{/if}

	<!-- Mobile: Fixed Bottom Sheet -->
	{#if activeDetails}
		<div 
			transition:fly={{ y: '100%', duration: 300, opacity: 1 }}
			class="fixed bottom-0 left-0 w-full bg-surface border-t border-border p-6 xl:hidden z-50 shadow-[0_-5px_20px_rgba(0,0,0,0.5)]"
		>
			<div class="flex justify-between items-center mb-4">
				<span class="text-xs font-mono text-primary uppercase tracking-wider">Under the hood</span>
				<button class="text-muted hover:text-white" onclick={() => activeDetails = null}>
					<X size={20} />
				</button>
			</div>
			<div class="grid grid-cols-1 gap-6 max-h-[40vh] overflow-y-auto">
				{#if activeDetails.files?.length}
					<div>
						<h4 class="text-white font-mono text-xs mb-2 border-b border-border/50 pb-1">Generated Files</h4>
						<ul class="space-y-1">
							{#each activeDetails.files as file}
								<li class="text-xs text-gray-400 font-mono flex items-start gap-2">
									<span class="text-border">└</span> {file}
								</li>
							{/each}
						</ul>
					</div>
				{/if}
				{#if activeDetails.features?.length}
					<div>
						<h4 class="text-white font-mono text-xs mb-2 border-b border-border/50 pb-1">Logic Wired</h4>
						<ul class="space-y-1">
							{#each activeDetails.features as feature}
								<li class="text-xs text-gray-400 font-sans flex items-start gap-2">
									<span class="text-primary text-[10px] mt-0.5">●</span> {feature}
								</li>
							{/each}
						</ul>
					</div>
				{/if}
			</div>
		</div>
	{/if}

	<!-- MAIN CONTENT -->

	<!-- Command Box -->
	<div class="relative z-10 bg-surface border border-border rounded-xl p-4 md:p-6 w-full max-w-2xl shadow-lg mb-8">
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
	<div class="relative flex-grow w-full max-w-2xl flex flex-col items-start md:items-center pl-8 md:pl-0">
		<!-- The Line -->
		<div 
			bind:this={lineElement}
			class="absolute top-0 left-0 md:left-1/2 md:-translate-x-1/2 w-0.5 bg-gradient-to-b from-primary via-primary/50 to-transparent h-0 shadow-[0_0_10px_rgba(16,185,129,0.5)]"
		></div>

		<!-- Output Items -->
		        <div bind:this={itemsContainer} class="w-full relative z-10 pt-4 pb-12 space-y-6">
		            {#each outputs as output, i}
		                <div class={`relative flex items-center w-full group ${output.dependency ? 'pt-12' : ''}`}>
		                    
		                    <!-- Dependency Badge (Vertical Node) -->
		                    {#if output.dependency}
		                        <div class="absolute top-2 left-0 md:left-1/2 -translate-x-1/2 z-20">
		                            <div class="bg-surface border border-primary/30 rounded-full px-2 py-1 text-[10px] font-mono text-primary flex items-center gap-1 shadow-sm whitespace-nowrap">
		                                <Zap size={10} class="text-primary fill-primary/20" />
		                                {output.dependency}
		                            </div>
		                        </div>
		                    {/if}
		
		                    <!-- Connector Dot -->
		                    <!-- Mobile: Left aligned with line. Desktop: Centered -->
		                    <div
		                        class="absolute left-0 -translate-x-1/2 md:left-1/2 w-3 h-3 bg-bg border-2 border-primary rounded-full z-20 shadow-[0_0_8px_rgba(16,185,129,0.4)] transition-transform group-hover:scale-125 duration-300"
		                    ></div>
		
		                    <!-- Content Card -->
		                    <!-- Mobile: Always right of line (pl-8). Desktop: Alternating -->
		                    <div
		                        class={`flex-1 flex w-full relative
		                            pl-6 md:pl-0 
		                            ${i % 2 === 0 ? 'md:justify-end md:pr-12' : 'md:justify-start md:pl-12 md:order-last'}`}
		                    >
		                        <!-- Interactive Item -->
		                        <button
		                            class="group/btn relative text-left bg-surface/80 backdrop-blur border border-border/50 px-4 py-3 rounded-lg text-sm md:text-base text-gray-300 font-mono shadow-sm hover:border-primary/50 hover:bg-surface-hover hover:text-white hover:shadow-[0_0_15px_rgba(16,185,129,0.1)] transition-all duration-300 cursor-pointer flex items-center gap-2 pr-8 w-full md:w-auto"
		                            onmouseenter={() => handleMouseEnter(output.details)}
		                            onmouseleave={handleMouseLeave}
		                            onclick={() => handleTap(output.details)}
		                        >
		                            <span class="text-success">✓</span>
		                            <span class="flex-grow truncate">{typeof output.text === 'function' ? output.text(appState) : output.text}</span>
		                            
		                            <!-- Info Icon (Mobile: Visible, Desktop: Hover) -->
		                            <span class="absolute right-3 top-1/2 -translate-y-1/2 opacity-50 md:opacity-0 md:group-hover/btn:opacity-100 transition-opacity text-primary">
		                                <Info size={14} />
		                            </span>
		                        </button>
		                    </div>
		
		                    <!-- Spacer (Desktop only) -->
		                    <div class="hidden md:block flex-1"></div>
		                </div>
		            {/each}
		        </div>	</div>

	<!-- Next Step Picker (Slot) -->
	<div class="w-full flex justify-center mt-auto min-h-[120px]">
		{#if isCompleted}
			<div in:fade={{ duration: 500 }} class="w-full flex justify-center">
				{@render children?.()}
			</div>
		{/if}
	</div>
</section>