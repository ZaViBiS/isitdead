<script lang="ts">
	import { onMount } from 'svelte';
	import { ExternalLink, Globe2, RefreshCw, Save } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import type { Server } from '$lib/utils';

	let servers = $state<Server[]>([]);
	let isLoading = $state(true);
	let error = $state('');
	let saving = $state<number | null>(null);

	function slugify(value: string) {
		return value
			.trim()
			.toLowerCase()
			.replace(/[^a-z0-9]+/g, '-')
			.replace(/^-+|-+$/g, '');
	}

	async function loadServers() {
		const token = localStorage.getItem('token');
		if (!token) {
			goto(resolve('/login'));
			return;
		}

		try {
			const res = await fetch('/api/servers', {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (!res.ok) {
				error = 'Failed to load monitors';
				return;
			}
			servers = (await res.json()) as Server[];
		} catch {
			error = 'Connection error';
		} finally {
			isLoading = false;
		}
	}

	async function savePublic(server: Server) {
		const token = localStorage.getItem('token');
		saving = server.id;
		error = '';

		try {
			const res = await fetch(`/api/admin/servers/${server.id}/public`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({
					public: server.public,
					public_slug: server.public_slug || slugify(server.name)
				})
			});
			if (!res.ok) {
				const data = await res.json().catch(() => ({ error: 'Failed to save public page' }));
				error = data.error || 'Failed to save public page';
				return;
			}
			const updated = (await res.json()) as Server;
			const idx = servers.findIndex((s) => s.id === updated.id);
			if (idx !== -1) servers[idx] = updated;
		} catch {
			error = 'Connection error';
		} finally {
			saving = null;
		}
	}

	onMount(loadServers);
</script>

<div class="container mx-auto max-w-6xl px-4 py-10 sm:px-6 lg:py-12">
	<section class="mb-8">
		<div
			class="mb-4 inline-flex items-center gap-2 rounded-full border border-brand-primary/20 bg-brand-primary/10 px-3 py-1.5 text-xs font-black text-brand-primary uppercase"
		>
			<Globe2 class="h-4 w-4" />
			Admin
		</div>
		<h1 class="text-3xl font-black text-brand-light sm:text-5xl">Public pages</h1>
		<p class="mt-3 max-w-2xl text-sm leading-6 text-brand-light/45 sm:text-base">
			Publish curated monitor pages for search traffic. Normal users cannot enable these pages from the dashboard.
		</p>
	</section>

	{#if error}
		<div class="mb-6 rounded-2xl border border-brand-accent/20 bg-brand-accent/10 p-4 text-brand-accent">
			{error}
		</div>
	{/if}

	{#if isLoading}
		<div class="flex min-h-80 items-center justify-center rounded-[2rem] border border-brand-light/10 bg-brand-light/[0.025]">
			<RefreshCw class="h-9 w-9 animate-spin text-brand-primary" />
		</div>
	{:else}
		<section class="grid gap-4">
			{#each servers as server (server.id)}
				<article class="rounded-[1.5rem] border border-brand-light/10 bg-[#111f1c]/90 p-4 sm:p-5">
					<div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center">
						<div class="min-w-0">
							<div class="flex flex-wrap items-center gap-2">
								<h2 class="truncate text-lg font-black">{server.name}</h2>
								<span class="rounded-lg border border-brand-light/10 bg-brand-light/[0.04] px-2 py-0.5 text-[10px] font-black text-brand-light/45 uppercase">
									{server.check_type}
								</span>
							</div>
							<p class="mt-2 truncate text-sm text-brand-light/35">{server.url}</p>
						</div>

						<div class="grid gap-3 sm:grid-cols-[auto_minmax(16rem,1fr)_auto] sm:items-center">
							<label class="flex cursor-pointer items-center gap-3 rounded-xl border border-brand-light/10 bg-brand-light/[0.03] px-4 py-3">
								<input type="checkbox" bind:checked={server.public} class="h-5 w-5 accent-brand-primary" />
								<span class="text-sm font-bold text-brand-light/75">Public</span>
							</label>
							<input
								type="text"
								value={server.public_slug || slugify(server.name)}
								oninput={(e) => (server.public_slug = slugify(e.currentTarget.value))}
								class="w-full rounded-xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 text-sm outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
								placeholder="wikipedia-down"
							/>
							<div class="flex gap-2">
								{#if server.public && server.public_slug}
									<a
										href={`/status/${server.public_slug}`}
										target="_blank"
										class="inline-flex h-11 w-11 items-center justify-center rounded-xl bg-brand-light/[0.06] text-brand-light/45 transition hover:bg-brand-light/10 hover:text-brand-primary"
										title="Open public page"
									>
										<ExternalLink class="h-4 w-4" />
									</a>
								{/if}
								<button
									onclick={() => savePublic(server)}
									class="inline-flex h-11 items-center justify-center gap-2 rounded-xl bg-brand-primary px-4 font-black text-brand-dark transition hover:bg-brand-primary/90 disabled:opacity-50"
									disabled={saving === server.id}
								>
									{#if saving === server.id}
										<RefreshCw class="h-4 w-4 animate-spin" />
									{:else}
										<Save class="h-4 w-4" />
									{/if}
									Save
								</button>
							</div>
						</div>
					</div>
				</article>
			{/each}
		</section>
	{/if}
</div>
