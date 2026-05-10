<script lang="ts">
	import { onMount } from 'svelte';
	import { Activity, Plus, Trash2, ExternalLink, RefreshCw, AlertCircle, Clock, BarChart3, Globe, ShieldCheck, Settings, X } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { getStatusColor, getHourlyBuckets, getFaviconUrl, type Server, type CheckResult } from '$lib/utils';

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
			goto('/login');
			return;
		}

		try {
			const res = await fetch('/api/servers', {
				headers: { Authorization: `Bearer ${token}` }
			});

			if (res.ok) {
				const data = await res.json();
				servers = data.map((s: any) => ({
					...s,
					history: [],
					history30d: []
				}));
				servers.forEach(s => fetchHistory(s));
			} else if (res.status === 401) {
				localStorage.removeItem('token');
				goto('/login');
			}
		} catch (err) {
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
		} catch (err) {
			console.error('Failed to fetch history');
		}
	}

	async function addServer(e: SubmitEvent) {
		e.preventDefault();
		const token = localStorage.getItem('token');
		try {
			const res = await fetch('/api/servers', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({ name: newName, url: newUrl, check_type: newType, check_interval: Number(newInterval) })
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
		} catch (err) {
			error = 'Failed to add server';
		}
	}

	function openEdit(server: Server) {
		editingServer = server;
		editName = server.name;
		editUrl = server.url;
		editType = server.check_type;
		editInterval = server.check_interval;
		editSliderVal = secondsToSlider(server.check_interval);
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
				body: JSON.stringify({ name: editName, url: editUrl, check_type: editType, check_interval: Number(editInterval) })
			});

			if (res.ok) {
				const updated = await res.json();
				const idx = servers.findIndex(s => s.id === updated.id);
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
		} catch (err) {
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
				const idx = servers.findIndex(s => s.id === id);
				if (idx !== -1) servers.splice(idx, 1);
			}
		} catch (err) {
			error = 'Failed to delete server';
		}
	}

	function calculateUptime(history: CheckResult[]) {
		if (!history || history.length === 0) return 0;
		const online = history.filter(r => r.status.startsWith('2') || r.status === 'Connected').length;
		return (online / history.length) * 100;
	}

	function calculateAvgLatency(history: CheckResult[]) {
		if (!history || history.length === 0) return 0;
		const sum = history.reduce((acc, r) => acc + r.latency, 0);
		return Math.round(sum / history.length);
	}

	function formatInterval(seconds: number) {
		if (seconds < 60) return `${seconds}s`;
		const m = Math.floor(seconds / 60);
		const s = seconds % 60;
		if (m < 60) return s > 0 ? `${m}m ${s}s` : `${m}m`;
		const h = Math.floor(m / 60);
		const mm = m % 60;
		return mm > 0 ? `${h}h ${mm}m` : `${h}h`;
	}

	// Non-linear mapping: 0-100 slider value to 10-86400 seconds
	function sliderToSeconds(val: number): number {
		if (val <= 25) {
			// 0-25%: 10s to 5m (300s), 1s precision
			return Math.round(10 + (val / 25) * 290);
		}
		if (val <= 50) {
			// 25-50%: 5m to 30m (1800s), 1m (60s) steps
			const raw = 300 + ((val - 25) / 25) * 1500;
			return Math.round(raw / 60) * 60;
		}
		if (val <= 75) {
			// 50-75%: 30m to 1h (3600s), 5m (300s) steps
			const raw = 1800 + ((val - 50) / 25) * 1800;
			return Math.round(raw / 300) * 300;
		}
		// 75-100%: 1h to 24h (86400s), 15m (900s) steps
		const raw = 3600 + ((val - 75) / 25) * 82800;
		return Math.round(raw / 900) * 900;
	}

	function secondsToSlider(secs: number): number {
		if (secs <= 300) return ((secs - 10) / 290) * 25;
		if (secs <= 1800) return 25 + ((secs - 300) / 1500) * 25;
		if (secs <= 3600) return 50 + ((secs - 1800) / 1800) * 25;
		return 75 + ((secs - 3600) / 82800) * 25;
	}

	let newSliderVal = $state(secondsToSlider(newInterval));
	let editSliderVal = $state(secondsToSlider(editInterval));

	onMount(fetchServers);
