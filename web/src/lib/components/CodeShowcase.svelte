<script>
    import { fade } from 'svelte/transition';
    import showcaseData from '$lib/data/showcase-code.json';

    let activeTab = $state(showcaseData[0].id);
    let activeFile = $derived(showcaseData.find(f => f.id === activeTab));
</script>

<section class="py-24 px-6 relative overflow-hidden">
    <!-- Background Glow -->
    <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[600px] bg-primary/5 blur-[120px] rounded-full pointer-events-none"></div>

    <div class="max-w-7xl mx-auto relative z-10">
        <div class="text-center mb-16">
            <h2 class="text-3xl md:text-5xl font-bold mb-4 text-white">
                It's just Go.
            </h2>
            <p class="text-xl text-muted max-w-2xl mx-auto">
                No hidden runtimes or complex abstractions. Just a solid foundation of idiomatic code using the standard library, ConnectRPC, SQLC, and Goose.
            </p>
        </div>

        <div class="flex flex-col lg:grid lg:grid-cols-[300px_1fr] gap-0 bg-surface/30 border border-border rounded-xl backdrop-blur-sm overflow-hidden h-[800px] lg:h-[600px]">
            <!-- Sidebar / Tabs -->
            <div class="border-b lg:border-b-0 lg:border-r border-border flex flex-col bg-surface/50 h-[200px] lg:h-auto shrink-0">
                <div class="p-4 border-b border-border">
                    <div class="text-xs font-mono text-muted uppercase tracking-wider">Project Files</div>
                </div>
                <div class="flex-1 overflow-y-auto">
                    {#each showcaseData as file}
                        <button 
                            class="w-full text-left p-4 hover:bg-white/5 transition-colors border-l-2 relative group {activeTab === file.id ? 'border-primary bg-white/5' : 'border-transparent'}"
                            onclick={() => activeTab = file.id}
                        >
                            <div class="font-bold text-sm mb-1 text-white group-hover:text-primary transition-colors">{file.title}</div>
                            <div class="text-xs text-muted truncate font-mono opacity-70">{file.path}</div>
                        </button>
                    {/each}
                </div>
                
                <!-- Description Box (Bottom of sidebar) -->
                <div class="hidden lg:block p-6 border-t border-border bg-bg/50">
                    <h3 class="font-bold text-white mb-2 text-sm">{activeFile.title}</h3>
                    <p class="text-sm text-muted leading-relaxed">
                        {activeFile.description}
                    </p>
                </div>
            </div>

            <!-- Code View -->
            <div class="overflow-hidden flex flex-col bg-[#0d1117]"> <!-- github-dark bg color -->
                <div class="flex items-center justify-between px-4 py-2 border-b border-border bg-[#0d1117]">
                    <div class="flex items-center gap-2">
                        <span class="w-3 h-3 rounded-full bg-red-500/50"></span>
                        <span class="w-3 h-3 rounded-full bg-yellow-500/50"></span>
                        <span class="w-3 h-3 rounded-full bg-green-500/50"></span>
                    </div>
                    <div class="text-xs font-mono text-muted">{activeFile.path}</div>
                </div>
                <div class="flex-1 overflow-auto custom-scrollbar p-6">
                    {#key activeFile.id}
                        <div in:fade={{ duration: 200 }}>
                            <!-- eslint-disable-next-line svelte/no-at-html-tags -->
                            {@html activeFile.html}
                        </div>
                    {/key}
                </div>
            </div>
        </div>
    </div>
</section>

<style>
    /* Custom Scrollbar for the code area */
    .custom-scrollbar::-webkit-scrollbar {
        width: 10px;
        height: 10px;
    }
    .custom-scrollbar::-webkit-scrollbar-track {
        background: #0d1117;
    }
    .custom-scrollbar::-webkit-scrollbar-thumb {
        background: #30363d; /* github-dark border color roughly */
        border-radius: 5px;
        border: 2px solid #0d1117;
    }
    .custom-scrollbar::-webkit-scrollbar-thumb:hover {
        background: #404751;
    }

    /* Shiki container styling override if needed */
    :global(pre.shiki) {
        background-color: transparent !important;
        margin: 0;
        padding: 0;
        font-family: 'JetBrains Mono', 'Fira Code', monospace; /* Ensure a nice font if available, fallback to monospace */
        font-size: 0.875rem;
        line-height: 1.5;
    }
</style>
