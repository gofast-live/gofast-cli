<script>
    import { state as appState } from "$lib/stores/state.svelte.js";
    import { fade, fly, scale } from "svelte/transition";
    import { onMount } from "svelte";

    let mounted = $state(false);
    let step = $state(0);
    let animating = $state(false);

    const hasMon = $derived(appState.has("mon"));

    // The observability pipeline
    const pipeline = [
        {
            id: "request",
            label: "Request",
            icon: "arrow-right",
            color: "text-white"
        },
        {
            id: "otel",
            label: "OpenTelemetry",
            icon: "code",
            color: "text-blue-400"
        },
        {
            id: "alloy",
            label: "Alloy",
            icon: "funnel",
            color: "text-purple-400"
        }
    ];

    const backends = [
        {
            id: "tempo",
            label: "Tempo",
            type: "Traces",
            color: "text-red-400",
            bg: "bg-red-400/10 border-red-400/30"
        },
        {
            id: "loki",
            label: "Loki",
            type: "Logs",
            color: "text-amber-400",
            bg: "bg-amber-400/10 border-amber-400/30"
        },
        {
            id: "prometheus",
            label: "Prometheus",
            type: "Metrics",
            color: "text-orange-400",
            bg: "bg-orange-400/10 border-orange-400/30"
        }
    ];

    onMount(() => {
        mounted = true;
        startAnimation();
    });

    function startAnimation() {
        if (animating) return;
        animating = true;
        step = 0;

        const interval = setInterval(() => {
            step++;
            // 0: request, 1: otel, 2: alloy, 3: backends fan out, 4: grafana
            if (step > 4) {
                clearInterval(interval);
                setTimeout(() => {
                    animating = false;
                    setTimeout(startAnimation, 1000);
                }, 1500);
            }
        }, 600);
    }
</script>

<section
    class="flex flex-col items-center justify-center p-6 py-24 text-center max-w-6xl mx-auto"