</script>

<div class="container mx-auto px-4 py-12 max-w-6xl">
	<div class="mb-12 flex flex-col md:flex-row md:items-end justify-between gap-6">
		<div>
			<h1 class="text-4xl font-black tracking-tight mb-2">Monitors</h1>
			<p class="text-brand-light/40 flex items-center gap-2">
				<ShieldCheck class="h-4 w-4 text-brand-primary" />
				Professional infrastructure health monitoring.
			</p>
		</div>
		<button
			onclick={() => (isAdding = !isAdding)}
			class="flex items-center justify-center gap-2 rounded-2xl bg-brand-primary px-6 py-3 font-bold text-brand-dark hover:scale-105 active:scale-95 transition-all shadow-lg shadow-brand-primary/20"
		>
			<Plus class="h-5 w-5" /> Add New Monitor
		</button>
	</div>

	{#if error}
		<div class="mb-8 flex items-center gap-3 rounded-2xl bg-brand-accent/10 border border-brand-accent/20 p-4 text-brand-accent animate-in fade-in slide-in-from-top-4">
			<AlertCircle class="h-5 w-5" /> {error}
		</div>
	{/if}

	{#if isAdding}
		<div class="mb-12 rounded-3xl border border-brand-light/10 bg-brand-light/5 p-8 shadow-2xl backdrop-blur-sm animate-in zoom-in-95 duration-200">
			<div class="flex items-center gap-3 mb-6">
				<div class="p-2 rounded-xl bg-brand-primary/10 text-brand-primary">
					<Activity class="h-6 w-6" />
				</div>
				<h2 class="text-2xl font-bold">New Monitor Configuration</h2>
			</div>
			<form onsubmit={addServer} class="grid gap-6 md:grid-cols-4">
				<div class="space-y-2">
					<label for="name" class="text-xs font-bold text-brand-light/40 uppercase tracking-widest ml-1">Friendly Name</label>
					<input id="name" type="text" bind:value={newName} required class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/50 px-5 py-3 focus:border-brand-primary focus:ring-1 focus:ring-brand-primary outline-none transition-all" placeholder="Production API" />
				</div>
				<div class="space-y-2">
					<label for="url" class="text-xs font-bold text-brand-light/40 uppercase tracking-widest ml-1">Endpoint URL / Host</label>
					<input id="url" type="text" bind:value={newUrl} required class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/50 px-5 py-3 focus:border-brand-primary focus:ring-1 focus:ring-brand-primary outline-none transition-all" placeholder={newType === 'http' ? 'https://api.example.com' : 'example.com:80'} />
				</div>
				<div class="space-y-2">
					<label for="type" class="text-xs font-bold text-brand-light/40 uppercase tracking-widest ml-1">Check Type</label>
					<div class="relative">
						<select id="type" bind:value={newType} class="w-full appearance-none rounded-2xl border border-brand-light/10 bg-brand-dark/50 px-5 py-3 focus:border-brand-primary focus:ring-1 focus:ring-brand-primary outline-none transition-all cursor-pointer">
							<option value="http">HTTP (GET)</option>
							<option value="ping">TCP Ping</option>
						</select>
					</div>
				</div>
				<div class="space-y-2">
					<div class="flex justify-between items-center ml-1">
						<label for="interval" class="text-xs font-bold text-brand-light/40 uppercase tracking-widest">Interval</label>
						<span class="text-xs font-black text-brand-primary bg-brand-primary/10 px-2 py-0.5 rounded-full whitespace-nowrap">{formatInterval(newInterval)}</span>
					</div>
					<div class="flex items-center h-12">
						<input 
							id="interval" 
							type="range" 
							min="0" 
							max="100" 
							step="0.1"
							bind:value={newSliderVal} 
							oninput={() => (newInterval = sliderToSeconds(newSliderVal))}
							class="w-full h-1.5 bg-brand-light/10 rounded-lg appearance-none cursor-pointer accent-brand-primary" 
						/>
					</div>
				</div>
				<div class="md:col-span-4 flex justify-end gap-3 mt-2">
					<button type="button" onclick={() => (isAdding = false)} class="px-6 py-3 rounded-2xl border border-brand-light/10 font-bold hover:bg-brand-light/5 transition-colors">Cancel</button>
					<button type="submit" class="px-8 py-3 rounded-2xl bg-brand-primary font-bold text-brand-dark hover:bg-brand-primary/90 transition-all shadow-lg shadow-brand-primary/10">Start Monitoring</button>
				</div>
			</form>
		</div>
	{/if}

	{#if isEditing && editingServer}
		<div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-brand-dark/80 backdrop-blur-sm animate-in fade-in duration-200">
			<div class="w-full max-w-2xl rounded-[2.5rem] border border-brand-light/10 bg-brand-dark p-8 lg:p-12 shadow-2xl shadow-brand-primary/5 animate-in zoom-in-95 duration-200">
				<div class="flex items-center justify-between mb-8">
					<div class="flex items-center gap-4">
						<div class="p-3 rounded-2xl bg-brand-primary/10 text-brand-primary">
							<Settings class="h-7 w-7" />
						</div>
						<div>
							<h2 class="text-2xl font-bold">Edit Monitor</h2>
							<p class="text-sm text-brand-light/40">Adjust your monitoring parameters.</p>
						</div>
					</div>
					<button onclick={() => (isEditing = false)} class="p-2 hover:bg-brand-light/5 rounded-xl transition-colors">
						<X class="h-6 w-6 text-brand-light/20 hover:text-brand-light" />
					</button>
				</div>

				<form onsubmit={updateServer} class="grid gap-6">
					<div class="grid md:grid-cols-2 gap-6">
						<div class="space-y-2">
							<label for="edit-name" class="text-xs font-bold text-brand-light/40 uppercase tracking-widest ml-1">Friendly Name</label>
							<input id="edit-name" type="text" bind:value={editName} required class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/50 px-5 py-3 focus:border-brand-primary focus:ring-1 focus:ring-brand-primary outline-none transition-all" />
						</div>
						<div class="space-y-2">
							<label for="edit-url" class="text-xs font-bold text-brand-light/40 uppercase tracking-widest ml-1">Endpoint URL / Host</label>
							<input id="edit-url" type="text" bind:value={editUrl} required class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/50 px-5 py-3 focus:border-brand-primary focus:ring-1 focus:ring-brand-primary outline-none transition-all" />
						</div>
						<div class="space-y-2">
							<label for="edit-type" class="text-xs font-bold text-brand-light/40 uppercase tracking-widest ml-1">Check Type</label>
							<div class="relative">
								<select id="edit-type" bind:value={editType} class="w-full appearance-none rounded-2xl border border-brand-light/10 bg-brand-dark/50 px-5 py-3 focus:border-brand-primary focus:ring-1 focus:ring-brand-primary outline-none transition-all cursor-pointer">
									<option value="http">HTTP (GET)</option>
									<option value="ping">TCP Ping</option>
								</select>
							</div>
						</div>
						<div class="space-y-2">
							<div class="flex justify-between items-center ml-1">
								<label for="edit-interval" class="text-xs font-bold text-brand-light/40 uppercase tracking-widest">Interval</label>
								<span class="text-xs font-black text-brand-primary bg-brand-primary/10 px-2 py-0.5 rounded-full whitespace-nowrap">{formatInterval(editInterval)}</span>
							</div>
							<div class="flex items-center h-12">
								<input 
									id="edit-interval" 
									type="range" 
									min="0" 
									max="100" 
									step="0.1"
									bind:value={editSliderVal} 
									oninput={() => (editInterval = sliderToSeconds(editSliderVal))}
									class="w-full h-1.5 bg-brand-light/10 rounded-lg appearance-none cursor-pointer accent-brand-primary" 
								/>
							</div>
						</div>
					</div>
					<div class="flex justify-end gap-3 mt-4">
						<button type="button" onclick={() => (isEditing = false)} class="px-8 py-3 rounded-2xl border border-brand-light/10 font-bold hover:bg-brand-light/5 transition-colors">Discard</button>
						<button type="submit" class="px-10 py-3 rounded-2xl bg-brand-primary font-bold text-brand-dark hover:bg-brand-primary/90 transition-all shadow-lg shadow-brand-primary/10">Save Changes</button>
					</div>
				</form>
			</div>
		</div>
	{/if}

	{#if isLoading}
		<div class="flex h-96 flex-col items-center justify-center gap-4">
			<RefreshCw class="h-10 w-10 animate-spin text-brand-primary" />
			<p class="text-brand-light/20 font-medium animate-pulse">Initializing dashboard...</p>
		</div>
	{:else if servers.length === 0}
		<div class="flex flex-col items-center justify-center rounded-[3rem] border-2 border-dashed border-brand-light/5 bg-brand-light/[0.02] py-32 text-center">
			<div class="mb-6 p-6 rounded-full bg-brand-light/5 text-brand-light/10">
				<Activity class="h-16 w-16" />
			</div>
			<h3 class="text-2xl font-bold mb-2">System Silence</h3>
			<p class="text-brand-light/30 max-w-xs mx-auto">No monitors configured yet. Add your first server to start tracking availability.</p>
		</div>
	{:else}
		<div class="grid gap-6">
			{#each servers as s (s.id)}
				{@const uptime = calculateUptime(s.history30d || [])}
				{@const avgLatency = calculateAvgLatency(s.history30d || [])}
				{@const isOnline = (s.status.startsWith('2') || s.status === 'Connected')}
				
				<div class="group relative rounded-[2rem] border border-brand-light/10 bg-gradient-to-br from-brand-dark to-brand-dark/50 p-1 transition-all hover:border-brand-primary/30 hover:shadow-2xl hover:shadow-brand-primary/5">
					<div class="bg-brand-dark/40 rounded-[1.9rem] p-6 lg:p-8 flex flex-col lg:flex-row lg:items-center justify-between gap-8">
						
						<!-- Info & Status -->
						<div class="flex items-start gap-5 flex-1 min-w-0">
							<div class="relative flex-shrink-0 mt-1">
								<div class="h-14 w-14 rounded-2xl bg-brand-light/5 flex items-center justify-center border border-brand-light/10 group-hover:border-brand-primary/20 transition-colors overflow-hidden">
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
												icon.innerHTML = s.check_type === 'http' ? '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-globe h-6 w-6 text-brand-light/40 group-hover:text-brand-primary/60 transition-colors"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>' : '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-activity h-6 w-6 text-brand-light/40 group-hover:text-brand-primary/60 transition-colors"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>';
												parent.appendChild(icon);
											}
										}}
									/>
								</div>
								<div 
									class="absolute -bottom-1 -right-1 h-5 w-5 rounded-full border-4 border-brand-dark flex items-center justify-center"
									style="background-color: {getStatusColor(s.status, s.latency)}"
								>
									{#if isOnline}
										<div class="h-1.5 w-1.5 rounded-full bg-brand-dark animate-pulse"></div>
									{/if}
								</div>
							</div>

							<div class="min-w-0 flex-1">
								<div class="flex items-center gap-3 mb-1">
									<h3 class="font-bold text-xl truncate tracking-tight">{s.name}</h3>
									<span class="text-[10px] px-2 py-0.5 rounded-lg bg-brand-light/5 border border-brand-light/10 text-brand-light/40 uppercase font-black tracking-widest">{s.check_type}</span>
								</div>
								<div class="flex items-center gap-2 text-brand-light/30">
									<p class="truncate text-sm font-medium">{s.url}</p>
									<a href={s.url} target="_blank" class="p-1 hover:text-brand-primary transition-colors"><ExternalLink class="h-3.5 w-3.5" /></a>
								</div>
							</div>
						</div>

						<!-- Metrics -->
						<div class="flex items-center gap-10 lg:gap-14">
							<div class="hidden sm:block text-right">
								<div class="flex items-center justify-end gap-1.5 text-brand-light/40 text-[10px] font-bold uppercase tracking-widest mb-1">
									<Clock class="h-3 w-3" /> Uptime 30d
								</div>
								<div class="text-2xl font-black {uptime >= 99 ? 'text-brand-primary' : uptime >= 95 ? 'text-brand-soft' : 'text-brand-accent'}">
									{uptime.toFixed(1)}%
								</div>
							</div>

							<div class="hidden sm:block text-right">
								<div class="flex items-center justify-end gap-1.5 text-brand-light/40 text-[10px] font-bold uppercase tracking-widest mb-1">
									<BarChart3 class="h-3 w-3" /> Avg Latency 30d
								</div>
								<div class="text-2xl font-black text-brand-light/80">
									{avgLatency}<span class="text-xs font-bold text-brand-light/20 ml-0.5">ms</span>
								</div>
							</div>

							<!-- History Strip -->
							<div class="flex flex-col gap-2">
								<div class="flex gap-1 h-10 w-48 flex-shrink-0 items-end">
									{#each getHourlyBuckets(s.history || []) as color, i}
										<div 
											class="flex-1 rounded-sm opacity-60 hover:opacity-100 hover:h-12 transition-all cursor-help relative group/item"
											style="background-color: {color}; height: {color === '#1f332f' ? '40%' : '100%'}"
										>
											<div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-3 hidden group-hover/item:block z-50 pointer-events-none">
												<div class="bg-brand-dark/95 border border-brand-light/10 text-[11px] whitespace-nowrap px-3 py-2 rounded-xl shadow-2xl backdrop-blur-xl ring-1 ring-white/5">
													<div class="font-black text-brand-light/40 mb-1">{23-i}h ago</div>
													{#if color !== '#1f332f'}
														<div class="flex items-center gap-2">
															<div class="h-2 w-2 rounded-full" style="background-color: {color}"></div>
															<span class="font-bold">System {color === '#73E2A7' ? 'Healthy' : color === '#E5B181' ? 'Degraded' : 'Critical'}</span>
														</div>
													{:else}
														<div class="text-brand-light/20 italic">No check data</div>
													{/if}
												</div>
												<div class="w-2 h-2 bg-brand-dark border-r border-b border-brand-light/10 rotate-45 mx-auto -mt-1"></div>
											</div>
										</div>
									{/each}
								</div>
								<div class="flex justify-between text-[8px] font-bold text-brand-light/20 uppercase tracking-widest px-0.5">
									<span>24h ago</span>
									<span>Now</span>
								</div>
							</div>
						</div>

						<!-- Actions -->
						<div class="flex flex-col gap-2 border-t lg:border-t-0 lg:border-l border-brand-light/5 pt-6 lg:pt-0 lg:pl-6 min-w-[120px]">
							<a 
								href="/dashboard/{s.id}" 
								class="w-full text-center px-5 py-2.5 rounded-xl bg-brand-light/5 hover:bg-brand-light/10 text-xs font-bold transition-all"
							>
								Details
							</a>
							<div class="flex gap-2">
								<button 
									onclick={() => openEdit(s)} 
									class="flex-1 flex items-center justify-center p-1.5 rounded-lg bg-brand-light/5 hover:bg-brand-light/10 text-brand-light/40 hover:text-brand-light transition-all"
									title="Edit Monitor"
								>
									<Settings class="h-3.5 w-3.5" />
								</button>
								<button 
									onclick={() => deleteServer(s.id)} 
									class="flex-1 flex items-center justify-center p-1.5 rounded-lg bg-brand-accent/5 hover:bg-brand-accent/20 text-brand-accent/40 hover:text-brand-accent transition-all"
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
	@keyframes fade-in { from { opacity: 0; } to { opacity: 1; } }
	@keyframes slide-in-from-top-4 { from { transform: translateY(-1rem); } to { transform: translateY(0); } }
	@keyframes zoom-in-95 { from { transform: scale(0.95); opacity: 0; } to { transform: scale(1); opacity: 1; } }

	.animate-in {
		animation-duration: 300ms;
		animation-fill-mode: both;
	}
	.fade-in { animation-name: fade-in; }
	.slide-in-from-top-4 { animation-name: slide-in-from-top-4; }
	.zoom-in-95 { animation-name: zoom-in-95; }
</style>
