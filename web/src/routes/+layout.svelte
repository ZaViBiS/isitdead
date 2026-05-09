<script lang="ts">
	import '../routes/layout.css';
	import { Activity, ShieldCheck, Globe } from 'lucide-svelte';

	let { children } = $props();
</script>

<div class="flex min-h-screen flex-col bg-slate-50 text-slate-900 transition-colors dark:bg-slate-950 dark:text-slate-50">
	<!-- Navbar -->
	<header class="sticky top-0 z-50 border-b border-slate-200 bg-white/80 backdrop-blur-md dark:border-slate-800 dark:bg-slate-950/80">
		<nav class="container mx-auto flex h-16 items-center justify-between px-4 sm:px-6">
			<a href="/" class="flex items-center gap-2 text-xl font-bold tracking-tight">
				<Activity class="h-6 w-6 text-indigo-600 dark:text-indigo-400" />
				<span>isitdead</span>
			</a>

			<div class="flex items-center gap-4 sm:gap-6">
				<a href="/" class="hidden text-sm font-medium hover:text-indigo-600 dark:hover:text-indigo-400 md:block">
					Features
				</a>
				<a href="/" class="hidden text-sm font-medium hover:text-indigo-600 dark:hover:text-indigo-400 md:block">
					Pricing
				</a>
				<div class="h-4 w-px bg-slate-200 dark:bg-slate-800 md:block"></div>
				<a href="/" class="text-sm font-medium hover:text-indigo-600 dark:hover:text-indigo-400">
					Log in
				</a>
				<a
					href="/register"
					class="inline-flex h-9 items-center justify-center rounded-full bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
				>
					Sign up
				</a>
			</div>
		</nav>
	</header>

	<!-- Main Content -->
	<main class="flex-1">
		{@render children()}
	</main>

	<!-- Footer -->
	<footer class="border-t border-slate-200 py-12 dark:border-slate-800">
		<div class="container mx-auto px-4 sm:px-6">
			<div class="flex flex-col items-center justify-between gap-6 md:flex-row">
				<div class="flex items-center gap-2">
					<Activity class="h-5 w-5 text-indigo-600" />
					<span class="font-bold">isitdead</span>
				</div>
				<p class="text-center text-sm text-slate-500 dark:text-slate-400">
					&copy; 2026 isitdead. All rights reserved. Built with SvelteKit & Go.
				</p>
				<div class="flex items-center gap-4">
					<a href="/" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200">
						<Globe class="h-5 w-5" />
					</a>
				</div>
			</div>
		</div>
	</footer>
</div>
