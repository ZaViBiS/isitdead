<script lang="ts">
	import '../routes/layout.css';
	import { Activity, ShieldCheck, Globe, LayoutDashboard, LogOut } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	let { children } = $props();
	let isLoggedIn = $state(false);

	onMount(() => {
		isLoggedIn = !!localStorage.getItem('token');
	});

	function handleLogout() {
		localStorage.removeItem('token');
		localStorage.removeItem('user');
		isLoggedIn = false;
		goto('/');
	}
</script>

<div class="flex min-h-screen flex-col bg-brand-dark text-brand-light transition-colors">
	<!-- Navbar -->
	<header class="sticky top-0 z-50 border-b border-brand-light/10 bg-brand-dark/80 backdrop-blur-md">
		<nav class="container mx-auto flex h-16 items-center justify-between px-4 sm:px-6">
			<a href="/" class="flex items-center gap-2 text-xl font-bold tracking-tight">
				<Activity class="h-6 w-6 text-brand-primary" />
				<span>isitdead</span>
			</a>

			<div class="flex items-center gap-4 sm:gap-6">
				{#if isLoggedIn}
					<a href="/dashboard" class="flex items-center gap-2 text-sm font-medium hover:text-brand-primary">
						<LayoutDashboard class="h-4 w-4" />
						Dashboard
					</a>
					<div class="h-4 w-px bg-brand-light/10"></div>
					<button
						onclick={handleLogout}
						class="flex items-center gap-2 text-sm font-medium text-brand-accent hover:opacity-80"
					>
						<LogOut class="h-4 w-4" />
						Log out
					</button>
				{:else}
					<a href="/" class="hidden text-sm font-medium hover:text-brand-primary md:block">
						Features
					</a>
					<a href="/" class="hidden text-sm font-medium hover:text-brand-primary md:block">
						Pricing
					</a>
					<div class="h-4 w-px bg-brand-light/10 md:block"></div>
					<a href="/login" class="text-sm font-medium hover:text-brand-primary">
						Log in
					</a>
					<a
						href="/register"
						class="inline-flex h-9 items-center justify-center rounded-full bg-brand-primary px-4 py-2 text-sm font-medium text-brand-dark shadow-sm transition-colors hover:bg-brand-primary/90 focus:outline-none focus:ring-2 focus:ring-brand-primary focus:ring-offset-2"
					>
						Sign up
					</a>
				{/if}
			</div>
		</nav>
	</header>

	<!-- Main Content -->
	<main class="flex-1">
		{@render children()}
	</main>

	<!-- Footer -->
	<footer class="border-t border-brand-light/10 py-12">
		<div class="container mx-auto px-4 sm:px-6">
			<div class="flex flex-col items-center justify-between gap-6 md:flex-row">
				<div class="flex items-center gap-2">
					<Activity class="h-5 w-5 text-brand-primary" />
					<span class="font-bold">isitdead</span>
				</div>
				<p class="text-center text-sm text-brand-light/60">
					&copy; 2026 isitdead. All rights reserved. Built with SvelteKit & Go.
				</p>
				<div class="flex items-center gap-4">
					<a href="/" class="text-brand-light/40 hover:text-brand-light">
						<Globe class="h-5 w-5" />
					</a>
				</div>
			</div>
		</div>
	</footer>
</div>
