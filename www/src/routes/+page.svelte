<script>
	import { tick } from 'svelte';
	import Hero from '$lib/components/Hero.svelte';
	import CommandFlow from '$lib/components/CommandFlow.svelte';
	import CommandPicker from '$lib/components/CommandPicker.svelte';
	import ModelShowcase from '$lib/components/ModelShowcase.svelte';
	import InfraShowcase from '$lib/components/InfraShowcase.svelte';
	import MonitoringShowcase from '$lib/components/MonitoringShowcase.svelte';
	import CodeShowcase from '$lib/components/CodeShowcase.svelte';
	import ProjectShowcase from '$lib/components/ProjectShowcase.svelte';
	import Summary from '$lib/components/Summary.svelte';
	import Testimonials from '$lib/components/Testimonials.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import { state as appState } from '$lib/stores/state.svelte.js';

	/** @type {{ id: string, variant?: any }[]} */
	let history = $state([]);

	/** @type {HTMLElement | undefined} */
	let mainContainer = $state();

	/** @type {HTMLElement | undefined} */
	let showcaseSection = $state();

	async function scrollToBottom() {
		await tick();
		if (mainContainer) {
			// Find the last section
			const sections = mainContainer.querySelectorAll('section');
			const lastSection = sections[sections.length - 1];
			if (lastSection) {
                // Scroll so the new section is visible, but not necessarily at the very top if it fits
				lastSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
			}
		}
	}

	async function scrollToShowcase() {
		await tick();
		if (showcaseSection) {
			showcaseSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
		}
	}

	function handleStart() {
		history = [{ id: 'init' }];
		scrollToBottom();
	}

	/**
	 * @param {{ id: string, variant?: any }} cmd
	 */
	async function handleSelect(cmd) {
		if (cmd.id === 'finish') {
			// appState.finish() is called in CommandPicker
			// Scroll to showcase section
			scrollToShowcase();
			return;
		}

		if (cmd.id === 'model' && cmd.variant) {
			appState.addModel(cmd.variant.name);
		} else {
			appState.add(cmd.id);
		}

		history = [...history, cmd];
		scrollToBottom();
	}
</script>

<main 
	bind:this={mainContainer}
	class="w-full min-h-screen bg-bg text-text pb-24"
>
	<Hero onStart={handleStart} />

	{#each history as step, i (step)}
		<!-- We need a key to ensure components don't get reused weirdly if history changes -->
		<CommandFlow 
			commandId={step.id} 
			variant={step.variant} 
		>
			{#if i === history.length - 1 && !appState.finished}
				<CommandPicker onSelect={handleSelect} />
			{/if}
		</CommandFlow>
	{/each}

	{#if appState.finished}
		<div bind:this={showcaseSection}>
			<ModelShowcase />
		</div>
		<MonitoringShowcase />
		<InfraShowcase />
		<CodeShowcase />
		<Summary />
		<ProjectShowcase />
		<Testimonials />
        <Footer />
	{/if}
</main>

<style>
</style>
