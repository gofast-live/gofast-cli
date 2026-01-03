<script>
	/**
	 * @type {{
	 *   stackSummary: string,
	 *   adminUrl?: string,
	 *   discordUrl?: string,
	 *   v1Url?: string
	 * }}
	 */
	let {
		stackSummary = 'Go + ConnectRPC',
		adminUrl = '#',
		discordUrl = '#',
		v1Url = 'https://gofast.live'
	} = $props();
</script>

<section class="py-16 px-4">
	<div class="max-w-xl mx-auto text-center">
		<!-- Stack summary -->
		<p class="text-[var(--text-muted)] text-sm mb-2">Your stack:</p>
		<h2 class="text-2xl md:text-3xl font-semibold text-[var(--text)] mb-2">
			{stackSummary}
		</h2>
		<p class="text-[var(--primary)] mb-8">Ready for production.</p>

		<!-- Pricing card -->
		<div class="bg-[var(--surface)] border border-[var(--border)] rounded-xl p-6 mb-8">
			<div class="text-3xl font-bold text-[var(--text)] mb-1">$40</div>
			<div class="text-[var(--text-muted)] text-sm mb-6">one-time payment</div>

			<div class="text-left space-y-2 mb-6">
				<div class="flex items-center gap-2 text-sm">
					<span class="text-[var(--success)]">✓</span>
					<span class="text-[var(--text)]">GoFast V2 (this CLI)</span>
				</div>
				<div class="flex items-center gap-2 text-sm">
					<span class="text-[var(--success)]">✓</span>
					<span class="text-[var(--text)]">GoFast V1 (Next.js, Vue, AWS, GCP...)</span>
				</div>
				<div class="flex items-center gap-2 text-sm">
					<span class="text-[var(--success)]">✓</span>
					<span class="text-[var(--text)]">Lifetime access + updates</span>
				</div>
			</div>

			<a
				href={adminUrl}
				class="block w-full py-3 px-6 bg-[var(--primary)] hover:bg-[var(--primary-hover)] text-white font-medium rounded-lg text-center transition-colors"
			>
				Get Access →
			</a>
		</div>

		<!-- Discord -->
		<div class="mb-8">
			<p class="text-[var(--text-muted)] text-sm mb-3">Join the community</p>
			<a
				href={discordUrl}
				target="_blank"
				rel="noopener noreferrer"
				class="inline-flex items-center gap-2 px-4 py-2 bg-[var(--surface)] border border-[var(--border)] rounded-lg text-[var(--text)] hover:border-[var(--primary)] transition-colors"
			>
				<svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
					<path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z"/>
				</svg>
				<span>Discord</span>
				<span class="text-[var(--text-muted)] text-sm">100+ devs</span>
			</a>
			<p class="text-[var(--text-subtle)] text-xs mt-2">Free access, no purchase required</p>
		</div>

		<!-- V1 link -->
		<div class="text-sm text-[var(--text-muted)]">
			<p>Want more options? V1 has Next.js, Vue, AWS S3, GCP...</p>
			<a
				href={v1Url}
				target="_blank"
				rel="noopener noreferrer"
				class="text-[var(--primary)] hover:underline"
			>
				Visit gofast.live →
			</a>
		</div>
	</div>
</section>
