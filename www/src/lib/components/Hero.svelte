<script>
	import { state as appState } from '$lib/stores/state.svelte.js';
	import { gsap } from '$lib/animations/gsap.js';

	/** @type {{ onStart: () => void }} */
	let { onStart } = $props();

	function handleStart() {
		appState.init();
		onStart();
	}
</script>

<section class="min-h-screen flex flex-col items-center justify-center relative p-6 snap-start">
	<!-- Top Nav -->
	<nav class="absolute top-0 left-0 w-full p-6 flex justify-between items-center text-sm text-muted">
		<div class="flex items-center gap-3">
			<img src="/logo.svg" alt="GoFast Logo" class="w-8 h-8 object-contain" />
			<div class="font-mono font-bold text-white tracking-tight">GoFast</div>
		</div>
		<div class="flex gap-6">
			<a href="https://discord.gg/gofast" target="_blank" class="hover:text-white transition-colors">Discord</a>
			<a href="https://gofast.live" target="_blank" class="hover:text-white transition-colors">V1 (Next.js, Vue, HTMX...) →</a>
		</div>
	</nav>

	<!-- Main Content -->
	<div class="flex flex-col items-center text-center max-w-3xl w-full">
		<!-- Logo -->
		<div class="mb-8 w-24 h-24 bg-surface rounded-2xl flex items-center justify-center border border-border shadow-2xl p-4">
			<img src="/logo.svg" alt="GoFast Logo" class="w-full h-full object-contain" />
		</div>

		<h1 class="text-4xl md:text-6xl font-bold tracking-tight mb-4 text-white">
			Building blocks for <span class="text-primary">Go</span>
		</h1>
		
		<p class="text-xl text-muted mb-12 max-w-xl">
			Generate production-ready Go apps with ConnectRPC, SvelteKit, and PostgreSQL.
		</p>

		<!-- Stack Badges -->
		<div class="flex flex-wrap justify-center gap-3 mb-16 font-mono text-sm text-text-subtle">
			<span class="px-3 py-1 bg-surface border border-border rounded-full">Go 1.23+</span>
			<span class="px-3 py-1 bg-surface border border-border rounded-full">ConnectRPC</span>
			<span class="px-3 py-1 bg-surface border border-border rounded-full">SvelteKit</span>
			<span class="px-3 py-1 bg-surface border border-border rounded-full">SQLC</span>
		</div>

		<!-- Interactive Terminal Trigger -->
		<div class="w-full max-w-xl">
			<button 
				onclick={handleStart}
				class="group w-full bg-surface border border-border hover:border-primary/50 transition-all duration-300 rounded-xl p-4 md:p-6 text-left font-mono text-lg md:text-xl shadow-lg relative overflow-hidden"
			>
				<div class="absolute inset-0 bg-primary/5 opacity-0 group-hover:opacity-100 transition-opacity"></div>
				<div class="flex justify-between items-center relative z-10">
					<span class="text-white">
						<span class="text-primary mr-2">$</span>gof init myproject
					</span>
					<span class="flex items-center justify-center w-10 h-10 rounded-full bg-primary/10 text-primary group-hover:bg-primary group-hover:text-white transition-all">
						<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>
					</span>
				</div>
			</button>
			<div class="mt-4 text-sm text-muted animate-pulse">
				Click to start building
			</div>
		</div>
	</div>
</section>