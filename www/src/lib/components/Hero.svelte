<script>
	import Terminal from './Terminal.svelte';

	/**
	 * @type {{
	 *   onInitComplete?: () => void,
	 *   discordUrl?: string,
	 *   v1Url?: string
	 * }}
	 */
	let {
		onInitComplete = () => {},
		discordUrl = '#',
		v1Url = 'https://gofast.live'
	} = $props();

	/** @type {any} */
	let terminalRef = $state(null);
	let hasStarted = $state(false);

	const initCommand = 'gof init myproject';
	const initDescription = 'Creating project structure...';
	const initOutputs = [
		{ text: 'OAuth (GitHub + Google)' },
		{ text: 'Bitwise role authorization' },
		{ text: 'Docker Compose setup' },
		{ text: 'GitHub Actions CI/CD' },
		{ text: 'PR preview deployments' },
		{ text: 'PostgreSQL + SQLC' },
		{ text: 'ConnectRPC transport' }
	];

	function handleStart() {
		if (hasStarted) return;
		hasStarted = true;
		terminalRef?.run();
	}

	function handleComplete() {
		onInitComplete();
	}
</script>

<section class="min-h-screen flex flex-col">
	<!-- Header -->
	<header class="flex justify-end items-center gap-4 p-4 md:p-6">
		<a
			href={discordUrl}
			target="_blank"
			rel="noopener noreferrer"
			class="flex items-center gap-2 text-[var(--text-muted)] hover:text-[var(--text)] transition-colors"
		>
			<svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
				<path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z"/>
			</svg>
			<span class="hidden sm:inline text-sm">Discord</span>
		</a>
		<a
			href={v1Url}
			target="_blank"
			rel="noopener noreferrer"
			class="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors"
		>
			V1 →
		</a>
	</header>

	<!-- Main content -->
	<div class="flex-1 flex flex-col items-center justify-center px-4 -mt-16">
		<!-- Logo -->
		<img src="/logo.svg" alt="GoFast" class="h-10 md:h-12 mb-4" />

		<!-- Tagline -->
		<h1 class="text-xl md:text-2xl text-[var(--text-muted)] mb-12">
			Building blocks for Go
		</h1>

		<!-- Terminal -->
		<div class="relative">
			<Terminal
				bind:this={terminalRef}
				command={initCommand}
				description={initDescription}
				outputs={initOutputs}
				onComplete={handleComplete}
			/>

			<!-- Play button overlay -->
			{#if !hasStarted}
				<button
					onclick={handleStart}
					aria-label="Run init command"
					class="absolute inset-0 flex items-center justify-center bg-black/50 rounded-lg cursor-pointer group"
				>
					<div class="w-16 h-16 rounded-full bg-[var(--primary)] flex items-center justify-center group-hover:scale-110 transition-transform">
						<svg class="w-6 h-6 text-white ml-1" fill="currentColor" viewBox="0 0 24 24">
							<path d="M8 5v14l11-7z"/>
						</svg>
					</div>
				</button>
			{/if}
		</div>
	</div>
</section>
