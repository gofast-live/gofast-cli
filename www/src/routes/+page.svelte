<script>
	import Hero from '$lib/components/Hero.svelte';
	import Terminal from '$lib/components/Terminal.svelte';
	import CommandPicker from '$lib/components/CommandPicker.svelte';
	import StateTracker from '$lib/components/StateTracker.svelte';
	import CTA from '$lib/components/CTA.svelte';
	import { state as appState } from '$lib/stores/state.svelte.js';
	import { getCommand } from '$lib/data/commands.js';

	// Config - replace with actual URLs
	const discordUrl = '#'; // TODO: Add Discord URL
	const adminUrl = '#'; // TODO: Add admin app URL
	const v1Url = 'https://gofast.live';

	/** @type {Array<{id: string, command: string, description: string, outputs: any[], variant?: string}>} */
	let executedCommands = $state([]);

	/** @type {boolean} */
	let showPicker = $state(false);

	/** @type {HTMLElement | null} */
	let bottomRef = $state(null);

	function handleInitComplete() {
		appState.init();
		showPicker = true;
		scrollToBottom();
	}

	/**
	 * @param {string} commandId
	 * @param {string} [variant]
	 */
	function handleCommandSelect(commandId, variant) {
		if (commandId === 'finish') {
			appState.finish();
			showPicker = false;
			scrollToBottom();
			return;
		}

		const cmd = getCommand(commandId);
		if (!cmd) return;

		showPicker = false;

		// Build the command info
		let commandText = '';
		let outputs = [...cmd.baseOutputs];

		if (commandId === 'model' && variant) {
			const modelVariant = cmd.variants?.find((v) => v.name === variant);
			if (modelVariant) {
				commandText = modelVariant.command;
				appState.addModel(variant);
			}
		} else {
			commandText = cmd.command?.() ?? '';
			appState.add(commandId);
		}

		// Add context outputs
		const contextOutputs = cmd.contextOutputs.filter(
			(o) => !o.showIf || o.showIf(appState)
		);
		outputs = [...outputs, ...contextOutputs];

		executedCommands = [
			...executedCommands,
			{
				id: `${commandId}-${Date.now()}`,
				command: commandText,
				description: cmd.description,
				outputs,
				variant
			}
		];

		// Wait for terminal to render, then scroll
		setTimeout(() => scrollToBottom(), 100);
	}

	function handleTerminalComplete() {
		if (!appState.finished) {
			showPicker = true;
			scrollToBottom();
		}
	}

	function scrollToBottom() {
		setTimeout(() => {
			bottomRef?.scrollIntoView({ behavior: 'smooth', block: 'end' });
		}, 100);
	}

	// Get available commands for picker
	let availableCommands = $derived(
		appState.availableCommands.map((c) => ({
			id: c.id,
			label: c.label,
			hasSubPicker: c.hasSubPicker,
			variants: c.variants,
			disabled: false
		}))
	);
</script>

<svelte:head>
	<title>GoFast - Building blocks for Go</title>
</svelte:head>

<main class="min-h-screen">
	<!-- Hero with init command -->
	{#if !appState.initialized}
		<Hero {discordUrl} {v1Url} onInitComplete={handleInitComplete} />
	{:else}
		<!-- Header (shown after init) -->
		<header class="flex justify-between items-center p-4 md:p-6 sticky top-0 bg-[var(--bg)]/80 backdrop-blur-sm z-30">
			<a href="/" class="flex items-center gap-2">
				<img src="/logo.svg" alt="GoFast" class="h-6" />
			</a>
			<div class="flex items-center gap-4">
				<a
					href={discordUrl}
					target="_blank"
					rel="noopener noreferrer"
					aria-label="Join Discord"
					class="text-[var(--text-muted)] hover:text-[var(--text)] transition-colors"
				>
					<svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
						<path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z"/>
					</svg>
				</a>
				<a
					href={v1Url}
					target="_blank"
					rel="noopener noreferrer"
					class="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors"
				>
					V1 →
				</a>
			</div>
		</header>

		<!-- Main content area -->
		<div class="flex gap-6 p-4 md:p-6">
			<!-- State tracker sidebar -->
			<aside class="hidden lg:block sticky top-24 h-fit">
				<StateTracker
					initialized={appState.initialized}
					models={appState.models}
					completed={appState.completed}
				/>
			</aside>

			<!-- Terminal flow -->
			<div class="flex-1 flex flex-col items-center gap-8 pb-32">
				<!-- Init terminal (already complete) -->
				<div class="w-full max-w-xl">
					<Terminal
						command="gof init myproject"
						description="Creating project structure..."
						outputs={getCommand('init')?.baseOutputs ?? []}
						autoRun={true}
					/>
				</div>

				<!-- Executed commands -->
				{#each executedCommands as cmd (cmd.id)}
					<div class="w-full max-w-xl">
						<Terminal
							command={cmd.command}
							description={cmd.description}
							outputs={cmd.outputs}
							appState={appState}
							autoRun={true}
							onComplete={handleTerminalComplete}
						/>
					</div>
				{/each}

				<!-- Command picker -->
				{#if showPicker && !appState.finished}
					<div class="text-center">
						<CommandPicker
							commands={availableCommands}
							modelVariants={appState.availableModelVariants}
							onSelect={handleCommandSelect}
						/>
					</div>
				{/if}

				<!-- CTA -->
				{#if appState.finished}
					<CTA
						stackSummary={appState.stackSummary}
						{adminUrl}
						{discordUrl}
						{v1Url}
					/>
				{/if}

				<div bind:this={bottomRef}></div>
			</div>
		</div>
	{/if}
</main>
