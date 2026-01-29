<script>
    import { state as appState } from "$lib/stores/state.svelte.js";
    import { fade, fly, scale } from "svelte/transition";
    import { onMount } from "svelte";

    let mounted = $state(false);
    let step = $state(0);
    let animating = $state(false);

    const hasModel = $derived(appState.models.length > 0);
    const modelName = $derived(appState.models[0] || "task");

    const allGroups = [
        {
            layer: "Database",
            color: "text-blue-400",
            files: [
                { path: "storage/migrations/", file: "NNNN_create_{model}s.sql" },
                { path: "storage/", file: "query.sql", note: "+insert, +get, +list, +update, +delete" }
            ]
        },
        {
            layer: "Proto",
            color: "text-purple-400",
            files: [
                { path: "proto/v1/", file: "{model}.proto" },
                { path: "gen/proto/", file: "{model}.pb.go" }
            ]
        },
        {
            layer: "Domain",
            color: "text-green-400",
            files: [
                { path: "domain/{model}/", file: "service.go" },
                { path: "domain/{model}/", file: "service_test.go" },
                { path: "domain/{model}/", file: "validation.go" },
                { path: "domain/{model}/", file: "validation_test.go" }
            ]
        },
        {
            layer: "Transport",
            color: "text-orange-400",
            files: [
                { path: "transport/{model}/", file: "route.go" },
                { path: "transport/{model}/", file: "route_test.go" }
            ]
        },
        {
            layer: "Auth",
            color: "text-red-400",
            files: [
                { path: "pkg/auth/", file: "auth.go", note: "+Get, +Create, +Edit, +Remove flags" }
            ]
        },
        {
            layer: "Frontend",
            color: "text-amber-400",
            files: [
                { path: "src/routes/(app)/{model}s/", file: "+page.svelte" },
                { path: "src/routes/(app)/{model}s/[id]/", file: "+page.svelte" }
            ]
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
            if (step >= allGroups.length) {
                clearInterval(interval);
                setTimeout(() => {
                    animating = false;
                    setTimeout(startAnimation, 1000);
                }, 1500);
            }
        }, 600);
    }

    /**
     * @param {string} template
     * @param {string} model
     */
    function replaceName(template, model) {
        return template.replace(/\{model\}/g, model);
    }
</script>

<section
    class="flex flex-col items-center justify-center p-6 py-24 text-center max-w-6xl mx-auto"
>
    {#if mounted}
        <div in:fade={{ duration: 600 }} class="w-full">
            <div class="mb-4 text-purple-400 font-mono text-sm tracking-wider">
                MODEL GENERATION
            </div>
            <h2 class="text-3xl md:text-4xl font-bold mb-4 text-white">
                One command. Full stack.
            </h2>
            <p class="text-lg text-muted mb-8 max-w-2xl mx-auto">
                Define your model once. Get database, API, business logic, and UI — all type-safe, all connected.
            </p>

            <!-- Command display -->
            <div class="mb-8">
                <code
                    class="text-primary bg-surface px-6 py-3 rounded-lg font-mono text-lg inline-block"
                >
                    gof model {modelName} name:string done:bool
                </code>
            </div>

            <!-- File tree animation -->
            <div
                class="relative bg-surface/30 border border-border rounded-2xl p-8 mb-8 overflow-hidden text-left max-w-2xl mx-auto"
            >
                {#if !hasModel}
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
                                    d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                                />
                            </svg>
                        </div>
                        <p class="text-muted mb-4">
                            Add a model to see the magic
                        </p>
                        <code
                            class="text-primary bg-surface px-4 py-2 rounded-lg font-mono text-sm"
                        >
                            gof model task
                        </code>
                    </div>
                {/if}

                <!-- Layer by layer file generation -->
                <div class="space-y-4">
                    {#each allGroups as group, i}
                        <div
                            class="transition-all duration-300 {step > i
                                ? 'opacity-100'
                                : step === i
                                  ? 'opacity-100'
                                  : 'opacity-30'}"
                        >
                            <!-- Layer header -->
                            <div
                                class="flex items-center gap-2 mb-2 {group.color}"
                            >
                                {#if step > i}
                                    <span
                                        class="w-5 h-5 bg-success rounded-full flex items-center justify-center"
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
                                    </span>
                                {:else if step === i}
                                    <span
                                        class="w-5 h-5 border-2 border-current rounded-full animate-pulse"
                                    ></span>
                                {:else}
                                    <span
                                        class="w-5 h-5 border border-current/30 rounded-full"
                                    ></span>
                                {/if}
                                <span class="font-bold text-sm"
                                    >{group.layer}</span
                                >
                            </div>

                            <!-- Files in this layer -->
                            <div class="ml-7 space-y-1">
                                {#each group.files as file}
                                    <div
                                        class="flex items-center gap-2 text-sm font-mono"
                                    >
                                        <span class="text-muted"
                                            >{replaceName(
                                                file.path,
                                                modelName
                                            )}</span
                                        >
                                        <span
                                            class="text-white {step >= i
                                                ? ''
                                                : 'opacity-50'}"
                                        >
                                            {replaceName(file.file, modelName)}
                                        </span>
                                        {#if file.note && step > i}
                                            <span
                                                class="text-xs text-success"
                                                in:fade={{ duration: 200 }}
                                                >{file.note}</span
                                            >
                                        {/if}
                                    </div>
                                {/each}
                            </div>
                        </div>
                    {/each}
                </div>

                <!-- Summary -->
                <div
                    class="mt-6 pt-4 border-t border-border flex items-center justify-center gap-4 text-sm"
                >
                    <span class="text-muted"
                        >{allGroups.reduce(
                            (acc, g) => acc + g.files.length,
                            0
                        )} files</span
                    >
                    <span class="text-border">|</span>
                    <span class="text-muted"
                        >{allGroups.length} layers</span
                    >
                    <span class="text-border">|</span>
                    <span class="text-success">Type-safe end-to-end</span>
                </div>
            </div>

            <!-- Type safety callout -->
            <div
                class="flex flex-wrap justify-center gap-4 text-sm text-muted"
            >
                <div class="flex items-center gap-2">
                    <span class="text-blue-400">SQL</span>
                    <span>→</span>
                    <span class="text-purple-400">Proto</span>
                    <span>→</span>
                    <span class="text-green-400">Go</span>
                    <span>→</span>
                    <span class="text-amber-400">TypeScript</span>
                </div>
                <span class="text-success">Zero runtime type errors</span>
            </div>

            {#if hasModel}
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
                    {appState.models.length} model{appState.models.length > 1
                        ? "s"
                        : ""} in your stack
                </p>
            {/if}
        </div>
    {/if}
</section>
