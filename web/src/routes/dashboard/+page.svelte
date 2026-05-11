<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Activity,
		Plus,
		Trash2,
		ExternalLink,
		RefreshCw,
		AlertCircle,
		Clock,
		BarChart3,
		ShieldCheck,
		Settings,
		X
	} from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		getStatusColor,
		getHourlyBuckets,
		getFaviconUrl,
		calculateUptime,
		calculateAvgLatency,
		type Server,
		type CheckResult
	} from '$lib/utils';

	let servers = $state<Server[]>([]);
	let isLoading = $state(true);
	let isAdding = $state(false);
	let isEditing = $state(false);
	let error = $state('');

	let newName = $state('');
	let newUrl = $state('');
	let newType = $state('http');
	let newInterval = $state(60);

	let editingServer = $state<Server | null>(null);
	let editName = $state('');
	let editUrl = $state('');
	let editType = $state('http');
	let editInterval = $state(60);

	async function fetchServers() {
		const token = localStorage.getItem('token');
		if (!token) {
			goto(resolve('/login'));
			return;
		}

		try {
			const res = await fetch('/api/servers', {
				headers: { Authorization: `Bearer ${token}` }
			});

			if (res.ok) {
				const data = (await res.json()) as Server[];
				servers = data.map((s) => ({
					...s,
					history: [],
					history30d: []
				}));
				servers.forEach((s) => fetchHistory(s));
			} else if (res.status === 401) {
				localStorage.removeItem('token');
				goto(resolve('/login'));
			}
		} catch {
			error = 'Connection error';
		} finally {
			isLoading = false;
		}
	}

	async function fetchHistory(s: Server) {
		const token = localStorage.getItem('token');
		const url = `/api/servers/${s.id}/results?hours=720`;

		try {
			const res = await fetch(url, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				const data = await res.json();
				s.history30d = data;
				// Filter for 24h for the mini chart
				const dayAgo = new Date(Date.now() - 24 * 60 * 60 * 1000).getTime();
				s.history = data.filter((r: CheckResult) => new Date(r.created_at).getTime() > dayAgo);
			}
		} catch {
			console.error('Failed to fetch history');
		}
	}

	function normalizeMonitorUrl(url: string, checkType: string) {
		const trimmed = url.trim();
		if (checkType !== 'http' || trimmed === '') return trimmed;
		if (trimmed.startsWith('http://') || trimmed.startsWith('https://')) return trimmed;
		return `https://${trimmed}`;
	}

	function willNormalizeMonitorUrl(url: string, checkType: string) {
		const trimmed = url.trim();
		return (
			checkType === 'http' &&
			trimmed !== '' &&
			!trimmed.startsWith('http://') &&
			!trimmed.startsWith('https://')
		);
	}

	function isServerOnline(server: Server) {
		return server.status.startsWith('2') || server.status === 'Connected';
	}

	function getOverallUptime() {
		const histories = servers
			.map((server) => server.history30d || [])
			.filter((history) => history.length > 0);
		if (histories.length === 0) return 0;
		const total = histories.reduce((sum, history) => sum + calculateUptime(history), 0);
		return total / histories.length;
	}

	function getOverallLatency() {
		const histories = servers
			.map((server) => server.history30d || [])
			.filter((history) => history.length > 0);
		if (histories.length === 0) return 0;
		const total = histories.reduce((sum, history) => sum + calculateAvgLatency(history), 0);
		return Math.round(total / histories.length);
	}

	function getHealthyCount() {
		return servers.filter(isServerOnline).length;
	}

	async function addServer(e: SubmitEvent) {
		e.preventDefault();
		const token = localStorage.getItem('token');
		try {
			const res = await fetch('/api/servers', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({
					name: newName,
					url: normalizeMonitorUrl(newUrl, newType),
					check_type: newType,
					check_interval: Number(newInterval)
				})
			});

			if (res.ok) {
				const newSrv = await res.json();
				const server: Server = { ...newSrv, history: [], history30d: [] };
				servers.push(server);
				isAdding = false;
				newName = '';
				newUrl = '';
				newType = 'http';
				newInterval = 60;
				fetchHistory(server);
			}
		} catch {
			error = 'Failed to add server';
		}
	}

	function openEdit(server: Server) {
		editingServer = server;
		editName = server.name;
		editUrl = server.url;
		editType = server.check_type;
		editInterval = server.check_interval;
		isEditing = true;
	}

	async function updateServer(e: SubmitEvent) {
		e.preventDefault();
		if (!editingServer) return;
		const token = localStorage.getItem('token');
		try {
			const res = await fetch(`/api/servers/${editingServer.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({
					name: editName,
					url: normalizeMonitorUrl(editUrl, editType),
					check_type: editType,
					check_interval: Number(editInterval)
				})
			});

			if (res.ok) {
				const updated = await res.json();
				const idx = servers.findIndex((s) => s.id === updated.id);
				if (idx !== -1) {
					// Зберігаємо історію, яку ми вже завантажили
					servers[idx] = {
						...servers[idx],
						...updated,
						history: servers[idx].history,
						history30d: servers[idx].history30d
					};
				}
				isEditing = false;
				editingServer = null;
			}
		} catch {
			error = 'Failed to update server';
		}
	}

	async function deleteServer(id: number) {
		if (!confirm('Are you sure you want to delete this monitor?')) return;
		const token = localStorage.getItem('token');
		try {
			const res = await fetch(`/api/servers/${id}`, {
				method: 'DELETE',
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				const idx = servers.findIndex((s) => s.id === id);
				if (idx !== -1) servers.splice(idx, 1);
			}
		} catch {
			error = 'Failed to delete server';
		}
	}

	const intervalPresets = [
		30, 60, 180, 300, 600, 1800, 3600, 7200, 14400, 21600, 36000, 43200, 86400
	];

	function formatInterval(seconds: number) {
		if (seconds < 60) return `${seconds}s`;
		const m = Math.floor(seconds / 60);
		const s = seconds % 60;
		if (m < 60) return s > 0 ? `${m}m ${s}s` : `${m}m`;
		const h = Math.floor(m / 60);
		const mm = m % 60;
		return mm > 0 ? `${h}h ${mm}m` : `${h}h`;
	}

	onMount(fetchServers);
</script>

<div class="container mx-auto max-w-6xl px-4 py-12">
	<div class="mb-8 flex flex-col justify-between gap-6 md:flex-row md:items-end">
		<div>
			<h1 class="mb-2 text-4xl font-black tracking-tight">Your monitors</h1>
			<p class="flex items-center gap-2 text-brand-light/40">
				<ShieldCheck class="h-4 w-4 text-brand-primary" />
				Track uptime, response time, and recent failures in one place.
			</p>
		</div>
		<button
			onclick={() => (isAdding = !isAdding)}
			class="flex items-center justify-center gap-2 rounded-2xl bg-brand-primary px-6 py-3 font-bold text-brand-dark shadow-lg shadow-brand-primary/20 transition-all hover:scale-105 active:scale-95"
		>
			<Plus class="h-5 w-5" /> Add monitor
		</button>
	</div>

	{#if !isLoading && servers.length > 0}
		<div class="mb-10 grid gap-4 sm:grid-cols-3">
			<div class="rounded-[2rem] border border-brand-light/10 bg-brand-light/[0.03] p-5">
				<div class="mb-2 text-xs font-black tracking-widest text-brand-light/30 uppercase">
					Healthy now
				</div>
				<div class="text-3xl font-black text-brand-primary">
					{getHealthyCount()}<span class="text-sm text-brand-light/30">/{servers.length}</span>
				</div>
			</div>
			<div class="rounded-[2rem] border border-brand-light/10 bg-brand-light/[0.03] p-5">
				<div class="mb-2 text-xs font-black tracking-widest text-brand-light/30 uppercase">
					Average uptime
				</div>
				<div class="text-3xl font-black text-brand-primary">{getOverallUptime().toFixed(1)}%</div>
			</div>
			<div class="rounded-[2rem] border border-brand-light/10 bg-brand-light/[0.03] p-5">
				<div class="mb-2 text-xs font-black tracking-widest text-brand-light/30 uppercase">
					Average response
				</div>
				<div class="text-3xl font-black text-brand-light/80">
					{getOverallLatency()}<span class="ml-1 text-sm text-brand-light/30">ms</span>
				</div>
			</div>
		</div>
	{/if}

	{#if error}
		<div
			class="animate-in fade-in slide-in-from-top-4 mb-8 flex items-center gap-3 rounded-2xl border border-brand-accent/20 bg-brand-accent/10 p-4 text-brand-accent"
		>
			<AlertCircle class="h-5 w-5" />
			{error}
		</div>
	{/if}

	{#if isAdding}
		<div
			class="animate-in zoom-in-95 mb-12 rounded-3xl border border-brand-light/10 bg-brand-light/5 p-8 shadow-2xl backdrop-blur-sm duration-200"
		>
			<div class="mb-6 flex items-center gap-3">
				<div class="rounded-xl bg-brand-primary/10 p-2 text-brand-primary">
					<Activity class="h-6 w-6" />
				</div>
				<h2 class="text-2xl font-bold">Add monitor</h2>
			</div>
			<form onsubmit={addServer} class="grid gap-6 md:grid-cols-4">
				<div class="space-y-2">
					<label
						for="name"
						class="ml-1 text-xs font-bold tracking-widest text-brand-light/40 uppercase"
						>Monitor name</label
					>
					<input
						id="name"
						type="text"
						bind:value={newName}
						required
						class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/50 px-5 py-3 transition-all outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
						placeholder="Production API"
					/>
				</div>
				<div class="space-y-2">
					<label
						for="url"
						class="ml-1 text-xs font-bold tracking-widest text-brand-light/40 uppercase"
						>URL or host</label
					>
					<input
						id="url"
						type="text"
						bind:value={newUrl}
						required
						class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/50 px-5 py-3 transition-all outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
						placeholder={newType === 'http'
							? 'example.com or https://example.com'
							: 'example.com:80'}
					/>
					{#if willNormalizeMonitorUrl(newUrl, newType)}
						<p
							class="rounded-2xl border border-brand-primary/20 bg-brand-primary/10 px-4 py-3 text-sm font-medium text-brand-primary"
						>
							We will save this as <span class="font-black"
								>{normalizeMonitorUrl(newUrl, newType)}</span
							>.
						</p>
					{/if}
				</div>
				<div class="space-y-2">
					<label
						for="type"
						class="ml-1 text-xs font-bold tracking-widest text-brand-light/40 uppercase"
						>Check Type</label
					>
					<div class="relative">
						<select
							id="type"
							bind:value={newType}
							class="w-full cursor-pointer appearance-none rounded-2xl border border-brand-light/10 bg-brand-dark/50 px-5 py-3 transition-all outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
						>
							<option value="http">HTTP (GET)</option>
							<option value="ping">TCP port</option>
						</select>
					</div>
				</div>
				<div class="space-y-2 md:col-span-1">
					<div class="ml-1 flex items-center justify-between">
						<label
							for="interval"
							class="text-xs font-bold tracking-widest text-brand-light/40 uppercase"
							>Interval</label
						>
						<span
							class="rounded-full bg-brand-primary/10 px-2 py-0.5 text-xs font-black whitespace-nowrap text-brand-primary"
							>{formatInterval(newInterval)}</span
						>
					</div>
					<div class="flex flex-wrap gap-1.5 pt-1">
						{#each intervalPresets as preset (preset)}
							<button
								type="button"
								onclick={() => (newInterval = preset)}
								class="rounded-xl border px-2.5 py-1.5 text-[10px] font-bold transition-all {newInterval ===
								preset
									? 'border-brand-primary bg-brand-primary text-brand-dark shadow-lg shadow-brand-primary/20'
									: 'border-brand-light/10 bg-brand-light/5 text-brand-light/40 hover:border-brand-light/20'}"
							>
								{formatInterval(preset)}
							</button>
						{/each}
					</div>
				</div>
				<div class="mt-2 flex justify-end gap-3 md:col-span-4">
					<button
						type="button"
						onclick={() => (isAdding = false)}
						class="rounded-2xl border border-brand-light/10 px-6 py-3 font-bold transition-colors hover:bg-brand-light/5"
						>Cancel</button
					>
					<button
						type="submit"
						class="rounded-2xl bg-brand-primary px-8 py-3 font-bold text-brand-dark shadow-lg shadow-brand-primary/10 transition-all hover:bg-brand-primary/90"
						>Start monitoring</button
					>
				</div>
			</form>
		</div>
	{/if}

	{#if isEditing && editingServer}
		<div
			class="animate-in fade-in fixed inset-0 z-50 flex items-center justify-center bg-brand-dark/80 p-4 backdrop-blur-sm duration-200"
		>
			<div
				class="animate-in zoom-in-95 w-full max-w-2xl rounded-[2.5rem] border border-brand-light/10 bg-brand-dark p-8 shadow-2xl shadow-brand-primary/5 duration-200 lg:p-12"
			>
				<div class="mb-8 flex items-center justify-between">
					<div class="flex items-center gap-4">
						<div class="rounded-2xl bg-brand-primary/10 p-3 text-brand-primary">
							<Settings class="h-7 w-7" />
						</div>
						<div>
							<h2 class="text-2xl font-bold">Edit Monitor</h2>
							<p class="text-sm text-brand-light/40">Update the target and check schedule.</p>
						</div>
					</div>
					<button
						onclick={() => (isEditing = false)}
						class="rounded-xl p-2 transition-colors hover:bg-brand-light/5"
					>
						<X class="h-6 w-6 text-brand-light/20 hover:text-brand-light" />
					</button>
				</div>

				<form onsubmit={updateServer} class="grid gap-6">
					<div class="grid gap-6 md:grid-cols-2">
						<div class="space-y-2">
							<label
								for="edit-name"
								class="ml-1 text-xs font-bold tracking-widest text-brand-light/40 uppercase"
								>Monitor name</label
							>
							<input
								id="edit-name"
								type="text"
								bind:value={editName}
								required
								class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/50 px-5 py-3 transition-all outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
							/>
						</div>
						<div class="space-y-2">
							<label
								for="edit-url"
								class="ml-1 text-xs font-bold tracking-widest text-brand-light/40 uppercase"
								>URL or host</label
							>
							<input
								id="edit-url"
								type="text"
								bind:value={editUrl}
								required
								class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/50 px-5 py-3 transition-all outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
							/>
							{#if willNormalizeMonitorUrl(editUrl, editType)}
								<p
									class="rounded-2xl border border-brand-primary/20 bg-brand-primary/10 px-4 py-3 text-sm font-medium text-brand-primary"
								>
									We will save this as <span class="font-black"
										>{normalizeMonitorUrl(editUrl, editType)}</span
									>.
								</p>
							{/if}
						</div>
						<div class="space-y-2">
							<label
								for="edit-type"
								class="ml-1 text-xs font-bold tracking-widest text-brand-light/40 uppercase"
								>Check Type</label
							>
							<div class="relative">
								<select
									id="edit-type"
									bind:value={editType}
									class="w-full cursor-pointer appearance-none rounded-2xl border border-brand-light/10 bg-brand-dark/50 px-5 py-3 transition-all outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
								>
									<option value="http">HTTP (GET)</option>
									<option value="ping">TCP port</option>
								</select>
							</div>
						</div>
						<div class="space-y-2 md:col-span-2">
							<div class="ml-1 flex items-center justify-between">
								<label
									for="edit-interval"
									class="text-xs font-bold tracking-widest text-brand-light/40 uppercase"
									>Interval</label
								>
								<span
									class="rounded-full bg-brand-primary/10 px-2 py-0.5 text-xs font-black whitespace-nowrap text-brand-primary"
									>{formatInterval(editInterval)}</span
								>
							</div>
							<div class="flex flex-wrap gap-2 pt-2">
								{#each intervalPresets as preset (preset)}
									<button
										type="button"
										onclick={() => (editInterval = preset)}
										class="min-w-[80px] flex-1 rounded-2xl border px-4 py-2.5 text-xs font-bold transition-all {editInterval ===
										preset
											? 'border-brand-primary bg-brand-primary text-brand-dark shadow-lg shadow-brand-primary/20'
											: 'border-brand-light/10 bg-brand-light/5 text-brand-light/40 hover:border-brand-light/20'}"
									>
										{formatInterval(preset)}
									</button>
								{/each}
							</div>
						</div>
					</div>
					<div class="mt-4 flex justify-end gap-3">
						<button
							type="button"
							onclick={() => (isEditing = false)}
							class="rounded-2xl border border-brand-light/10 px-8 py-3 font-bold transition-colors hover:bg-brand-light/5"
							>Cancel</button
						>
						<button
							type="submit"
							class="rounded-2xl bg-brand-primary px-10 py-3 font-bold text-brand-dark shadow-lg shadow-brand-primary/10 transition-all hover:bg-brand-primary/90"
							>Save changes</button
						>
					</div>
				</form>
			</div>
		</div>
	{/if}

	{#if isLoading}
		<div class="flex h-96 flex-col items-center justify-center gap-4">
			<RefreshCw class="h-10 w-10 animate-spin text-brand-primary" />
			<p class="animate-pulse font-medium text-brand-light/25">Loading monitors...</p>
		</div>
	{:else if servers.length === 0}
		<div
			class="flex flex-col items-center justify-center rounded-[3rem] border-2 border-dashed border-brand-light/5 bg-brand-light/[0.02] py-32 text-center"
		>
			<div class="mb-6 rounded-full bg-brand-light/5 p-6 text-brand-light/10">
				<Activity class="h-16 w-16" />
			</div>
			<h3 class="mb-2 text-2xl font-bold">No monitors yet</h3>
			<p class="mx-auto max-w-xs text-brand-light/30">
				Add your first website, API endpoint, or TCP service to start tracking uptime.
			</p>
			<button
				onclick={() => (isAdding = true)}
				class="mt-8 inline-flex items-center gap-2 rounded-2xl bg-brand-primary px-6 py-3 font-bold text-brand-dark shadow-lg shadow-brand-primary/20 transition-all hover:scale-105 active:scale-95"
			>
				<Plus class="h-5 w-5" /> Add monitor
			</button>
		</div>
	{:else}
		<div class="grid gap-6">
			{#each servers as s (s.id)}
				{@const uptime = calculateUptime(s.history30d || [])}
				{@const avgLatency = calculateAvgLatency(s.history30d || [])}
				{@const isOnline = s.status.startsWith('2') || s.status === 'Connected'}

				<div
					class="group relative rounded-[2rem] border border-brand-light/10 bg-gradient-to-br from-brand-dark to-brand-dark/50 p-1 transition-all hover:border-brand-primary/30 hover:shadow-2xl hover:shadow-brand-primary/5"
				>
					<div
						class="flex flex-col justify-between gap-8 rounded-[1.9rem] bg-brand-dark/40 p-6 lg:flex-row lg:items-center lg:p-8"
					>
						<!-- Info & Status -->
						<div class="flex min-w-0 flex-1 items-start gap-5">
							<div class="relative mt-1 flex-shrink-0">
								<div
									class="flex h-14 w-14 items-center justify-center overflow-hidden rounded-2xl border border-brand-light/10 bg-brand-light/5 transition-colors group-hover:border-brand-primary/20"
								>
									<img
										src={getFaviconUrl(s.url)}
										alt={s.name}
										class="h-8 w-8 object-contain"
										onerror={(e) => {
											const target = e.currentTarget as HTMLImageElement;
											target.style.display = 'none';
											const parent = target.parentElement;
											if (parent && !parent.querySelector('.fallback-icon')) {
												const icon = document.createElement('div');
												icon.className = 'fallback-icon flex items-center justify-center';
												icon.innerHTML =
													s.check_type === 'http'
														? '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-globe h-6 w-6 text-brand-light/40 group-hover:text-brand-primary/60 transition-colors"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>'
														: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-activity h-6 w-6 text-brand-light/40 group-hover:text-brand-primary/60 transition-colors"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>';
												parent.appendChild(icon);
											}
										}}
									/>
								</div>
								<div
									class="absolute -right-1 -bottom-1 flex h-5 w-5 items-center justify-center rounded-full border-4 border-brand-dark"
									style="background-color: {getStatusColor(s.status, s.latency)}"
								>
									{#if isOnline}
										<div class="h-1.5 w-1.5 animate-pulse rounded-full bg-brand-dark"></div>
									{/if}
								</div>
							</div>

							<div class="min-w-0 flex-1">
								<div class="mb-1 flex items-center gap-3">
									<h3 class="truncate text-xl font-bold tracking-tight">{s.name}</h3>
									<span
										class="rounded-lg border border-brand-light/10 bg-brand-light/5 px-2 py-0.5 text-[10px] font-black tracking-widest text-brand-light/40 uppercase"
										>{s.check_type}</span
									>
								</div>
								<div class="flex items-center gap-2 text-brand-light/30">
									<p class="truncate text-sm font-medium">{s.url}</p>
									<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
									<a
										href={s.url}
										target="_blank"
										rel="external noreferrer"
										class="p-1 transition-colors hover:text-brand-primary"
										><ExternalLink class="h-3.5 w-3.5" /></a
									>
								</div>
							</div>
						</div>

						<!-- Metrics -->
						<div class="flex items-center gap-10 lg:gap-14">
							<div class="hidden text-right sm:block">
								<div
									class="mb-1 flex items-center justify-end gap-1.5 text-[10px] font-bold tracking-widest text-brand-light/40 uppercase"
								>
									<Clock class="h-3 w-3" /> Uptime 30d
								</div>
								<div
									class="text-2xl font-black {uptime >= 99
										? 'text-brand-primary'
										: uptime >= 95
											? 'text-brand-soft'
											: 'text-brand-accent'}"
								>
									{uptime.toFixed(1)}%
								</div>
							</div>

							<div class="hidden text-right sm:block">
								<div
									class="mb-1 flex items-center justify-end gap-1.5 text-[10px] font-bold tracking-widest text-brand-light/40 uppercase"
								>
									<BarChart3 class="h-3 w-3" /> Avg Latency 30d
								</div>
								<div class="text-2xl font-black text-brand-light/80">
									{avgLatency}<span class="ml-0.5 text-xs font-bold text-brand-light/20">ms</span>
								</div>
							</div>

							<!-- History Strip -->
							<div class="flex flex-col gap-2">
								<div class="flex h-10 w-48 flex-shrink-0 items-end gap-1">
									{#each getHourlyBuckets(s.history || []) as color, i (i)}
										<div
											class="group/item relative flex-1 cursor-help rounded-sm opacity-60 transition-all hover:h-12 hover:opacity-100"
											style="background-color: {color}; height: {color === '#1f332f'
												? '40%'
												: '100%'}"
										>
											<div
												class="pointer-events-none absolute bottom-full left-1/2 z-50 mb-3 hidden -translate-x-1/2 group-hover/item:block"
											>
												<div
													class="rounded-xl border border-brand-light/10 bg-brand-dark/95 px-3 py-2 text-[11px] whitespace-nowrap shadow-2xl ring-1 ring-white/5 backdrop-blur-xl"
												>
													<div class="mb-1 font-black text-brand-light/40">{23 - i}h ago</div>
													{#if color !== '#1f332f'}
														<div class="flex items-center gap-2">
															<div
																class="h-2 w-2 rounded-full"
																style="background-color: {color}"
															></div>
															<span class="font-bold"
																>{color === '#73E2A7'
																	? 'Healthy'
																	: color === '#E5B181'
																		? 'Slow response'
																		: 'Down'}</span
															>
														</div>
													{:else}
														<div class="text-brand-light/20 italic">No check data</div>
													{/if}
												</div>
												<div
													class="mx-auto -mt-1 h-2 w-2 rotate-45 border-r border-b border-brand-light/10 bg-brand-dark"
												></div>
											</div>
										</div>
									{/each}
								</div>
								<div
									class="flex justify-between px-0.5 text-[8px] font-bold tracking-widest text-brand-light/20 uppercase"
								>
									<span>24h ago</span>
									<span>Now</span>
								</div>
							</div>
						</div>

						<!-- Actions -->
						<div
							class="flex min-w-[120px] flex-col gap-2 border-t border-brand-light/5 pt-6 lg:border-t-0 lg:border-l lg:pt-0 lg:pl-6"
						>
							<a
								href={resolve('/dashboard/[id]', { id: String(s.id) })}
								class="w-full rounded-xl bg-brand-light/5 px-5 py-2.5 text-center text-xs font-bold transition-all hover:bg-brand-light/10"
							>
								Details
							</a>
							<div class="flex gap-2">
								<button
									onclick={() => openEdit(s)}
									class="flex flex-1 items-center justify-center rounded-lg bg-brand-light/5 p-1.5 text-brand-light/40 transition-all hover:bg-brand-light/10 hover:text-brand-light"
									title="Edit Monitor"
								>
									<Settings class="h-3.5 w-3.5" />
								</button>
								<button
									onclick={() => deleteServer(s.id)}
									class="flex flex-1 items-center justify-center rounded-lg bg-brand-accent/5 p-1.5 text-brand-accent/40 transition-all hover:bg-brand-accent/20 hover:text-brand-accent"
									title="Delete Monitor"
								>
									<Trash2 class="h-3.5 w-3.5" />
								</button>
							</div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	/* Custom animations for Svelte 5 */
	@keyframes fade-in {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}
	@keyframes slide-in-from-top-4 {
		from {
			transform: translateY(-1rem);
		}
		to {
			transform: translateY(0);
		}
	}
	@keyframes zoom-in-95 {
		from {
			transform: scale(0.95);
			opacity: 0;
		}
		to {
			transform: scale(1);
			opacity: 1;
		}
	}

	.animate-in {
		animation-duration: 300ms;
		animation-fill-mode: both;
	}
	.fade-in {
		animation-name: fade-in;
	}
	.slide-in-from-top-4 {
		animation-name: slide-in-from-top-4;
	}
	.zoom-in-95 {
		animation-name: zoom-in-95;
	}
</style>
