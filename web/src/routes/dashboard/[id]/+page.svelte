<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { ArrowLeft, RefreshCw, ExternalLink, Globe, Activity, Clock, BarChart3, AlertCircle, ShieldCheck, History } from 'lucide-svelte';
	import { getStatusColor, getFaviconUrl, type Server, type CheckResult } from '$lib/utils';
	import StatusChart from '$lib/StatusChart.svelte';
	import { goto } from '$app/navigation';

	let server = $state<Server | null>(null);
	let error = $state('');
	let isLoading = $state(true);

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

	onMount(async () => {
		const id = page.params.id;
		const token = localStorage.getItem('token');
		if (!token) {
			goto('/login');
			return;
		}

		try {
			const res = await fetch(`/api/servers`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			
			if (res.ok) {
				const data = await res.json();
				const found = data.find((s: any) => s.id.toString() === id);
				
				if (found) {
					server = { ...found, history: [], history30d: [], incidents: [] };
					// Fetch 30 days for metrics
					const resHist = await fetch(`/api/servers/${id}/results?hours=720`, {
						headers: { Authorization: `Bearer ${token}` }
					});
					if (resHist.ok) {
						const dataHist = await resHist.json();
						const s = server;
						if (s) {
							s.history30d = dataHist;
							// Filter last 24h for chart
							const dayAgo = new Date(Date.now() - 24 * 60 * 60 * 1000).getTime();
							s.history = dataHist.filter((r: CheckResult) => new Date(r.created_at).getTime() > dayAgo);
						}
					}

					// Fetch last 50 incidents
					const resIncidents = await fetch(`/api/servers/${id}/results?incidents=true&limit=50`, {
						headers: { Authorization: `Bearer ${token}` }
					});
					if (resIncidents.ok) {
						const dataIncidents = await resIncidents.json();
						if (server) {
							server.incidents = dataIncidents;
						}
					}
				} else {
					error = 'Server not found';
				}
			} else {
				error = 'Failed to fetch servers';
			}
		} catch (err) {
			error = 'Connection error';
		} finally {
			isLoading = false;
		}
	});
</script>

<div class="container mx-auto px-4 py-12 max-w-5xl">
	<a href="/dashboard" class="group flex items-center gap-2 text-brand-light/40 hover:text-brand-primary mb-12 transition-all w-fit">
		<ArrowLeft class="h-4 w-4 group-hover:-translate-x-1 transition-transform" /> 
		<span class="text-sm font-bold uppercase tracking-widest">Dashboard</span>
	</a>

	{#if error}
		<div class="mb-8 flex items-center gap-4 rounded-[2rem] bg-brand-accent/10 border border-brand-accent/20 p-8 text-brand-accent animate-in fade-in slide-in-from-top-4">
			<AlertCircle class="h-8 w-8" />
			<div>
				<div class="text-xl font-bold">Analysis Failed</div>
				<div class="opacity-60">{error}</div>
			</div>
		</div>
	{:else if server}
		{@const uptime = calculateUptime(server.history30d || [])}
		{@const avgLatency = calculateAvgLatency(server.history30d || [])}
		{@const isOnline = (server.status.startsWith('2') || server.status === 'Connected')}

		<div class="flex flex-col lg:flex-row lg:items-center justify-between gap-8 mb-12">
			<div class="flex items-start gap-6">
				<div class="relative flex-shrink-0">
					<div class="h-20 w-20 rounded-[2rem] bg-brand-light/5 flex items-center justify-center border border-brand-light/10 shadow-2xl overflow-hidden">
						<img 
							src={getFaviconUrl(server.url)} 
							alt={server.name} 
							class="h-10 w-10 object-contain"
							onerror={(e) => {
								const target = e.currentTarget as HTMLImageElement;
								target.style.display = 'none';
								const parent = target.parentElement;
								const s = server;
								if (parent && s && !parent.querySelector('.fallback-icon')) {
									const icon = document.createElement('div');
									icon.className = 'fallback-icon flex items-center justify-center';
									icon.innerHTML = s.check_type === 'http' ? '<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-globe h-10 w-10 text-brand-primary/60"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>' : '<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-activity h-10 w-10 text-brand-primary/60"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>';
									parent.appendChild(icon);
								}
							}}
						/>
					</div>
					<div 
						class="absolute -bottom-1 -right-1 h-7 w-7 rounded-full border-4 border-brand-dark flex items-center justify-center"
						style="background-color: {getStatusColor(server.status, server.latency)}"
					>
						{#if isOnline}
							<div class="h-2 w-2 rounded-full bg-brand-dark animate-pulse"></div>
						{/if}
					</div>
				</div>

				<div>
					<div class="flex items-center gap-3 mb-2">
						<h1 class="text-4xl font-black tracking-tight">{server.name}</h1>
						<span class="text-[10px] px-2.5 py-1 rounded-lg bg-brand-light/5 border border-brand-light/10 text-brand-light/40 uppercase font-black tracking-widest">{server.check_type}</span>
					</div>
					<div class="flex items-center gap-2 text-brand-light/30">
						<p class="text-lg font-medium">{server.url}</p>
						<a href={server.url} target="_blank" class="p-2 bg-brand-light/5 rounded-full hover:bg-brand-primary/10 transition-colors"><ExternalLink class="h-4 w-4" /></a>
					</div>
				</div>
			</div>

			<div class="flex gap-8 bg-brand-light/[0.02] border border-brand-light/5 rounded-[2rem] p-6 backdrop-blur-sm">
				<div class="text-right">
					<div class="flex items-center justify-end gap-1.5 text-brand-light/40 text-[10px] font-bold uppercase tracking-widest mb-1">
						<Clock class="h-3 w-3" /> 30d Uptime
					</div>
					<div class="text-3xl font-black {uptime >= 99 ? 'text-brand-primary' : uptime >= 95 ? 'text-brand-soft' : 'text-brand-accent'}">
						{uptime.toFixed(2)}%
					</div>
				</div>
				<div class="w-px h-10 bg-brand-light/10 self-center"></div>
				<div class="text-right min-w-[100px]">
					<div class="flex items-center justify-end gap-1.5 text-brand-light/40 text-[10px] font-bold uppercase tracking-widest mb-1">
						<BarChart3 class="h-3 w-3" /> 30d Avg
					</div>
					<div class="text-3xl font-black text-brand-light/80">
						{avgLatency}<span class="text-xs font-bold text-brand-light/20 ml-0.5">ms</span>
					</div>
				</div>
			</div>
		</div>

		<div class="grid gap-8">
			<!-- Chart Card -->
			<div class="rounded-[2.5rem] border border-brand-light/10 bg-gradient-to-b from-brand-light/[0.03] to-transparent p-1 shadow-2xl">
				<div class="bg-brand-dark rounded-[2.4rem] p-8 lg:p-10">
					<div class="mb-10 flex flex-col sm:flex-row sm:items-center justify-between gap-6">
						<div>
							<h3 class="text-xl font-bold flex items-center gap-2 mb-1">
								<Activity class="h-5 w-5 text-brand-primary" />
								Performance Metrics
							</h3>
							<p class="text-sm text-brand-light/30">Visualizing latency and availability over the last 24 hours.</p>
						</div>
						<div class="flex flex-wrap items-center gap-4 text-[10px] font-black uppercase tracking-widest px-4 py-2 bg-brand-light/5 rounded-2xl border border-brand-light/10">
							<span class="flex items-center gap-1.5"><span class="h-2 w-2 rounded-full bg-brand-primary shadow-[0_0_8px_rgba(115,226,167,0.5)]"></span> Optimal</span>
							<span class="flex items-center gap-1.5"><span class="h-2 w-2 rounded-full bg-[#E5B181]"></span> High Latency</span>
							<span class="flex items-center gap-1.5"><span class="h-2 w-2 rounded-full bg-brand-accent"></span> Critical</span>
						</div>
					</div>

					<div class="relative">
						<StatusChart history={server.history} height={500} />
						<div class="absolute bottom-4 left-4 flex gap-4 text-[10px] font-bold text-brand-light/20 uppercase tracking-widest">
							<span>&larr; 24 hours ago</span>
							<span>Peak: {Math.max(...server.history.map((r: CheckResult) => r.latency), 0)}ms</span>
						</div>
					</div>
				</div>
			</div>

			<!-- Logs Card -->
			<div class="rounded-[2.5rem] border border-brand-light/10 bg-brand-dark/40 p-8 lg:p-10">
				<div class="mb-8 flex items-center justify-between">
					<h3 class="text-xl font-bold flex items-center gap-2">
						<History class="h-5 w-5 text-brand-primary" />
						Incidents Log
					</h3>
					<span class="text-[10px] font-black uppercase tracking-widest px-3 py-1 bg-brand-light/5 rounded-full text-brand-light/40">
						Showing last 50 incidents
					</span>
				</div>

				<div class="overflow-hidden rounded-3xl border border-brand-light/5 bg-brand-light/[0.01]">
					<div class="grid grid-cols-12 px-6 py-4 border-b border-brand-light/10 text-[10px] font-black uppercase tracking-widest text-brand-light/20 bg-brand-light/[0.02]">
						<div class="col-span-1">Status</div>
						<div class="col-span-5 sm:col-span-6">Timestamp</div>
						<div class="col-span-3 sm:col-span-3">Check Response</div>
						<div class="col-span-3 sm:col-span-2 text-right">Latency</div>
					</div>
					<div class="divide-y divide-brand-light/5 max-h-[600px] overflow-y-auto custom-scrollbar">
						{#if server.incidents && server.incidents.length > 0}
							{#each server.incidents as result}
								<div class="grid grid-cols-12 px-6 py-4 items-center hover:bg-brand-light/[0.02] transition-colors group">
									<div class="col-span-1">
										<div
											class="h-3 w-3 rounded-full border-2 border-brand-dark shadow-sm"
											style="background-color: {getStatusColor(result.status, result.latency)}"
										></div>
									</div>
									<div class="col-span-5 sm:col-span-6">
										<span class="text-sm font-medium text-brand-light/80">{new Date(result.created_at).toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' })}</span>
									</div>
									<div class="col-span-3 sm:col-span-3">
										<span class="text-[10px] font-black uppercase px-2 py-0.5 rounded-md bg-brand-light/5 text-brand-light/40 border border-brand-light/10 group-hover:text-brand-light/60 transition-colors">
											{result.status === 'Connected' ? 'Online' : result.status}
										</span>
									</div>
									<div class="col-span-3 sm:col-span-2 text-right">
										<span class="text-sm font-black font-mono tracking-tight" style="color: {getStatusColor(result.status, result.latency)}">
											{result.latency}<span class="text-[10px] ml-0.5 opacity-40">ms</span>
										</span>
									</div>
								</div>
							{/each}
						{:else}
							<div class="flex flex-col items-center justify-center py-20 text-brand-light/10 gap-4">
								<ShieldCheck class="h-12 w-12 opacity-20" />
								<p class="font-bold uppercase tracking-widest text-xs">No incidents recorded yet!</p>
							</div>
						{/if}
					</div>
				</div>
			</div>

		</div>
	{:else}
		<div class="flex h-[60vh] flex-col items-center justify-center gap-6 animate-pulse">
			<RefreshCw class="h-16 w-16 animate-spin text-brand-primary opacity-20" />
			<div class="text-center">
				<h3 class="text-lg font-black uppercase tracking-widest text-brand-light/20">Analyzing Infrastructure</h3>
				<p class="text-sm text-brand-light/10">Establishing secure connection to telemetry...</p>
			</div>
		</div>
	{/if}
</div>

<style>
	/* Custom animations */
	@keyframes fade-in { from { opacity: 0; } to { opacity: 1; } }
	@keyframes slide-in-from-top-4 { from { transform: translateY(-1rem); } to { transform: translateY(0); } }

	.animate-in {
		animation-duration: 400ms;
		animation-fill-mode: both;
	}
	.fade-in { animation-name: fade-in; }
	.slide-in-from-top-4 { animation-name: slide-in-from-top-4; }

	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: rgba(222, 244, 198, 0.05);
		border-radius: 10px;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: rgba(222, 244, 198, 0.1);
	}
</style>
