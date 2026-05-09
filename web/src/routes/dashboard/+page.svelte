<script lang="ts">
	import { onMount } from 'svelte';
	import { Activity, Plus, Trash2, ExternalLink, RefreshCw, AlertCircle } from 'lucide-svelte';
	import { goto } from '$app/navigation';

	interface Server {
		id: number;
		name: string;
		url: string;
		status: string;
		latency: number;
		check_interval: number;
	}

	let servers = $state<Server[]>([]);
	let isLoading = $state(true);
	let isAdding = $state(false);
	let error = $state('');

	// Form for adding new server
	let newName = $state('');
	let newUrl = $state('');
	let newInterval = $state(60);

	async function fetchServers() {
		const token = localStorage.getItem('token');
		if (!token) {
			goto('/login');
			return;
		}

		try {
			const res = await fetch('/api/servers', {
				headers: { Authorization: `Bearer ${token}` }
			});

			if (res.ok) {
				servers = await res.json();
			} else if (res.status === 401) {
				localStorage.removeItem('token');
				goto('/login');
			} else {
				error = 'Failed to fetch servers';
			}
		} catch (err) {
			error = 'Connection error';
		} finally {
			isLoading = false;
		}
	}

	async function addServer(e: SubmitEvent) {
		e.preventDefault();
		const token = localStorage.getItem('token');
		
		try {
			const res = await fetch('/api/servers', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`
				},
				body: JSON.stringify({
					name: newName,
					url: newUrl,
					check_interval: Number(newInterval)
				})
			});

			if (res.ok) {
				const newServer = await res.json();
				servers = [...servers, newServer];
				isAdding = false;
				newName = '';
				newUrl = '';
			}
		} catch (err) {
			error = 'Failed to add server';
		}
	}

	async function deleteServer(id: number) {
		const token = localStorage.getItem('token');
		try {
			const res = await fetch(`/api/servers/${id}`, {
				method: 'DELETE',
				headers: { Authorization: `Bearer ${token}` }
			});

			if (res.ok) {
				servers = servers.filter((s) => s.id !== id);
			}
		} catch (err) {
			error = 'Failed to delete server';
		}
	}

	onMount(fetchServers);
</script>

<div class="container mx-auto px-4 py-12">
	<div class="mb-8 flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold">Dashboard</h1>
			<p class="text-brand-light/60">Manage and monitor your services in real-time.</p>
		</div>
		<button
			onclick={() => (isAdding = !isAdding)}
			class="flex items-center gap-2 rounded-xl bg-brand-primary px-4 py-2 font-semibold text-brand-dark transition-all hover:bg-brand-primary/90"
		>
			<Plus class="h-5 w-5" />
			Add Server
		</button>
	</div>

	{#if error}
		<div class="mb-6 flex items-center gap-2 rounded-xl bg-brand-accent/10 p-4 text-brand-accent">
			<AlertCircle class="h-5 w-5" />
			{error}
		</div>
	{/if}

	<!-- Add Server Modal/Panel -->
	{#if isAdding}
		<div class="mb-8 rounded-2xl border border-brand-light/10 bg-brand-light/5 p-6 shadow-xl">
			<h2 class="mb-4 text-xl font-bold">Add New Monitor</h2>
			<form onsubmit={addServer} class="grid gap-4 md:grid-cols-4">
				<div class="space-y-1">
					<label for="name" class="text-xs font-medium text-brand-light/60 uppercase">Name</label>
					<input
						id="name"
						type="text"
						bind:value={newName}
						required
						placeholder="Production API"
						class="w-full rounded-xl border border-brand-light/10 bg-brand-dark px-4 py-2 focus:border-brand-primary focus:outline-none"
					/>
				</div>
				<div class="space-y-1">
					<label for="url" class="text-xs font-medium text-brand-light/60 uppercase">URL</label>
					<input
						id="url"
						type="url"
						bind:value={newUrl}
						required
						placeholder="https://api.example.com"
						class="w-full rounded-xl border border-brand-light/10 bg-brand-dark px-4 py-2 focus:border-brand-primary focus:outline-none"
					/>
				</div>
				<div class="space-y-1">
					<label for="interval" class="text-xs font-medium text-brand-light/60 uppercase">Interval (sec)</label>
					<select
						id="interval"
						bind:value={newInterval}
						class="w-full rounded-xl border border-brand-light/10 bg-brand-dark px-4 py-2 focus:border-brand-primary focus:outline-none"
					>
						<option value={30}>30s</option>
						<option value={60}>1m</option>
						<option value={300}>5m</option>
						<option value={600}>10m</option>
					</select>
				</div>
				<div class="flex items-end">
					<button
						type="submit"
						class="w-full rounded-xl bg-brand-primary py-2 font-semibold text-brand-dark hover:bg-brand-primary/90"
					>
						Save Monitor
					</button>
				</div>
			</form>
		</div>
	{/if}

	<!-- Servers List -->
	{#if isLoading}
		<div class="flex h-64 items-center justify-center">
			<RefreshCw class="h-8 w-8 animate-spin text-brand-primary" />
		</div>
	{:else if servers.length === 0}
		<div class="flex flex-col items-center justify-center rounded-3xl border border-dashed border-brand-light/20 py-20 text-center">
			<Activity class="mb-4 h-12 w-12 text-brand-light/20" />
			<h3 class="text-xl font-bold">No monitors yet</h3>
			<p class="text-brand-light/60">Add your first server to start monitoring.</p>
		</div>
	{:else}
		<div class="grid gap-4 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
			{#each servers as s (s.id)}
				<div class="group rounded-2xl border border-brand-light/10 bg-brand-dark p-6 transition-all hover:border-brand-primary/30 hover:shadow-lg">
					<div class="mb-4 flex items-start justify-between">
						<div>
							<h3 class="font-bold">{s.name}</h3>
							<p class="truncate text-sm text-brand-light/40">{s.url}</p>
						</div>
						<div class="flex gap-2">
							<a href={s.url} target="_blank" class="rounded-lg p-2 text-brand-light/40 hover:bg-brand-light/5 hover:text-brand-light">
								<ExternalLink class="h-4 w-4" />
							</a>
							<button 
								onclick={() => deleteServer(s.id)}
								class="rounded-lg p-2 text-brand-light/40 hover:bg-brand-accent/10 hover:text-brand-accent"
							>
								<Trash2 class="h-4 w-4" />
							</button>
						</div>
					</div>

					<div class="flex items-center justify-between rounded-xl bg-brand-light/5 p-4">
						<div class="flex items-center gap-2">
							<span class="h-2 w-2 rounded-full {s.status.startsWith('2') ? 'bg-brand-primary shadow-[0_0_8px_#73E2A7]' : s.status === 'unknown' ? 'bg-brand-soft shadow-[0_0_8px_#E3C0D3]' : 'bg-brand-accent shadow-[0_0_8px_#D62246]'}"></span>
							<span class="text-sm font-medium uppercase tracking-wider">
								{#if s.status.startsWith('2')}
									Online
								{:else if s.status === 'unknown'}
									Pending
								{:else}
									Error
								{/if}
							</span>
						</div>
						<div class="text-right">
							<div class="text-lg font-bold text-brand-primary">{s.latency}ms</div>
							<div class="text-[10px] text-brand-light/40 uppercase">Latency</div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
