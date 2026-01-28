<script>
    import { state as appState } from "$lib/stores/state.svelte.js";
    import { buildWallData, sourceLabels, sourceColors } from "$lib/data/summary.js";
    import { fade, fly } from "svelte/transition";
    import { onMount } from "svelte";

    let mounted = $state(false);

    onMount(() => {
        mounted = true;
    });

    let wallData = $derived(buildWallData(appState));
</script>

<section
    class="flex flex-col items-center justify-center p-6 py-24 text-center max-w-7xl mx-auto min-h-screen"
>
    <div in:fade={{ duration: 800, delay: 200 }} class="w-full">
        <h2 class="text-3xl md:text-5xl font-bold mb-4 text-white">
            Your stack is ready.
        </h2>
        <p class="text-xl text-muted mb-12">
            Production-ready. Type-safe. Deployable.
        </p>

        <!-- The Wall of Value -->
        <div
            class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-16 text-left"
        >
            {#each wallData as section, i}
                <div
                    class="bg-surface/50 border border-border p-6 rounded-xl backdrop-blur-sm hover:border-primary/30 transition-colors"
                    in:fly={{ y: 20, duration: 500, delay: 300 + i * 100 }}
                >
                    <h3
                        class="text-primary font-mono font-bold mb-4 border-b border-border/50 pb-2"
                    >
                        {section.category}
                    </h3>
                    <ul class="space-y-2">
                        {#each section.items as item}
                            <li
                                class="text-sm text-gray-300 flex items-start gap-2"
                            >
                                <span class="text-success mt-0.5 shrink-0">✓</span>
                                <span class="flex-1">{item.text}</span>
                                {#if item.source}
                                    <span class="shrink-0 text-[10px] px-1.5 py-0.5 rounded font-mono {sourceColors[item.source] || 'bg-gray-500/20 text-gray-400'}">
                                        {sourceLabels[item.source] || item.source}
                                    </span>
                                {/if}
                            </li>
                        {/each}
                    </ul>
                </div>
            {/each}
        </div>

        <!-- Pricing / CTA -->
        <div
            class="flex flex-col md:flex-row gap-6 justify-center items-stretch max-w-4xl mx-auto"
        >
            <div
                class="flex-1 bg-surface border border-border p-8 rounded-xl flex flex-col items-center justify-center hover:border-primary/50 transition-all shadow-lg hover:shadow-primary/5 relative overflow-hidden group"
            >
                <div
                    class="absolute inset-0 bg-gradient-to-br from-primary/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"
                ></div>

                <div class="relative z-10 w-full">
                    <div class="text-3xl font-bold text-white mb-2 text-center">
                        $40 <span class="text-sm font-normal text-muted"
                            >/ one-time</span
                        >
                    </div>
                    <div
                        class="text-center text-sm text-primary mb-6 font-medium"
                    >
                        Lifetime Access
                    </div>

                    <ul
                        class="text-sm text-muted space-y-3 mb-8 text-left max-w-[240px] mx-auto"
                    >
                        <li class="flex gap-2 items-center">
                            <span
                                class="text-primary bg-primary/10 rounded-full p-0.5 w-5 h-5 flex items-center justify-center text-xs"
                                >✓</span
                            > GoFast V2 (This CLI)
                        </li>
                        <li class="flex gap-2 items-center">
                            <span
                                class="text-primary bg-primary/10 rounded-full p-0.5 w-5 h-5 flex items-center justify-center text-xs"
                                >✓</span
                            > GoFast V1 (Next.js, Vue...)
                        </li>
                        <li class="flex gap-2 items-center">
                            <span
                                class="text-primary bg-primary/10 rounded-full p-0.5 w-5 h-5 flex items-center justify-center text-xs"
                                >✓</span
                            > All future updates
                        </li>
                    </ul>

                    <a
                        href="https://admin.gofast.live"
                        class="block w-full py-4 bg-primary hover:bg-primary-hover text-white rounded-lg font-bold transition-all transform hover:-translate-y-0.5 text-center shadow-lg shadow-primary/20"
                    >
                        Get Access Now
                    </a>
                </div>
            </div>

            <div
                class="flex-1 bg-transparent border border-border p-8 rounded-xl flex flex-col items-center justify-center hover:border-white/20 transition-colors"
            >
                <div class="text-2xl font-bold text-white mb-2">Community</div>
                <p class="text-sm text-muted mb-8 text-center">
                    Join 100+ developers building with GoFast. <br />Open source
                    discussion & support.
                </p>
                <a
                    href="https://discord.com/invite/EdSZbQbRyJ"
                    target="_blank"
                    class="w-full py-4 bg-surface hover:bg-surface-hover border border-border text-white rounded-lg font-medium transition-colors text-center flex items-center justify-center gap-2"
                >
                    <!-- Discord Icon -->
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="20"
                        height="20"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        class="lucide lucide-message-circle"
                        ><path d="m3 21 1.9-5.7a8.5 8.5 0 1 1 3.8 3.8z" /></svg
                    >
                    Join Discord (Free)
                </a>
            </div>
        </div>

        <div class="mt-16 pt-8 border-t border-border/30 text-sm text-muted">
            <p>
                Need a different stack? <a
                    href="https://gofast.live"
                    target="_blank"
                    class="text-gray-400 hover:text-white underline decoration-gray-700 underline-offset-4"
                    >V1 includes Next.js, Vue, HTMX, AWS & GCP support.</a
                >
            </p>
        </div>
    </div>
</section>
