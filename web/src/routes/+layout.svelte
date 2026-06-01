<script lang="ts">
	import '../routes/layout.css';
	import { Globe, LayoutDashboard, LogOut } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import LogoMark from '$lib/LogoMark.svelte';

	let { children } = $props();
	let isLoggedIn = $state(false);

	onMount(() => {
		const token = localStorage.getItem('token');
		isLoggedIn = !!token;
		if (token) void refreshUser(token);
	});

	async function refreshUser(token: string) {
		try {
			const res = await fetch('/api/me', {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (!res.ok) return;
			const user = await res.json();
			localStorage.setItem('user', JSON.stringify(user));
		} catch {
			// Keep navigation usable even if refreshing user details fails.
		}
	}

	function handleLogout() {
		localStorage.removeItem('token');
		localStorage.removeItem('user');
		isLoggedIn = false;
		goto(resolve('/'));
	}
</script>

<div class="app-shell flex min-h-screen flex-col text-brand-light transition-colors">
	<!-- Navbar -->
	<header
		class="sticky top-0 z-50 border-b border-brand-light/80 bg-brand-ink/90 px-3 backdrop-blur-xl sm:px-4"
	>
		<nav class="container mx-auto flex h-[60px] items-center justify-between gap-3 px-1 sm:px-2">
			<a href={resolve('/')} class="flex items-center gap-2 text-xl font-bold tracking-tight">
				<LogoMark class="h-8 w-8" title="isitdead home" />
				<span class="hidden min-[360px]:inline">isitdead</span>
			</a>

			<div class="flex items-center gap-3 sm:gap-6">
				{#if isLoggedIn}
					<a
						href={resolve('/dashboard')}
						class="flex items-center gap-2 text-sm font-medium hover:text-brand-primary"
						aria-label="Dashboard"
					>
						<LayoutDashboard class="h-4 w-4" />
						<span class="hidden sm:inline">Dashboard</span>
					</a>
					<a
						href={resolve('/pricing')}
						class="hidden text-sm font-medium hover:text-brand-primary md:block"
					>
						Pricing
					</a>
					<div class="h-4 w-px bg-brand-light/10"></div>
					<button
						onclick={handleLogout}
						class="flex items-center gap-2 text-sm font-medium text-brand-accent hover:opacity-80"
						aria-label="Log out"
					>
						<LogOut class="h-4 w-4" />
						<span class="hidden sm:inline">Log out</span>
					</button>
				{:else}
					<a
						href={resolve('/#features')}
						class="hidden text-sm font-medium hover:text-brand-primary md:block"
					>
						Features
					</a>
					<a
						href={resolve('/pricing')}
						class="hidden text-sm font-medium hover:text-brand-primary md:block"
					>
						Pricing
					</a>
					<a
						href={resolve('/#how-it-works')}
						class="hidden text-sm font-medium hover:text-brand-primary md:block"
					>
						How it works
					</a>
					<div class="h-4 w-px bg-brand-light/10 md:block"></div>
					<a href={resolve('/login')} class="text-sm font-medium hover:text-brand-primary">
						Log in
					</a>
					<a
						href={resolve('/register')}
						class="inline-flex h-9 items-center justify-center rounded-xl bg-gradient-to-br from-brand-primary to-[#f0989e] px-4 py-2 text-sm font-black text-white shadow-lg shadow-brand-primary/20 transition hover:-translate-y-0.5 hover:brightness-110"
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
	<footer class="mt-8 border-t border-brand-light/40 bg-brand-ink py-10 sm:py-12">
		<div class="container mx-auto px-4 sm:px-6">
			<div
				class="terminal-panel flex flex-col items-center justify-between gap-6 px-5 py-6 md:flex-row md:px-7"
			>
				<div class="flex items-center gap-2">
					<LogoMark class="h-7 w-7" title="isitdead" />
					<span class="font-bold">isitdead</span>
				</div>
				<p class="text-center text-sm text-brand-light/60">
					&copy; 2026 isitdead. Uptime monitoring for websites, APIs, and services.
				</p>
				<div class="flex items-center gap-4">
					<a href={resolve('/')} class="text-brand-light/40 hover:text-brand-light">
						<Globe class="h-5 w-5" />
					</a>
				</div>
			</div>
		</div>
	</footer>
</div>
