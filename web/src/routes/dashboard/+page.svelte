<script lang="ts">
	import { onMount } from 'svelte';
	import { Activity, Plus, Trash2, ExternalLink, RefreshCw, AlertCircle, ChevronDown, ChevronUp } from 'lucide-svelte';
	import { goto } from '$app/navigation';

	interface CheckResult {
		id: number;
		status: string;
		latency: number;
		created_at: string;
	}

	interface Server {
		id: number;
		name: string;
		url: string;
		check_type: string;
		status: string;
		latency: number;
		check_interval: number;
		history: CheckResult[];
		isExpanded: boolean;
		hoveredResult: CheckResult | null;
		isLoadingFull: boolean;
	}

	let servers = $state<Server[]>([]);
	let isLoading = $state(true);
	let isAdding = $state(false);
	let error = $state('');

	let newName = $state('');
	let newUrl = $state('');
	let newType = $state('http');
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
				const data = await res.json();
				servers = data.map((s: any) => ({
					...s,
					history: [],
					isExpanded: false,
					hoveredResult: null,
					isLoadingFull: false
				}));
				servers.forEach(s => fetchHistory(s, 20));
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

	async function fetchHistory(s: Server, limit?: number) {
		const token = localStorage.getItem('token');
		const url = `/api/servers/${s.id}/results${limit ? `?limit=${limit}` : ''}`;
		
		try {
			const res = await fetch(url, {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				s.history = await res.json();
			}
		} catch (err) {
			console.error('Failed to fetch history');
		}
	}

	async function toggleExpand(s: Server) {
		s.isExpanded = !s.isExpanded;
		if (s.isExpanded && !s.isLoadingFull) {
			s.isLoadingFull = true;
			await fetchHistory(s);
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
				const server: Server = { ...newSrv, history: [], isExpanded: false, hoveredResult: null, isLoadingFull: false };
				servers.push(server);
				isAdding = false;
				newName = '';
				newUrl = '';
				newType = 'http';
				fetchHistory(server, 20);
			}
		} catch (err) {
			error = 'Failed to add server';
		}
	}

	async function deleteServer(id: number) {
		if (!confirm('Are you sure?')) return;
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

	function getStatusColor(result: CheckResult | { status: string, latency: number }) {
		if (!result.status) return '#D62246';
		if (!(result.status.startsWith('2') || result.status === 'Connected')) return '#D62246';
		if (result.latency > 300) return '#E5B181'; // Using requested brand color
		return '#73E2A7'; // Green
	}

	// Отримуємо координати для точок графіка
	function getChartPoints(history: CheckResult[], width: number, height: number) {
		if (!history || history.length < 2) return [];
		
		const maxLatency = Math.max(...history.map(r => r.latency), 100);
		const step = width / (history.length - 1);
		
		return history.map((r, i) => ({
			x: i * step,
			y: height - (r.latency / maxLatency) * height,
			result: r
		}));
	}

	function getPathD(points: {x: number, y: number}[]) {
		if (points.length < 2) return '';
		return points.map((p, i) => (i === 0 ? `M ${p.x},${p.y}` : `L ${p.x},${p.y}`)).join(' ');
	}

	function formatDate(dateStr: string) {
		const date = new Date(dateStr);
		return new Intl.DateTimeFormat('uk-UA', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit',
			hour12: false
		}).format(date);
	}

	function handleMouseMove(e: MouseEvent, s: Server) {
		if (!s.history || s.history.length === 0) return;
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const x = e.clientX - rect.left;
		const ratio = x / rect.width;
		const index = Math.min(Math.max(Math.round(ratio * (s.history.length - 1)), 0), s.history.length - 1);
		s.hoveredResult = s.history[index];
	}

	onMount(fetchServers);
</script>

