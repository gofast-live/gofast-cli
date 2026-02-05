<script>
    const testimonials = [
        {
            name: "busfactor",
            role: "@busf4ctor",
            quote: "Hey! Yesterday I used GoFast and it's amazing. I even learnt some nice Golang ways of doing things out of it. This price for a lifetime access is a bargain! I am sure it will improve a lot and I am anxious for that. I love how easy it is to integrate OAuth, the directories organization, the clean code, etc. Amazing product.",
        },
        {
            name: "prasanna513",
            role: "@prasanna5135661",
            quote: "Thanks for making this wonderful foundation starter kit. Got me up and running fast. This was the tech stack I was looking for: SvelteKit+Svelte; Go; PostgreSQL. Very well written and organized code. Way way more worth than the lifetime membership fee paid.",
        },
        {
            name: "silenta salmans",
            role: "@five.hundred.cigarettes",
            quote: "GoFast is the nicest thing I've came across in some time. you run the cli and get clean boilerplate, sparing you all the hassle of setting everything up so you can start executing your ideas right away. It allows you to pre-configure it with many options, select your db, preferred storage options, emails and so on, you can tune it however you want.",
        },
        {
            name: "nadia",
            role: "@nadia__1756",
            quote: "Using GoFast has been a great experience. The CLI is intuitive, and the support team is incredibly responsive. The generated code is clean and well-structured, providing a great foundation for building new apps.",
        },
        {
            name: "Frendic",
            role: "@frendicc",
            quote: "GoFast changed how I kick off new projects. The fact that you can pre-configure things like database and storage is a huge timesaver. The code is super clean and well organized, which I really appreciate. It’s great for just jumping straight into the building phase without messing around with setup. And the Grafana monitoring is fck amazing.",
        },
        {
            name: "Milen Todev",
            role: "@milentodev",
            quote: "One of the best projects ever made! Real guide and foundation for building a solid software, not just another 150$+ full of vulnerabilities fancy landing page boilerplate!",
        },
        {
            name: "Safaros",
            role: "@safaros",
            quote: "Just gotta say I bought this a few months ago and loved the idea but dismantled it to use my preferred libraries (like sqlc, atlas, HTMX) and gave up. Trying it after seeing all the updated and this has come a very very long way. Getting re-familiarized with the project and at this point I think it's everything I wanted this to be!",
        },
    ];
</script>

<section class="py-24 px-6 border-t border-border/50 bg-bg/50">
    <div class="max-w-7xl mx-auto">
        <div class="text-center mb-16">
            <h2 class="text-3xl md:text-5xl font-bold mb-4 text-white">
                Loved by developers
            </h2>
            <p class="text-xl text-muted">
                Join hundreds of developers building faster with GoFast.
            </p>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {#each testimonials as t}
                <div
                    class="bg-surface/30 border border-border p-6 rounded-xl backdrop-blur-sm hover:border-primary/30 transition-colors flex flex-col h-full"
                >
                    <div class="flex-1 mb-6">
                        <p class="text-gray-300 leading-relaxed text-sm">
                            "{t.quote}"
                        </p>
                    </div>
                    <div class="flex items-center gap-3 mt-auto">
                        <div
                            class="w-10 h-10 rounded-full bg-primary/20 text-primary flex items-center justify-center font-bold text-lg shrink-0"
                        >
                            {t.name[0].toUpperCase()}
                        </div>
                        <div class="overflow-hidden">
                            <div class="font-bold text-white text-sm truncate">
                                {t.name}
                            </div>
                            <div class="text-xs text-muted truncate">
                                {t.role}
                            </div>
                        </div>
                    </div>
                </div>
            {/each}
        </div>
    </div>
</section>
