<script lang="ts">
	import { Activity, Mail, Lock, User, ArrowRight } from 'lucide-svelte';

	let username = $state('');
	let email = $state('');
	let password = $state('');
	let isLoading = $state(false);
	let message = $state('');

	async function handleRegister(e: SubmitEvent) {
		e.preventDefault();
		isLoading = true;
		message = '';

		try {
			const res = await fetch('/api/register', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ username, email, password })
			});

			if (res.ok) {
				message = 'Registration successful! Redirecting...';
				// In a real app, we'd redirect to the dashboard
			} else {
				const data = await res.json();
				message = data.error || 'Registration failed. Please try again.';
			}
		} catch (err) {
			message = 'Connection error. Is the backend running?';
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="flex min-h-[calc(100vh-16rem)] items-center justify-center px-4 py-12">
	<div class="w-full max-w-md space-y-8 rounded-3xl border border-brand-light/10 bg-brand-dark p-8 shadow-xl md:p-10">
		<div class="text-center">
			<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-primary text-brand-dark">
				<Activity class="h-6 w-6" />
			</div>
			<h2 class="mt-6 text-3xl font-extrabold tracking-tight">Create your account</h2>
			<p class="mt-2 text-sm text-brand-light/60">
				Start monitoring your services in less than a minute.
			</p>
		</div>

		<form class="mt-8 space-y-6" onsubmit={handleRegister}>
			<div class="space-y-4">
				<div>
					<label for="username" class="block text-sm font-medium text-brand-light/80">Username</label>
					<div class="relative mt-1">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-brand-light/40">
							<User class="h-5 w-5" />
						</div>
						<input
							id="username"
							name="username"
							type="text"
							required
							bind:value={username}
							class="block w-full rounded-xl border border-brand-light/10 bg-brand-light/5 py-3 pl-10 pr-3 leading-5 placeholder-brand-light/30 focus:border-brand-primary focus:bg-brand-light/10 focus:outline-none focus:ring-1 focus:ring-brand-primary"
							placeholder="johndoe"
						/>
					</div>
				</div>

				<div>
					<label for="email" class="block text-sm font-medium text-brand-light/80">Email address</label>
					<div class="relative mt-1">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-brand-light/40">
							<Mail class="h-5 w-5" />
						</div>
						<input
							id="email"
							name="email"
							type="email"
							autocomplete="email"
							required
							bind:value={email}
							class="block w-full rounded-xl border border-brand-light/10 bg-brand-light/5 py-3 pl-10 pr-3 leading-5 placeholder-brand-light/30 focus:border-brand-primary focus:bg-brand-light/10 focus:outline-none focus:ring-1 focus:ring-brand-primary"
							placeholder="john@example.com"
						/>
					</div>
				</div>

				<div>
					<label for="password" class="block text-sm font-medium text-brand-light/80">Password</label>
					<div class="relative mt-1">
						<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-brand-light/40">
							<Lock class="h-5 w-5" />
						</div>
						<input
							id="password"
							name="password"
							type="password"
							autocomplete="new-password"
							required
							bind:value={password}
							class="block w-full rounded-xl border border-brand-light/10 bg-brand-light/5 py-3 pl-10 pr-3 leading-5 placeholder-brand-light/30 focus:border-brand-primary focus:bg-brand-light/10 focus:outline-none focus:ring-1 focus:ring-brand-primary"
							placeholder="••••••••"
						/>
					</div>
				</div>
			</div>

			{#if message}
				<div class="rounded-lg bg-brand-primary/10 p-4 text-sm text-brand-primary">
					{message}
				</div>
			{/if}

			<div>
				<button
					type="submit"
					disabled={isLoading}
					class="group relative flex w-full justify-center rounded-xl bg-brand-primary px-4 py-3 text-sm font-semibold text-brand-dark transition-all hover:bg-brand-primary/90 focus:outline-none focus:ring-2 focus:ring-brand-primary focus:ring-offset-2 disabled:opacity-50"
				>
					{#if isLoading}
						<svg class="h-5 w-5 animate-spin text-brand-dark" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
						</svg>
					{:else}
						Sign up
						<ArrowRight class="ml-2 h-5 w-5 transition-transform group-hover:translate-x-1" />
					{/if}
				</button>
			</div>
		</form>

		<div class="mt-6 text-center text-sm">
			<span class="text-brand-light/60">Already have an account?</span>
			<a href="/login" class="ml-1 font-semibold text-brand-primary hover:text-brand-primary/80">Log in</a>
		</div>
	</div>
</div>