>
    {#if mounted}
        <div in:fade={{ duration: 600 }} class="w-full">
            <div class="mb-4 text-cyan-400 font-mono text-sm tracking-wider">
                OBSERVABILITY
            </div>
            <h2 class="text-3xl md:text-4xl font-bold mb-4 text-white">
                Full OpenTelemetry. Zero config.
            </h2>
            <p class="text-lg text-muted mb-12 max-w-2xl mx-auto">
                Every request traced. Every log correlated. Every metric
                recorded. Click a trace, see the logs.
            </p>

            <!-- Pipeline visualization -->
            <div
                class="relative bg-surface/30 border border-border rounded-2xl p-8 mb-8 overflow-hidden"
            >
                {#if !hasMon}
                    <div
                        class="absolute inset-0 bg-bg/80 backdrop-blur-sm z-10 flex flex-col items-center justify-center"
                    >
                        <div class="text-2xl mb-4">
                            <svg
                                class="w-12 h-12 text-muted mx-auto mb-4"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="1.5"
                                    d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                                />
                            </svg>
                        </div>
                        <p class="text-muted mb-4">
                            Add monitoring for full observability
                        </p>
                        <code
                            class="text-primary bg-surface px-4 py-2 rounded-lg font-mono text-sm"
                        >
                            gof mon
                        </code>
                    </div>
                {/if}

                <!-- Top row: Request → OTEL → Alloy -->
                <div class="flex items-center justify-center gap-4 mb-8">
                    {#each pipeline as node, i}
                        <div class="flex items-center gap-4">
                            <div
                                class="flex flex-col items-center transition-all duration-300 {step >
                                i
                                    ? 'opacity-100'
                                    : step === i
                                      ? 'opacity-100 scale-110'
                                      : 'opacity-30'}"
                            >
                                <div
                                    class="w-16 h-16 rounded-full border-2 flex items-center justify-center transition-all duration-300 {step >
                                    i
                                        ? 'border-success bg-success/20'
                                        : step === i
                                          ? 'border-current animate-pulse'
                                          : 'border-border bg-surface'} {node.color}"
                                >
                                    {#if node.icon === "arrow-right"}
                                        <svg
                                            class="w-6 h-6"
                                            fill="none"
                                            stroke="currentColor"
                                            viewBox="0 0 24 24"
                                        >
                                            <path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="2"
                                                d="M14 5l7 7m0 0l-7 7m7-7H3"
                                            />
                                        </svg>
                                    {:else if node.icon === "code"}
                                        <svg
                                            class="w-6 h-6"
                                            fill="none"
                                            stroke="currentColor"
                                            viewBox="0 0 24 24"
                                        >
                                            <path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="2"
                                                d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
                                            />
                                        </svg>
                                    {:else if node.icon === "funnel"}
                                        <svg
                                            class="w-6 h-6"
                                            fill="none"
                                            stroke="currentColor"
                                            viewBox="0 0 24 24"
                                        >
                                            <path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="2"
                                                d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
                                            />
                                        </svg>
                                    {/if}
                                </div>
                                <span
                                    class="text-xs font-medium mt-2 {node.color}"
                                    >{node.label}</span
                                >
                            </div>

                            {#if i < pipeline.length - 1}
                                <div
                                    class="w-8 h-0.5 transition-all duration-300 {step >
                                    i
                                        ? 'bg-success'
                                        : 'bg-border'}"
                                ></div>
                            {/if}
                        </div>
                    {/each}
                </div>

                <!-- Fan out lines -->
                <div
                    class="flex justify-center mb-4 transition-opacity duration-300 {step >=
                    3
                        ? 'opacity-100'
                        : 'opacity-0'}"
                >
                    <svg class="w-64 h-8" viewBox="0 0 256 32">
                        <path
                            d="M128 0 L64 32"
                            stroke="currentColor"
                            stroke-width="2"
                            fill="none"
                            class="text-red-400/50"
                        />
                        <path
                            d="M128 0 L128 32"
                            stroke="currentColor"
                            stroke-width="2"
                            fill="none"
                            class="text-amber-400/50"
                        />
                        <path
                            d="M128 0 L192 32"
                            stroke="currentColor"
                            stroke-width="2"
                            fill="none"
                            class="text-orange-400/50"
                        />
                    </svg>
                </div>

                <!-- Backend row: Tempo | Loki | Prometheus -->
                <div class="flex justify-center gap-6 mb-8">
                    {#each backends as backend, i}
                        <div
                            class="flex flex-col items-center p-4 rounded-xl border transition-all duration-300 {step >=
                            3
                                ? backend.bg + ' opacity-100'
                                : 'border-border bg-surface opacity-30'}"
                            in:fly={{ y: 10, duration: 300, delay: i * 100 }}
                        >
                            <span class="font-bold {backend.color}"
                                >{backend.label}</span
                            >
                            <span class="text-xs text-muted">{backend.type}</span
                            >
                        </div>
                    {/each}
                </div>

                <!-- Converge to Grafana -->
                <div
                    class="flex justify-center mb-4 transition-opacity duration-300 {step >=
                    4
                        ? 'opacity-100'
                        : 'opacity-0'}"
                >
                    <svg class="w-64 h-8" viewBox="0 0 256 32">
                        <path
                            d="M64 0 L128 32"
                            stroke="currentColor"
                            stroke-width="2"
                            fill="none"
                            class="text-green-400/50"
                        />
                        <path
                            d="M128 0 L128 32"
                            stroke="currentColor"
                            stroke-width="2"
                            fill="none"
                            class="text-green-400/50"
                        />
                        <path
                            d="M192 0 L128 32"
                            stroke="currentColor"
                            stroke-width="2"
                            fill="none"
                            class="text-green-400/50"
                        />
                    </svg>
                </div>

                <!-- Grafana -->
                <div
                    class="flex justify-center transition-all duration-500 {step >=
                    4
                        ? 'opacity-100 scale-100'
                        : 'opacity-30 scale-95'}"
                >
                    <div
                        class="flex flex-col items-center p-6 rounded-xl border transition-all duration-300 {step >=
                        4
                            ? 'bg-green-400/10 border-green-400/30'
                            : 'border-border bg-surface'}"
                    >
                        <div class="flex items-center gap-2 mb-2">
                            <svg
                                class="w-8 h-8 text-green-400"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                                />
                            </svg>
                            <span class="font-bold text-lg text-green-400"
                                >Grafana</span
                            >
                        </div>
                        <span class="text-xs text-muted"
                            >Dashboards + Correlation</span
                        >
                    </div>
                </div>

                <!-- Correlation callout -->
                <div
                    class="mt-8 pt-6 border-t border-border transition-all duration-300 {step >= 4 ? 'opacity-100 blur-0' : 'opacity-50 blur-sm'}"
                >
                    <div
                        class="flex flex-wrap justify-center gap-6 text-sm"
                    >
                        <div class="flex items-center gap-2">
                            <span class="text-red-400">Trace</span>
                            <span class="text-muted">→</span>
                            <span class="text-amber-400">Logs</span>
                            <span class="text-success text-xs"
                                >(1-click)</span
                            >
                        </div>
                        <div class="flex items-center gap-2">
                            <span class="text-orange-400">Metric</span>
                            <span class="text-muted">→</span>
                            <span class="text-red-400">Exemplar</span>
                            <span class="text-muted">→</span>
                            <span class="text-red-400">Trace</span>
                        </div>
                    </div>
                </div>
            </div>

            <!-- What's included -->
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4 max-w-2xl mx-auto">
                {#each [{ name: "RED Metrics", desc: "Rate, Errors, Duration" }, { name: "Service Graph", desc: "Auto-mapped dependencies" }, { name: "Exemplars", desc: "Metrics → Traces" }, { name: "Pre-built Dashboards", desc: "Ready to use" }] as feature, i}
                    <div
                        class="bg-surface/50 border border-border rounded-xl p-3 text-left hover:border-cyan-400/30 transition-colors {!hasMon
                            ? 'opacity-50'
                            : ''}"
                        in:fly={{ y: 20, duration: 400, delay: 100 * i }}
                    >
                        <div class="text-cyan-400 font-medium text-sm">
                            {feature.name}
                        </div>
                        <div class="text-muted text-xs mt-1">{feature.desc}</div>
                    </div>
                {/each}
            </div>

            {#if hasMon}
                <p
                    class="text-success mt-8 flex items-center justify-center gap-2"
                    in:fade
                >
                    <svg
                        class="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M5 13l4 4L19 7"
                        />
                    </svg>
                    Monitoring included in your stack
                </p>
            {/if}
        </div>
    {/if}
</section>