<div class="container mx-auto px-4 py-12">
	<div class="mb-8 flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold">Dashboard</h1>
			<p class="text-brand-light/60">Live status monitoring.</p>
		</div>
		<button
			onclick={() => (isAdding = !isAdding)}
			class="flex items-center gap-2 rounded-xl bg-brand-primary px-4 py-2 font-semibold text-brand-dark hover:bg-brand-primary/90 transition-all"
		>
			<Plus class="h-5 w-5" /> Add Server
		</button>
	</div>

	{#if error}
		<div class="mb-6 flex items-center gap-2 rounded-xl bg-brand-accent/10 p-4 text-brand-accent">
			<AlertCircle class="h-5 w-5" /> {error}
		</div>
	{/if}

	{#if isAdding}
		<div class="mb-8 rounded-2xl border border-brand-light/10 bg-brand-light/5 p-6 shadow-xl">
			<h2 class="mb-4 text-xl font-bold text-brand-primary">New Monitor</h2>
			<form onsubmit={addServer} class="grid gap-4 md:grid-cols-4">
				<div class="space-y-1">
					<label for="name" class="text-[10px] font-bold text-brand-light/40 uppercase tracking-widest">Name</label>
					<input id="name" type="text" bind:value={newName} required class="w-full rounded-xl border border-brand-light/10 bg-brand-dark px-4 py-2 focus:border-brand-primary outline-none" />
				</div>
				<div class="space-y-1">
					<label for="url" class="text-[10px] font-bold text-brand-light/40 uppercase tracking-widest">URL / Host</label>
					<input id="url" type="text" bind:value={newUrl} required class="w-full rounded-xl border border-brand-light/10 bg-brand-dark px-4 py-2 focus:border-brand-primary outline-none" placeholder={newType === 'http' ? 'https://example.com' : 'example.com:80'} />
				</div>
				<div class="space-y-1">
					<label for="type" class="text-[10px] font-bold text-brand-light/40 uppercase tracking-widest">Type</label>
					<select id="type" bind:value={newType} class="w-full rounded-xl border border-brand-light/10 bg-brand-dark px-4 py-2 focus:border-brand-primary outline-none">
						<option value="http">HTTP (GET)</option>
						<option value="ping">TCP Ping</option>
					</select>
				</div>
				<div class="space-y-1">
					<label for="interval" class="text-[10px] font-bold text-brand-light/40 uppercase tracking-widest">Interval</label>
					<select id="interval" bind:value={newInterval} class="w-full rounded-xl border border-brand-light/10 bg-brand-dark px-4 py-2 focus:border-brand-primary outline-none">
						<option value={30}>30s</option>
						<option value={60}>1m</option>
						<option value={300}>5m</option>
					</select>
				</div>
				<div class="flex items-end">
					<button type="submit" class="w-full rounded-xl bg-brand-primary py-2 font-bold text-brand-dark hover:bg-brand-primary/90 transition-colors">Save</button>
				</div>
			</form>
		</div>
	{/if}

	{#if isLoading}
		<div class="flex h-64 items-center justify-center"><RefreshCw class="h-8 w-8 animate-spin text-brand-primary" /></div>
	{:else if servers.length === 0}
		<div class="flex flex-col items-center justify-center rounded-3xl border border-dashed border-brand-light/20 py-20 text-center text-brand-light/20">
			<Activity class="mb-4 h-12 w-12" />
			<h3 class="text-xl font-bold">No monitors yet</h3>
		</div>
	{:else}
		<div class="grid gap-6 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3 items-start">
			{#each servers as s (s.id)}
				<div class="group rounded-2xl border {s.isExpanded ? 'border-brand-primary/40 ring-1 ring-brand-primary/10' : 'border-brand-light/10'} bg-brand-dark p-6 transition-all shadow-sm hover:shadow-xl">
					<div class="mb-4 flex items-start justify-between">
						<button onclick={() => toggleExpand(s)} class="text-left flex-1 min-w-0">
							<h3 class="font-bold flex items-center gap-1.5 truncate text-lg">
								{s.name}
								<span class="text-[8px] px-1.5 py-0.5 rounded border border-brand-light/10 text-brand-light/40 uppercase font-black tracking-tighter">{s.check_type}</span>
								{#if s.isExpanded}<ChevronUp class="h-4 w-4 opacity-40" />{:else}<ChevronDown class="h-4 w-4 opacity-40" />{/if}
							</h3>
							<p class="truncate text-xs text-brand-light/30">{s.url}</p>
						</button>
						<div class="flex gap-1">
							<a href={s.url} target="_blank" class="p-2 text-brand-light/20 hover:text-brand-light transition-colors"><ExternalLink class="h-4 w-4" /></a>
							<button onclick={() => deleteServer(s.id)} class="p-2 text-brand-light/20 hover:text-brand-accent transition-colors"><Trash2 class="h-4 w-4" /></button>
						</div>
					</div>

					<!-- Міні-графік (20 смужок) -->
					<div class="mb-6 flex gap-1 h-8">
						{#if s.history && s.history.length > 0}
							{#each s.history.slice(-20) as result}
								<div 
									class="flex-1 rounded-sm opacity-80 hover:opacity-100 hover:scale-y-110 transition-all cursor-help relative group/item"
									style="background-color: {getStatusColor(result)}"
								>
									<div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 hidden group-hover/item:block z-50 pointer-events-none">
										<div class="bg-brand-dark border border-brand-light/20 text-[10px] whitespace-nowrap p-2 rounded shadow-2xl backdrop-blur-md">
											<div class="font-bold mb-0.5">{new Date(result.created_at).toLocaleTimeString()}</div>
											<div style="color: {getStatusColor(result)}">{result.latency}ms • {result.status}</div>
										</div>
									</div>
								</div>
							{/each}
							{#if s.history.length < 20}
								{#each Array(20 - s.history.length) as _}
									<div class="flex-1 rounded-sm bg-brand-light/5"></div>
								{/each}
							{/if}
						{:else}
							{#each Array(20) as _}
								<div class="flex-1 rounded-sm bg-brand-light/5"></div>
							{/each}
						{/if}
					</div>

					<div class="flex items-center justify-between rounded-xl bg-brand-light/5 p-4 border border-brand-light/5">
						<div class="flex items-center gap-2">
							<span class="h-2 w-2 rounded-full {s.status.startsWith('2') ? 'animate-pulse' : ''}" style="background-color: {getStatusColor(s)}; box-shadow: 0 0 10px {getStatusColor(s)}"></span>
							<span class="text-xs font-black uppercase tracking-widest">
								{s.status === 'unknown' ? 'Pending' : s.status.startsWith('2') ? 'Online' : 'Error'}
							</span>
						</div>
						<div class="text-right">
							<div class="text-xl font-black text-brand-primary" style="color: {getStatusColor(s)}">{s.latency}ms</div>
							<div class="text-[9px] text-brand-light/30 font-bold uppercase">Latency</div>
						</div>
					</div>

					{#if s.isExpanded}
						<div class="mt-6 pt-6 border-t border-brand-light/10 animate-in fade-in slide-in-from-top-2 duration-300">
							<div class="mb-4 flex items-center justify-between px-1">
								<h4 class="text-[10px] font-black uppercase tracking-widest text-brand-light/40">30-Day History</h4>
								<div class="flex items-center gap-3 text-[9px] font-bold">
									<span class="flex items-center gap-1"><span class="h-2 w-2 rounded-full bg-brand-primary"></span> OK</span>
									<span class="flex items-center gap-1"><span class="h-2 w-2 rounded-full bg-[#E5B181]"></span> WARNING</span>
									<span class="flex items-center gap-1"><span class="h-2 w-2 rounded-full bg-brand-accent"></span> ERROR</span>
								</div>
							</div>

							<div 
								class="relative h-48 w-full bg-brand-light/[0.02] rounded-xl overflow-hidden cursor-crosshair border border-brand-light/5"
								onmousemove={(e) => handleMouseMove(e, s)}
								onmouseleave={() => s.hoveredResult = null}
								role="presentation"
							>
								{#if s.history && s.history.length >= 2}
									{@const points = getChartPoints(s.history, 400, 100)}
									<svg class="h-full w-full" preserveAspectRatio="none" viewBox="0 0 400 100">
										<defs>
											<linearGradient id={`grad-${s.id}`} x1="0%" y1="0%" x2="0%" y2="100%">
												<stop offset="0%" stop-color="#73E2A7" stop-opacity="0.1" />
												<stop offset="100%" stop-color="#73E2A7" stop-opacity="0" />
											</linearGradient>
											<linearGradient id={`line-grad-${s.id}`} x1="0%" y1="0%" x2="100%" y2="0%">
												{#each points as p, i}
													<stop offset="{(i / (points.length - 1)) * 100}%" stop-color={getStatusColor(p.result)} />
												{/each}
											</linearGradient>
										</defs>
										
										<!-- Area (closes at bottom correctly) -->
										<path d={`M ${points[0].x},100 ` + points.map(p => `L ${p.x},${p.y}`).join(' ') + ` L ${points[points.length-1].x},100 Z`} fill={`url(#grad-${s.id})`} />
										
										<!-- Line with dynamic colors -->
										<path d={getPathD(points)} fill="none" stroke={`url(#line-grad-${s.id})`} stroke-width="2" stroke-linejoin="round" class="opacity-80" />
										
										{#if s.hoveredResult}
											{@const maxL = Math.max(...s.history.map(r => r.latency), 100)}
											{@const hIndex = s.history.indexOf(s.hoveredResult)}
											{@const hX = (hIndex / (s.history.length - 1)) * 400}
											
											<line x1={hX} y1="0" x2={hX} y2="100" stroke="#DEF4C6" stroke-width="0.5" stroke-dasharray="2,2" opacity="0.3" />
										{/if}
									</svg>

									{#if s.hoveredResult}
										{@const maxL = Math.max(...s.history.map(r => r.latency), 100)}
										{@const hIndex = s.history.indexOf(s.hoveredResult)}
										{@const hX_p = (hIndex / (s.history.length - 1)) * 100}
										{@const hY_p = (1 - s.hoveredResult.latency / maxL) * 100}
										
										<!-- Hover Dot (HTML for perfect circle) -->
										<div 
											class="absolute w-2.5 h-2.5 rounded-full bg-brand-light/40 border border-brand-light/30 backdrop-blur-[1px] -translate-x-1/2 -translate-y-1/2 pointer-events-none z-10 transition-all duration-75"
											style="left: {hX_p}%; top: {hY_p}%"
										></div>

										<div class="absolute top-4 right-4 bg-brand-dark/95 border border-brand-light/20 p-3 rounded-xl shadow-2xl backdrop-blur-md pointer-events-none z-10 border-l-4" style="border-l-color: {getStatusColor(s.hoveredResult)}">
											<div class="text-[10px] font-bold text-brand-light/40 mb-1">{formatDate(s.hoveredResult.created_at)}</div>
											<div class="flex items-center gap-4">
												<span class="text-sm font-black" style="color: {getStatusColor(s.hoveredResult)}">{s.hoveredResult.latency}ms</span>
												<span class="text-[10px] uppercase font-bold px-2 py-0.5 rounded bg-brand-light/10 text-brand-light/60">{s.hoveredResult.status}</span>
											</div>
										</div>
									{/if}
								{:else if s.isLoadingFull}
									<div class="flex h-full items-center justify-center flex-col gap-2">
										<RefreshCw class="h-6 w-6 animate-spin text-brand-primary" />
										<span class="text-[10px] font-bold uppercase tracking-widest text-brand-light/20">Analyzing history...</span>
									</div>
								{:else}
									<div class="flex h-full items-center justify-center text-[10px] font-bold uppercase tracking-widest text-brand-light/10 italic">Collecting data...</div>
								{/if}
							</div>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
