<script>
    import { projects } from '$lib/data/projects.js';
    import { ExternalLink } from '@lucide/svelte';
</script>

<section class="py-24 px-6 border-t border-border/50 bg-bg/50">
    <div class="max-w-7xl mx-auto">
        <div class="text-center mb-16">
            <h2 class="text-3xl md:text-5xl font-bold mb-4 text-white">
                Built with GoFast
            </h2>
            <p class="text-xl text-muted">
                See what others are shipping to production.
            </p>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {#each projects as p}
                <a 
                    href={p.url} 
                    target="_blank" 
                    rel="noopener noreferrer"
                    class="group bg-surface/30 border border-border p-6 rounded-xl backdrop-blur-sm hover:border-primary/30 hover:bg-surface/50 transition-all flex flex-col h-full relative overflow-hidden"
                >
                    <div class="absolute top-4 right-4 opacity-0 group-hover:opacity-100 transition-opacity text-muted group-hover:text-white">
                        <ExternalLink size={16} />
                    </div>

                    <div class="flex items-center gap-4 mb-4">
                        <div class="w-12 h-12 rounded-lg {p.color} flex items-center justify-center font-bold text-xl shadow-inner">
                            {p.title[0]}
                        </div>
                        <div>
                            <div class="font-bold text-white text-lg group-hover:text-primary transition-colors">{p.title}</div>
                            <div class="text-xs text-muted font-mono">{new URL(p.url).hostname}</div>
                        </div>
                    </div>
                    
                    <p class="text-gray-400 text-sm leading-relaxed">
                        {p.description}
                    </p>
                </a>
            {/each}
            
            <!-- Add Your Project Card -->
            <a 
                href="https://discord.com/invite/EdSZbQbRyJ" 
                target="_blank" 
                rel="noopener noreferrer"
                class="group bg-transparent border border-dashed border-border p-6 rounded-xl hover:border-primary/50 hover:bg-primary/5 transition-all flex flex-col items-center justify-center h-full text-center"
            >
                <div class="w-12 h-12 rounded-full bg-surface border border-border flex items-center justify-center mb-4 text-muted group-hover:text-primary group-hover:border-primary transition-colors">
                    +
                </div>
                <div class="font-bold text-white text-sm mb-1">Add your project</div>
                <div class="text-xs text-muted">Join the community showcase</div>
            </a>
        </div>
    </div>
</section>
