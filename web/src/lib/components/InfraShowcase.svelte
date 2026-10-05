<script>
    import { state as appState } from "$lib/stores/state.svelte.js";
    import { fade, fly, scale } from "svelte/transition";
    import { onMount } from "svelte";

    let mounted = $state(false);
    let step = $state(0);
    let animating = $state(false);

    const hasInfra = $derived(appState.has("infra"));

    const steps = [
        { label: "PR #42 opened", icon: "git-pr", delay: 0 },
        { label: "GitHub Action triggered", icon: "github", delay: 600 },
        { label: "Lint + Go tests", icon: "check", delay: 1200 },
        { label: "Stack up in Docker", icon: "docker", delay: 1800 },
        { label: "Playwright E2E", icon: "globe", delay: 2400 },
        { label: "Results on the PR", icon: "comment", delay: 3000 }
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
            if (step >= steps.length) {
                clearInterval(interval);
                // Reset after pause
                setTimeout(() => {
                    animating = false;
                    setTimeout(startAnimation, 1000);
                }, 1500);
            }
        }, 600);
    }

    const setupScripts = [
        {
            name: "./setup_rke2.sh",
            desc: "K8s cluster bootstrap"
        },
        {
            name: "./setup_app.sh",
            desc: "App deployment setup"
        }
    ];
</script>

<section
    class="flex flex-col items-center justify-center p-6 py-24 text-center max-w-6xl mx-auto"
