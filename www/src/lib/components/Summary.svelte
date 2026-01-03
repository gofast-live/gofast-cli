<script>
	import { state as appState } from '$lib/stores/state.svelte.js';
	import { fade, fly } from 'svelte/transition';
</script>

<section class="min-h-screen flex flex-col items-center justify-center p-6 snap-start text-center max-w-4xl mx-auto">
	<div in:fade={{ duration: 800, delay: 200 }}>
		<h2 class="text-3xl md:text-5xl font-bold mb-6 text-white">Your stack is ready.</h2>
		<p class="text-xl text-muted mb-12">
			Production-ready. Type-safe. Deployable.
		</p>

		<!-- Stack Summary -->
		<div class="bg-surface border border-border rounded-xl p-8 mb-12 text-left shadow-2xl relative overflow-hidden">
			<div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-primary to-blue-500"></div>
			
			<h3 class="text-sm font-mono text-muted uppercase tracking-wider mb-4">What you built</h3>
			
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
				<div class="flex items-center gap-2 text-gray-300">
					<span class="text-success">✓</span> Init (Go + ConnectRPC)
				</div>
				<div class="flex items-center gap-2 text-gray-300">
					<span class="text-success">✓</span> Auth (OAuth2)
				</div>
				
				{#if appState.models.length > 0}
					<div class="flex items-center gap-2 text-gray-300">
						<span class="text-success">✓</span> Models: <span class="text-white font-mono">{appState.models.join(', ')}</span>
					</div>
				{/if}

				{#if appState.has('client')}
					<div class="flex items-center gap-2 text-gray-300">
						<span class="text-success">✓</span> Client (SvelteKit)
					</div>
				{/if}

				{#if appState.has('stripe')}
					<div class="flex items-center gap-2 text-gray-300">
						<span class="text-success">✓</span> Stripe Payments
					</div>
				{/if}

				{#if appState.has('r2')}
					<div class="flex items-center gap-2 text-gray-300">
						<span class="text-success">✓</span> R2 Storage
					</div>
				{/if}
				
				{#if appState.has('postmark')}
					<div class="flex items-center gap-2 text-gray-300">
						<span class="text-success">✓</span> Postmark Email
					</div>
				{/if}

				{#if appState.has('infra')}
					<div class="flex items-center gap-2 text-gray-300">
						<span class="text-success">✓</span> Infrastructure (K8s/TF)
					</div>
				{/if}
			</div>

			<div class="pt-6 border-t border-border flex flex-col md:flex-row justify-between items-center gap-4">
				<div class="font-mono text-sm text-muted">
					Stack: {appState.stackSummary}
				</div>
				<!-- <div class="text-xs text-primary bg-primary/10 px-2 py-1 rounded">
					Ready to deploy
				</div> -->
			</div>
		</div>

		<!-- Pricing / CTA -->
		<div class="flex flex-col md:flex-row gap-6 justify-center items-stretch">
			<div class="flex-1 bg-surface/50 border border-border p-6 rounded-xl flex flex-col items-center justify-center hover:border-primary/30 transition-colors">
				<div class="text-2xl font-bold text-white mb-2">$40 <span class="text-sm font-normal text-muted">/ one-time</span></div>
				<ul class="text-sm text-muted space-y-2 mb-6 text-left">
					<li class="flex gap-2"><span class="text-primary">✓</span> GoFast V2 (CLI)</li>
					<li class="flex gap-2"><span class="text-primary">✓</span> GoFast V1 (Next.js/Vue)</li>
					<li class="flex gap-2"><span class="text-primary">✓</span> Lifetime updates</li>
				</ul>
				<a href="https://gofast.live/buy" class="w-full py-3 bg-primary hover:bg-primary-hover text-white rounded-lg font-bold transition-colors">
					Get Access
				</a>
			</div>

			<div class="flex-1 bg-transparent border border-border p-6 rounded-xl flex flex-col items-center justify-center hover:border-white/20 transition-colors">
				<div class="text-xl font-bold text-white mb-2">Community</div>
				<p class="text-sm text-muted mb-6">
					Join 100+ developers building with GoFast. Free access.
				</p>
				<a href="https://discord.gg/gofast" target="_blank" class="w-full py-3 bg-surface hover:bg-surface-hover border border-border text-white rounded-lg font-medium transition-colors">
					Join Discord
				</a>
			</div>
		</div>

		<div class="mt-12 text-sm text-muted">
			Want more options (Next.js, Vue, AWS)? <a href="https://gofast.live" class="underline hover:text-white">Check out V1</a>
		</div>
	</div>
</section>