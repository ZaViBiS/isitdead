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
		Globe2,
		ShieldCheck,
		Settings,
		Mail,
		X
	} from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		getStatusColor,
		getHourlyBuckets,
		getRecentHistory,
		getCurrentCheck,
		getFaviconUrl,
		calculateUptime,
		calculateAvgLatency,
		type Server,
		type NotificationPreference
	} from '$lib/utils';

	let servers = $state<Server[]>([]);
	let isLoading = $state(true);
	let isAdding = $state(false);
	let isEditing = $state(false);
	let error = $state('');

	let newName = $state('');
	let newUrl = $state('');
	let newType = $state('http');
	let newInterval = $state(300);
	let newTimeout = $state(10);
	let newNotifyEmailDown = $state(true);
	let newNotifyEmailRecovered = $state(true);

	let editingServer = $state<Server | null>(null);
	let editName = $state('');
	let editUrl = $state('');
	let editType = $state('http');
	let editInterval = $state(300);
	let editTimeout = $state(10);
	let editNotifyEmailDown = $state(true);
	let editNotifyEmailRecovered = $state(true);

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
				s.history = getRecentHistory(data, 24);
			}
		} catch {
			console.error('Failed to fetch history');
		}
	}

	function normalizeMonitorUrl(url: string, checkType: string) {
		const trimmed = url.trim();
		if (!['http', 'links'].includes(checkType) || trimmed === '') return trimmed;
		if (trimmed.startsWith('http://') || trimmed.startsWith('https://')) return trimmed;
		return `https://${trimmed}`;
	}

	function willNormalizeMonitorUrl(url: string, checkType: string) {
		const trimmed = url.trim();
		return (
			['http', 'links'].includes(checkType) &&
			trimmed !== '' &&
			!trimmed.startsWith('http://') &&
			!trimmed.startsWith('https://')
		);
	}

	function getTargetPlaceholder(checkType: string) {
		if (checkType === 'ping') return 'example.com:80';
		if (checkType === 'links') return 'example.com or https://example.com';
		return 'example.com or http://example.com';
	}

	function notificationPayload(
		emailDown: boolean,
		emailRecovered: boolean
	): NotificationPreference[] {
		return [
			{ channel: 'email', event: 'down', enabled: emailDown },
			{ channel: 'email', event: 'recovered', enabled: emailRecovered }
		];
	}

	function preferenceEnabled(prefs: NotificationPreference[], event: string) {
		return prefs.find((pref) => pref.channel === 'email' && pref.event === event)?.enabled ?? true;
	}

	async function fetchNotificationPreferences(serverID: number) {
		const token = localStorage.getItem('token');
		const res = await fetch(`/api/servers/${serverID}/notifications`, {
			headers: { Authorization: `Bearer ${token}` }
		});
		if (!res.ok) throw new Error('Failed to load notification preferences');
		return (await res.json()) as NotificationPreference[];
	}

	async function saveNotificationPreferences(
		serverID: number,
		emailDown: boolean,
		emailRecovered: boolean
	) {
		const token = localStorage.getItem('token');
		const res = await fetch(`/api/servers/${serverID}/notifications`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
			body: JSON.stringify(notificationPayload(emailDown, emailRecovered))
		});
		if (!res.ok) throw new Error('Failed to save notification preferences');
	}

	function isServerOnline(server: Server) {
		const current = getCurrentCheck(server);
		return current?.status.startsWith('2') === true || current?.status === 'Connected';
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

	function compactStatus(status: string) {
		const normalized = status || 'unknown';
		return normalized.length > 28 ? `${normalized.slice(0, 25)}...` : normalized;
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
					check_interval: Number(newInterval),
					timeout: Number(newTimeout)
				})
			});

			if (res.ok) {
				const newSrv = await res.json();
				const server: Server = { ...newSrv, history: [], history30d: [] };
				servers.push(server);
				await saveNotificationPreferences(server.id, newNotifyEmailDown, newNotifyEmailRecovered);
				isAdding = false;
				newName = '';
				newUrl = '';
				newType = 'http';
				newInterval = 300;
				newTimeout = 10;
				newNotifyEmailDown = true;
				newNotifyEmailRecovered = true;
				fetchHistory(server);
			}
		} catch {
			error = 'Failed to add server';
		}
	}

	async function openEdit(server: Server) {
		editingServer = server;
		editName = server.name;
		editUrl = server.url;
		editType = server.check_type;
		editInterval = server.check_interval;
		editTimeout = server.timeout;
		editNotifyEmailDown = true;
		editNotifyEmailRecovered = true;
		isEditing = true;

		try {
			const prefs = await fetchNotificationPreferences(server.id);
			editNotifyEmailDown = preferenceEnabled(prefs, 'down');
			editNotifyEmailRecovered = preferenceEnabled(prefs, 'recovered');
		} catch {
			error = 'Failed to load notification preferences';
		}
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
					check_interval: Number(editInterval),
					timeout: Number(editTimeout)
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
				await saveNotificationPreferences(
					editingServer.id,
					editNotifyEmailDown,
					editNotifyEmailRecovered
				);
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
	const timeoutPresets = [3, 5, 10, 15, 30, 60];

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

<div class="relative isolate min-h-[calc(100vh-4rem)] overflow-hidden">
	<div class="pointer-events-none absolute inset-0 -z-10">
		<div
			class="absolute inset-0 bg-[linear-gradient(to_right,rgba(222,244,198,0.035)_1px,transparent_1px),linear-gradient(to_bottom,rgba(222,244,198,0.035)_1px,transparent_1px)] [mask-image:linear-gradient(to_bottom,black,transparent_86%)] bg-[size:64px_64px]"
		></div>
		<div class="absolute inset-x-0 top-0 h-64 bg-brand-light/[0.025]"></div>
	</div>

	<div class="container mx-auto max-w-7xl px-4 py-6 sm:px-6 sm:py-10 lg:py-12">
		<section class="mb-6 grid gap-4 lg:grid-cols-[1fr_auto] lg:items-end">
			<div class="min-w-0">
				<div
					class="mb-4 inline-flex items-center gap-2 rounded-full border border-brand-primary/20 bg-brand-primary/10 px-3 py-1.5 text-xs font-black text-brand-primary uppercase"
				>
					<span class="h-2 w-2 rounded-full bg-brand-primary"></span>
					Live operations
				</div>
				<h1 class="text-3xl font-black text-brand-light sm:text-5xl">Your monitors</h1>
				<p
					class="mt-3 flex max-w-2xl items-start gap-2 text-sm leading-6 text-brand-light/45 sm:text-base"
				>
					<ShieldCheck class="mt-1 h-4 w-4 shrink-0 text-brand-primary" />
					Track uptime, response time, and recent failures in one focused workspace.
				</p>
			</div>
			<button
				onclick={() => (isAdding = !isAdding)}
				class="inline-flex w-full items-center justify-center gap-2 rounded-2xl bg-brand-primary px-5 py-3 font-black text-brand-dark shadow-2xl shadow-brand-primary/15 transition hover:-translate-y-0.5 hover:bg-brand-primary/90 active:translate-y-0 sm:w-auto"
			>
				<Plus class="h-5 w-5" /> Add monitor
			</button>
		</section>

		{#if !isLoading && servers.length > 0}
			<section class="mb-6 grid gap-3 sm:grid-cols-3">
				<div class="rounded-3xl border border-brand-light/10 bg-brand-light/[0.035] p-4 sm:p-5">
					<div class="mb-4 flex items-center justify-between gap-3 text-brand-light/35">
						<span class="text-xs font-bold uppercase">Healthy now</span>
						<Activity class="h-4 w-4 text-brand-primary" />
					</div>
					<div class="text-3xl font-black text-brand-primary">
						{getHealthyCount()}<span class="text-sm text-brand-light/30">/{servers.length}</span>
					</div>
				</div>
				<div class="rounded-3xl border border-brand-light/10 bg-brand-light/[0.035] p-4 sm:p-5">
					<div class="mb-4 flex items-center justify-between gap-3 text-brand-light/35">
						<span class="text-xs font-bold uppercase">Average uptime</span>
						<Clock class="h-4 w-4 text-brand-primary" />
					</div>
					<div class="text-3xl font-black text-brand-primary">{getOverallUptime().toFixed(1)}%</div>
				</div>
				<div class="rounded-3xl border border-brand-light/10 bg-brand-light/[0.035] p-4 sm:p-5">
					<div class="mb-4 flex items-center justify-between gap-3 text-brand-light/35">
						<span class="text-xs font-bold uppercase">Average response</span>
						<BarChart3 class="h-4 w-4 text-brand-primary" />
					</div>
					<div class="text-3xl font-black text-brand-light/85">
						{getOverallLatency()}<span class="ml-1 text-sm text-brand-light/30">ms</span>
					</div>
				</div>
			</section>
		{/if}

		{#if error}
			<div
				class="animate-in fade-in slide-in-from-top-4 mb-6 flex items-center gap-3 rounded-2xl border border-brand-accent/20 bg-brand-accent/10 p-4 text-brand-accent"
			>
				<AlertCircle class="h-5 w-5 shrink-0" />
				{error}
			</div>
		{/if}

		{#if isAdding}
			<section
				class="animate-in zoom-in-95 mb-6 overflow-hidden rounded-[1.75rem] border border-brand-light/10 bg-[#111f1c]/90 shadow-2xl shadow-black/20 backdrop-blur duration-200"
			>
				<div class="flex items-center gap-3 border-b border-brand-light/10 px-4 py-4 sm:px-6">
					<div class="rounded-xl bg-brand-primary/10 p-2 text-brand-primary">
						<Activity class="h-5 w-5" />
					</div>
					<h2 class="text-xl font-black">Add monitor</h2>
				</div>
				<form
					onsubmit={addServer}
					class="grid gap-4 p-4 sm:p-6 lg:grid-cols-[1fr_1.25fr_0.8fr_0.9fr_1.3fr]"
				>
					<div class="space-y-2">
						<label for="name" class="ml-1 text-xs font-bold text-brand-light/45 uppercase"
							>Monitor name</label
						>
						<input
							id="name"
							type="text"
							bind:value={newName}
							required
							class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
							placeholder="Production API"
						/>
					</div>
					<div class="space-y-2">
						<label for="url" class="ml-1 text-xs font-bold text-brand-light/45 uppercase"
							>URL or host</label
						>
						<input
							id="url"
							type="text"
							bind:value={newUrl}
							required
							class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
							placeholder={getTargetPlaceholder(newType)}
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
						<label for="type" class="ml-1 text-xs font-bold text-brand-light/45 uppercase"
							>Check type</label
						>
						<select
							id="type"
							bind:value={newType}
							class="w-full cursor-pointer rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
						>
							<option value="http">HTTP (GET)</option>
							<option value="ping">TCP port</option>
							<option value="links">Broken links</option>
						</select>
					</div>
					<div class="space-y-2">
						<div class="ml-1 flex items-center justify-between">
							<label for="timeout" class="text-xs font-bold text-brand-light/45 uppercase"
								>Timeout</label
							>
							<span
								class="rounded-full bg-brand-primary/10 px-2.5 py-1 text-xs font-black whitespace-nowrap text-brand-primary"
								>{newTimeout}s</span
							>
						</div>
						<select
							id="timeout"
							bind:value={newTimeout}
							class="w-full cursor-pointer rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
						>
							{#each timeoutPresets as preset (preset)}
								<option value={preset}>{preset}s</option>
							{/each}
						</select>
					</div>
					<div class="space-y-2">
						<div class="ml-1 flex items-center justify-between">
							<label for="interval" class="text-xs font-bold text-brand-light/45 uppercase"
								>Interval</label
							>
							<span
								class="rounded-full bg-brand-primary/10 px-2.5 py-1 text-xs font-black whitespace-nowrap text-brand-primary"
								>{formatInterval(newInterval)}</span
							>
						</div>
						<div class="flex gap-2 overflow-x-auto pb-1 sm:flex-wrap">
							{#each intervalPresets as preset (preset)}
								<button
									type="button"
									onclick={() => (newInterval = preset)}
									class="min-w-14 rounded-xl border px-3 py-2 text-xs font-bold whitespace-nowrap transition {newInterval ===
									preset
										? 'border-brand-primary bg-brand-primary text-brand-dark shadow-lg shadow-brand-primary/20'
										: 'border-brand-light/10 bg-brand-light/5 text-brand-light/45 hover:border-brand-light/20'}"
								>
									{formatInterval(preset)}
								</button>
							{/each}
						</div>
					</div>
					<div
						class="grid gap-3 rounded-2xl border border-brand-light/10 bg-brand-light/[0.025] p-4 sm:grid-cols-2 lg:col-span-5"
					>
						<div class="flex items-center gap-3 sm:col-span-2">
							<Mail class="h-4 w-4 text-brand-primary" />
							<div class="text-xs font-bold text-brand-light/45 uppercase">Notifications</div>
						</div>
						<label
							class="flex cursor-pointer items-center justify-between gap-4 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3"
						>
							<span class="text-sm font-bold text-brand-light/75">Email when down</span>
							<input
								type="checkbox"
								bind:checked={newNotifyEmailDown}
								class="h-5 w-5 accent-brand-primary"
							/>
						</label>
						<label
							class="flex cursor-pointer items-center justify-between gap-4 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3"
						>
							<span class="text-sm font-bold text-brand-light/75">Email when recovered</span>
							<input
								type="checkbox"
								bind:checked={newNotifyEmailRecovered}
								class="h-5 w-5 accent-brand-primary"
							/>
						</label>
					</div>
					<div class="flex flex-col-reverse gap-3 pt-2 sm:flex-row sm:justify-end lg:col-span-5">
						<button
							type="button"
							onclick={() => (isAdding = false)}
							class="rounded-2xl border border-brand-light/10 px-6 py-3 font-bold transition hover:bg-brand-light/5"
							>Cancel</button
						>
						<button
							type="submit"
							class="rounded-2xl bg-brand-primary px-7 py-3 font-black text-brand-dark shadow-lg shadow-brand-primary/10 transition hover:bg-brand-primary/90"
							>Start monitoring</button
						>
					</div>
				</form>
			</section>
		{/if}

		{#if isEditing && editingServer}
			<div
				class="animate-in fade-in fixed inset-0 z-50 flex items-end justify-center bg-brand-dark/80 p-3 backdrop-blur-sm duration-200 sm:items-center sm:p-4"
			>
				<div
					class="animate-in zoom-in-95 max-h-[calc(100vh-1.5rem)] w-full max-w-2xl overflow-y-auto rounded-[1.75rem] border border-brand-light/10 bg-[#111f1c] p-4 shadow-2xl shadow-black/30 duration-200 sm:p-6 lg:p-8"
				>
					<div class="mb-6 flex items-start justify-between gap-4">
						<div class="flex items-center gap-3">
							<div class="rounded-2xl bg-brand-primary/10 p-3 text-brand-primary">
								<Settings class="h-6 w-6" />
							</div>
							<div>
								<h2 class="text-2xl font-black">Edit monitor</h2>
								<p class="text-sm text-brand-light/40">Update the target and check schedule.</p>
							</div>
						</div>
						<button
							onclick={() => (isEditing = false)}
							class="rounded-xl p-2 transition hover:bg-brand-light/5"
						>
							<X class="h-6 w-6 text-brand-light/30 hover:text-brand-light" />
						</button>
					</div>

					<form onsubmit={updateServer} class="grid gap-5">
						<div class="grid gap-4 md:grid-cols-2">
							<div class="space-y-2">
								<label for="edit-name" class="ml-1 text-xs font-bold text-brand-light/45 uppercase"
									>Monitor name</label
								>
								<input
									id="edit-name"
									type="text"
									bind:value={editName}
									required
									class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
								/>
							</div>
							<div class="space-y-2">
								<label for="edit-url" class="ml-1 text-xs font-bold text-brand-light/45 uppercase"
									>URL or host</label
								>
								<input
									id="edit-url"
									type="text"
									bind:value={editUrl}
									required
									class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
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
								<label for="edit-type" class="ml-1 text-xs font-bold text-brand-light/45 uppercase"
									>Check type</label
								>
								<select
									id="edit-type"
									bind:value={editType}
									class="w-full cursor-pointer rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
								>
									<option value="http">HTTP (GET)</option>
									<option value="ping">TCP port</option>
									<option value="links">Broken links</option>
								</select>
							</div>
							<div class="space-y-2 md:col-span-2">
								<div class="ml-1 flex items-center justify-between">
									<label for="edit-interval" class="text-xs font-bold text-brand-light/45 uppercase"
										>Interval</label
									>
									<span
										class="rounded-full bg-brand-primary/10 px-2.5 py-1 text-xs font-black whitespace-nowrap text-brand-primary"
										>{formatInterval(editInterval)}</span
									>
								</div>
								<div class="flex gap-2 overflow-x-auto pb-1 sm:flex-wrap">
									{#each intervalPresets as preset (preset)}
										<button
											type="button"
											onclick={() => (editInterval = preset)}
											class="min-w-14 rounded-xl border px-3 py-2 text-xs font-bold whitespace-nowrap transition {editInterval ===
											preset
												? 'border-brand-primary bg-brand-primary text-brand-dark shadow-lg shadow-brand-primary/20'
												: 'border-brand-light/10 bg-brand-light/5 text-brand-light/45 hover:border-brand-light/20'}"
										>
											{formatInterval(preset)}
										</button>
									{/each}
								</div>
							</div>
							<div class="space-y-2 md:col-span-2">
								<div class="ml-1 flex items-center justify-between">
									<label for="edit-timeout" class="text-xs font-bold text-brand-light/45 uppercase"
										>Timeout</label
									>
									<span
										class="rounded-full bg-brand-primary/10 px-2.5 py-1 text-xs font-black whitespace-nowrap text-brand-primary"
										>{editTimeout}s</span
									>
								</div>
								<select
									id="edit-timeout"
									bind:value={editTimeout}
									class="w-full cursor-pointer rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
								>
									{#each timeoutPresets as preset (preset)}
										<option value={preset}>{preset}s</option>
									{/each}
								</select>
							</div>
							<div
								class="grid gap-3 rounded-2xl border border-brand-light/10 bg-brand-light/[0.025] p-4 md:col-span-2 md:grid-cols-2"
							>
								<div class="flex items-center gap-3 md:col-span-2">
									<Mail class="h-4 w-4 text-brand-primary" />
									<div class="text-xs font-bold text-brand-light/45 uppercase">Notifications</div>
								</div>
								<label
									class="flex cursor-pointer items-center justify-between gap-4 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3"
								>
									<span class="text-sm font-bold text-brand-light/75">Email when down</span>
									<input
										type="checkbox"
										bind:checked={editNotifyEmailDown}
										class="h-5 w-5 accent-brand-primary"
									/>
								</label>
								<label
									class="flex cursor-pointer items-center justify-between gap-4 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3"
								>
									<span class="text-sm font-bold text-brand-light/75">Email when recovered</span>
									<input
										type="checkbox"
										bind:checked={editNotifyEmailRecovered}
										class="h-5 w-5 accent-brand-primary"
									/>
								</label>
							</div>
						</div>
						<div class="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
							<button
								type="button"
								onclick={() => (isEditing = false)}
								class="rounded-2xl border border-brand-light/10 px-7 py-3 font-bold transition hover:bg-brand-light/5"
								>Cancel</button
							>
							<button
								type="submit"
								class="rounded-2xl bg-brand-primary px-8 py-3 font-black text-brand-dark shadow-lg shadow-brand-primary/10 transition hover:bg-brand-primary/90"
								>Save changes</button
							>
						</div>
					</form>
				</div>
			</div>
		{/if}

		{#if isLoading}
			<div
				class="flex min-h-80 flex-col items-center justify-center gap-4 rounded-[2rem] border border-brand-light/10 bg-brand-light/[0.025]"
			>
				<RefreshCw class="h-9 w-9 animate-spin text-brand-primary" />
				<p class="animate-pulse font-medium text-brand-light/30">Loading monitors...</p>
			</div>
		{:else if servers.length === 0}
			<section
				class="flex flex-col items-center justify-center rounded-[2rem] border border-dashed border-brand-light/10 bg-brand-light/[0.025] px-5 py-20 text-center sm:py-28"
			>
				<div
					class="mb-6 rounded-3xl border border-brand-light/10 bg-brand-light/[0.04] p-5 text-brand-primary"
				>
					<Activity class="h-12 w-12" />
				</div>
				<h3 class="mb-2 text-2xl font-black">No monitors yet</h3>
				<p class="mx-auto max-w-sm text-sm leading-6 text-brand-light/40">
					Add your first website, API endpoint, or TCP service to start tracking uptime.
				</p>
				<button
					onclick={() => (isAdding = true)}
					class="mt-8 inline-flex w-full items-center justify-center gap-2 rounded-2xl bg-brand-primary px-6 py-3 font-black text-brand-dark shadow-lg shadow-brand-primary/20 transition hover:-translate-y-0.5 hover:bg-brand-primary/90 active:translate-y-0 sm:w-auto"
				>
					<Plus class="h-5 w-5" /> Add monitor
				</button>
			</section>
			{:else}
				<section class="grid gap-4">
					{#each servers as s (s.id)}
						{@const uptime = calculateUptime(s.history30d || [])}
						{@const avgLatency = calculateAvgLatency(s.history30d || [])}
						{@const current = getCurrentCheck(s)}
						{@const currentStatus = current?.status ?? 'unknown'}
						{@const currentLatency = current?.latency ?? 0}
						{@const isOnline = currentStatus.startsWith('2') || currentStatus === 'Connected'}
						{@const isUnknown = currentStatus === 'unknown'}

					<article
						class="group rounded-[1.75rem] border border-brand-light/10 bg-[#111f1c]/90 p-4 shadow-xl shadow-black/10 transition hover:border-brand-primary/30 sm:p-5"
					>
						<div
							class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(22rem,0.72fr)_8.5rem] xl:items-center"
						>
							<div class="min-w-0">
								<div class="flex items-start gap-4">
									<div class="relative shrink-0">
										<div
											class="relative flex h-12 w-12 items-center justify-center overflow-hidden rounded-2xl border border-brand-light/10 bg-brand-light/[0.04] text-brand-light/35 transition group-hover:border-brand-primary/25 group-hover:text-brand-primary sm:h-14 sm:w-14"
										>
											{#if s.check_type === 'http'}
												<Globe2 class="h-6 w-6" />
											{:else}
												<Activity class="h-6 w-6" />
											{/if}
											<img
												src={getFaviconUrl(s.url)}
												alt=""
												class="absolute h-8 w-8 rounded-md bg-[#111f1c] object-contain"
												onerror={(e) => {
													const target = e.currentTarget as HTMLImageElement;
													target.style.display = 'none';
												}}
											/>
										</div>
										<div
											class="absolute -right-1 -bottom-1 flex h-5 w-5 items-center justify-center rounded-full border-4 border-[#111f1c]"
											style="background-color: {getStatusColor(currentStatus, currentLatency)}"
										>
											{#if isOnline}
												<span class="h-1.5 w-1.5 animate-pulse rounded-full bg-brand-dark"></span>
											{/if}
										</div>
									</div>

									<div class="min-w-0 flex-1">
										<div class="flex flex-wrap items-center gap-2">
											<h3 class="max-w-full min-w-0 truncate text-lg font-black sm:text-xl">
												{s.name}
											</h3>
											<span
												class="rounded-lg border border-brand-light/10 bg-brand-light/[0.04] px-2 py-0.5 text-[10px] font-black text-brand-light/45 uppercase"
												>{s.check_type}</span
											>
											<span
												class="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-bold {isOnline
													? 'border-brand-primary/20 bg-brand-primary/10 text-brand-primary'
													: isUnknown
														? 'border-brand-light/10 bg-brand-light/[0.04] text-brand-light/40'
														: 'border-brand-accent/20 bg-brand-accent/10 text-brand-accent'}"
											>
												<span
													class="h-1.5 w-1.5 rounded-full"
													style="background-color: {getStatusColor(currentStatus, currentLatency)}"
												></span>
												{isOnline ? 'Online' : isUnknown ? 'Unknown' : 'Down'}
											</span>
										</div>
										<div class="mt-2 flex min-w-0 items-center gap-2 text-brand-light/35">
											<p class="min-w-0 truncate text-sm font-medium">{s.url}</p>
											<a
												href={s.url}
												target="_blank"
												rel="external noreferrer"
												class="shrink-0 rounded-lg p-1 transition hover:bg-brand-light/5 hover:text-brand-primary"
												title="Open target"
											>
													<ExternalLink class="h-3.5 w-3.5" />
												</a>
											</div>
											<p
												class="mt-2 truncate text-xs font-medium text-brand-light/25"
												title={currentStatus}
											>
												{compactStatus(currentStatus)}
											</p>
										{#if s.public && s.public_slug}
											<a
												href={resolve('/status/[slug]', { slug: s.public_slug })}
												class="mt-2 inline-flex items-center gap-1.5 rounded-lg border border-brand-primary/15 bg-brand-primary/10 px-2 py-1 text-[10px] font-black text-brand-primary uppercase"
											>
												Public page <ExternalLink class="h-3 w-3" />
											</a>
										{/if}
									</div>
								</div>
							</div>

							<div class="min-w-0">
								<div
									class="grid grid-cols-3 gap-3 border-y border-brand-light/10 py-4 xl:border-y-0 xl:py-0"
								>
									<div>
										<div class="mb-1 text-xs font-bold text-brand-light/35 uppercase">Uptime</div>
										<div
											class="text-xl font-black {uptime >= 99
												? 'text-brand-primary'
												: uptime >= 95
													? 'text-brand-soft'
													: 'text-brand-accent'}"
										>
											{uptime.toFixed(1)}%
										</div>
									</div>
									<div>
										<div class="mb-1 text-xs font-bold text-brand-light/35 uppercase">Latency</div>
										<div class="text-xl font-black text-brand-light/85">
											{avgLatency}<span class="ml-0.5 text-xs font-bold text-brand-light/25"
												>ms</span
											>
										</div>
									</div>
									<div>
										<div class="mb-1 text-xs font-bold text-brand-light/35 uppercase">Interval</div>
										<div class="text-xl font-black text-brand-light/85">
											{formatInterval(s.check_interval)}
										</div>
									</div>
								</div>

								<div class="mt-4">
									<div class="flex h-9 w-full items-end gap-1">
										{#each getHourlyBuckets(s.history || []) as color, i (i)}
											<div
												class="group/item relative flex-1 cursor-help rounded-sm opacity-70 transition hover:opacity-100"
												style="background-color: {color}; height: {color === '#1f332f'
													? '38%'
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
										class="mt-2 flex justify-between text-[10px] font-bold text-brand-light/25 uppercase"
									>
										<span>24h ago</span>
										<span>Now</span>
									</div>
								</div>
							</div>

							<div class="grid grid-cols-[1fr_2.75rem_2.75rem] gap-2 xl:flex xl:flex-col">
								<a
									href={resolve('/dashboard/[id]', { id: String(s.id) })}
									class="inline-flex h-11 items-center justify-center rounded-xl bg-brand-light/[0.06] px-4 text-sm font-bold transition hover:bg-brand-light/10"
								>
									Details
								</a>
								<button
									onclick={() => openEdit(s)}
									class="inline-flex h-11 items-center justify-center rounded-xl bg-brand-light/[0.06] text-brand-light/45 transition hover:bg-brand-light/10 hover:text-brand-light"
									title="Edit monitor"
								>
									<Settings class="h-4 w-4" />
								</button>
								<button
									onclick={() => deleteServer(s.id)}
									class="inline-flex h-11 items-center justify-center rounded-xl bg-brand-accent/10 text-brand-accent/60 transition hover:bg-brand-accent/20 hover:text-brand-accent"
									title="Delete monitor"
								>
									<Trash2 class="h-4 w-4" />
								</button>
							</div>
						</div>
					</article>
				{/each}
			</section>
		{/if}
	</div>
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