>
    {#if mounted}
        <div in:fade={{ duration: 600 }} class="w-full">
            <div class="mb-4 text-primary font-mono text-sm tracking-wider">
                INFRASTRUCTURE
            </div>
            <h2 class="text-3xl md:text-4xl font-bold mb-4 text-white">
                Every PR Fully Tested
            </h2>
            <p class="text-lg text-muted mb-12 max-w-2xl mx-auto">
                Every pull request runs lints, Go tests and Playwright E2E against the full stack. <br
                />Inside the GitHub runner. No preview cluster to keep alive.
            </p>

            <!-- PR Flow Animation -->
            <div
                class="relative bg-surface/30 border border-border rounded-2xl p-8 mb-12 overflow-hidden"
            >
                {#if !hasInfra}
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
                                    d="M12 15v2m0 0v2m0-2h2m-2 0H9m3-4V8a3 3 0 00-3-3H6a3 3 0 00-3 3v4a3 3 0 003 3h3m6-3a3 3 0 013 3v4a3 3 0 01-3 3h-3"
                                />
                            </svg>
                        </div>
                        <p class="text-muted mb-4">
                            Add infrastructure to unlock PR checks
                        </p>
                        <code
                            class="text-primary bg-surface px-4 py-2 rounded-lg font-mono text-sm"
                        >
                            gof infra
                        </code>
                    </div>
                {/if}

                <!-- Flow visualization -->
                <div class="flex flex-col md:flex-row items-center justify-between gap-4 md:gap-2">
                    {#each steps as s, i}
                        <div class="flex items-center gap-2 md:gap-0 md:flex-col">
                            <!-- Node -->
                            <div
                                class="relative w-16 h-16 rounded-full border-2 flex items-center justify-center transition-all duration-300 {step >
                                i
                                    ? 'border-primary bg-primary/20 text-primary'
                                    : step === i
                                      ? 'border-primary bg-primary/10 text-primary animate-pulse'
                                      : 'border-border bg-surface text-muted'}"
                            >
                                {#if s.icon === "git-pr"}
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
                                            d="M8 7v8a2 2 0 002 2h6M8 7V5a2 2 0 012-2h4.586a1 1 0 01.707.293l4.414 4.414a1 1 0 01.293.707V15a2 2 0 01-2 2h-2M8 7H6a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2v-2"
                                        />
                                    </svg>
                                {:else if s.icon === "github"}
                                    <svg
                                        class="w-6 h-6"
                                        fill="currentColor"
                                        viewBox="0 0 24 24"
                                    >
                                        <path
                                            d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"
                                        />
                                    </svg>
                                {:else if s.icon === "docker"}
                                    <svg
                                        class="w-6 h-6"
                                        fill="currentColor"
                                        viewBox="0 0 24 24"
                                    >
                                        <path
                                            d="M13.983 11.078h2.119a.186.186 0 00.186-.185V9.006a.186.186 0 00-.186-.186h-2.119a.185.185 0 00-.185.185v1.888c0 .102.083.185.185.185m-2.954-5.43h2.118a.186.186 0 00.186-.186V3.574a.186.186 0 00-.186-.185h-2.118a.185.185 0 00-.185.185v1.888c0 .102.082.185.185.186m0 2.716h2.118a.187.187 0 00.186-.186V6.29a.186.186 0 00-.186-.185h-2.118a.185.185 0 00-.185.185v1.887c0 .102.082.185.185.186m-2.93 0h2.12a.186.186 0 00.184-.186V6.29a.185.185 0 00-.185-.185H8.1a.185.185 0 00-.185.185v1.887c0 .102.083.185.185.186m-2.964 0h2.119a.186.186 0 00.185-.186V6.29a.185.185 0 00-.185-.185H5.136a.186.186 0 00-.186.185v1.887c0 .102.084.185.186.186m5.893 2.715h2.118a.186.186 0 00.186-.185V9.006a.186.186 0 00-.186-.186h-2.118a.185.185 0 00-.185.185v1.888c0 .102.082.185.185.185m-2.93 0h2.12a.185.185 0 00.184-.185V9.006a.185.185 0 00-.184-.186h-2.12a.185.185 0 00-.184.185v1.888c0 .102.083.185.185.185m-2.964 0h2.119a.185.185 0 00.185-.185V9.006a.185.185 0 00-.185-.186h-2.119a.186.186 0 00-.186.186v1.887c0 .102.084.185.186.185m-2.92 0h2.12a.185.185 0 00.184-.185V9.006a.185.185 0 00-.184-.186h-2.12a.185.185 0 00-.184.185v1.888c0 .102.082.185.185.185M23.763 9.89c-.065-.051-.672-.51-1.954-.51-.338.001-.676.03-1.01.087-.248-1.7-1.653-2.53-1.716-2.566l-.344-.199-.226.327c-.284.438-.49.922-.612 1.43-.23.97-.09 1.882.403 2.661-.595.332-1.55.413-1.744.42H.751a.751.751 0 00-.75.748 11.376 11.376 0 00.692 4.062c.545 1.428 1.355 2.48 2.41 3.124 1.18.723 3.1 1.137 5.275 1.137.983.003 1.963-.086 2.93-.266a12.248 12.248 0 003.823-1.389c.98-.567 1.86-1.288 2.61-2.136 1.252-1.418 1.998-2.997 2.553-4.4h.221c1.372 0 2.215-.549 2.68-1.009.309-.293.55-.65.707-1.046l.098-.288Z"
                                        />
                                    </svg>
                                {:else if s.icon === "check"}
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
                                            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                                        />
                                    </svg>
                                {:else if s.icon === "comment"}
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
                                            d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"
                                        />
                                    </svg>
                                {:else if s.icon === "globe"}
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
                                            d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"
                                        />
                                    </svg>
                                {/if}

                                {#if step > i}
                                    <div
                                        class="absolute -top-1 -right-1 w-5 h-5 bg-success rounded-full flex items-center justify-center"
                                        in:scale={{ duration: 200 }}
                                    >
                                        <svg
                                            class="w-3 h-3 text-white"
                                            fill="none"
                                            stroke="currentColor"
                                            viewBox="0 0 24 24"
                                        >
                                            <path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="3"
                                                d="M5 13l4 4L19 7"
                                            />
                                        </svg>
                                    </div>
                                {/if}
                            </div>

                            <!-- Label -->
                            <div
                                class="mt-3 text-xs font-medium transition-colors duration-300 max-w-20 md:max-w-24 {step >=
                                i
                                    ? 'text-white'
                                    : 'text-muted'}"
                            >
                                {s.label}
                            </div>
                        </div>

                        <!-- Connector line -->
                        {#if i < steps.length - 1}
                            <div
                                class="hidden md:block flex-1 h-0.5 mx-2 transition-all duration-300 {step >
                                i
                                    ? 'bg-primary'
                                    : 'bg-border'}"
                            ></div>
                        {/if}
                    {/each}
                </div>

                <!-- PR comment result -->
                <div class="mt-8 pt-6 border-t border-border">
                    <div class="flex items-center justify-center gap-3">
                        <span class="text-success text-lg">E2E Tests Passed:</span>
                        <code
                            class="text-primary bg-primary/10 px-4 py-2 rounded-lg font-mono"
                        >
                            svelte: 39 passed, 0 failed
                        </code>
                    </div>
                    <p class="text-muted text-sm mt-2">
                        One run per frontend, nothing deployed
                    </p>
                </div>
            </div>

            <!-- Setup Scripts -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-2xl mx-auto">
                {#each setupScripts as script, i}
                    <div
                        class="bg-surface/50 border border-border rounded-xl p-4 text-left hover:border-primary/30 transition-colors {!hasInfra
                            ? 'opacity-50'
                            : ''}"
                        in:fly={{ y: 20, duration: 400, delay: 100 * i }}
                    >
                        <code class="text-primary font-mono text-sm"
                            >{script.name}</code
                        >
                        <p class="text-muted text-xs mt-1">{script.desc}</p>
                    </div>
                {/each}
            </div>

            {#if hasInfra}
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
                    Infrastructure included in your stack
                </p>
            {/if}
        </div>
    {/if}
</section>
