<script>
	import { onMount } from 'svelte';
	import { gsap } from 'gsap';
	import { TextPlugin } from 'gsap/TextPlugin';

	/**
	 * @type {{
	 *   command: string,
	 *   description?: string,
	 *   outputs?: Array<{text: string | ((s: any) => string), showIf?: (s: any) => boolean}>,
	 *   appState?: any,
	 *   onComplete?: () => void,
	 *   autoRun?: boolean
	 * }}
	 */
	let {
		command,
		description = '',
		outputs = [],
		appState = null,
		onComplete = () => {},
		autoRun = false
	} = $props();

	/** @type {HTMLElement | null} */
	let terminalEl = $state(null);
	/** @type {HTMLElement | null} */
	let commandEl = $state(null);
	/** @type {HTMLElement | null} */
	let outputEl = $state(null);
	/** @type {boolean} */
	let isRunning = $state(false);
	/** @type {boolean} */
	let isComplete = $state(false);
	/** @type {Array<{text: string, isDescription?: boolean, isOutput?: boolean, isDone?: boolean}>} */
	let visibleOutputs = $state([]);

	// Filter outputs based on state conditions
	let filteredOutputs = $derived(() => {
		return outputs.filter((/** @type {any} */ o) => !o.showIf || (appState && o.showIf(appState)));
	});

	onMount(() => {
		gsap.registerPlugin(TextPlugin);
		if (autoRun) {
			run();
		}
	});

	export function run() {
		if (isRunning || isComplete) return;
		isRunning = true;

		const tl = gsap.timeline({
			onComplete: () => {
				isComplete = true;
				isRunning = false;
				onComplete();
			}
		});

		// Type out command
		tl.to(commandEl, {
			duration: command.length * 0.03,
			text: command,
			ease: 'none'
		});

		// Show description
		if (description) {
			tl.to({}, {
				duration: 0.3,
				onComplete: () => {
					visibleOutputs = [{ text: description, isDescription: true }];
				}
			});
		}

		// Show outputs one by one
		const resolved = filteredOutputs();
		resolved.forEach((/** @type {any} */ output) => {
			tl.to({}, {
				duration: 0.15,
				onComplete: () => {
					const text = typeof output.text === 'function'
						? output.text(appState)
						: output.text;
					visibleOutputs = [...visibleOutputs.filter((/** @type {any} */ o) => !o.isDescription), { text, isOutput: true }];
				}
			});
		});

		// Final "Done" message
		tl.to({}, {
			duration: 0.2,
			onComplete: () => {
				visibleOutputs = [...visibleOutputs, { text: `Done in ${(Math.random() * 1.5 + 1).toFixed(1)}s`, isDone: true }];
			}
		});
	}
</script>

<div
	bind:this={terminalEl}
	class="terminal bg-[var(--surface)] border border-[var(--border)] rounded-lg overflow-hidden font-mono text-sm"
>
	<!-- Terminal header -->
	<div class="flex items-center gap-2 px-4 py-2 border-b border-[var(--border)] bg-[var(--bg)]">
		<div class="w-3 h-3 rounded-full bg-[#ff5f56]"></div>
		<div class="w-3 h-3 rounded-full bg-[#ffbd2e]"></div>
		<div class="w-3 h-3 rounded-full bg-[#27ca40]"></div>
	</div>

	<!-- Terminal content -->
	<div class="p-4 space-y-2">
		<!-- Command line -->
		<div class="flex items-center gap-2">
			<span class="text-[var(--primary)]">$</span>
			<span bind:this={commandEl} class="text-[var(--text)]"></span>
			{#if !isRunning && !isComplete}
				<span class="animate-pulse">_</span>
			{/if}
		</div>

		<!-- Outputs -->
		<div bind:this={outputEl} class="space-y-1 pl-4">
			{#each visibleOutputs as output}
				{#if output.isDescription}
					<div class="text-[var(--text-muted)]">{output.text}</div>
				{:else if output.isDone}
					<div class="text-[var(--text-muted)] mt-2">{output.text}</div>
				{:else}
					<div class="flex items-center gap-2 text-[var(--text)]">
						<span class="text-[var(--success)]">✓</span>
						<span>{output.text}</span>
					</div>
				{/if}
			{/each}
		</div>
	</div>
</div>

<style>
	.terminal {
		min-width: 400px;
		max-width: 600px;
	}

	@keyframes pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0; }
	}

	.animate-pulse {
		animation: pulse 1s ease-in-out infinite;
	}
</style>
