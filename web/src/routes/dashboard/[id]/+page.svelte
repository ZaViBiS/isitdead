<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import {
		ArrowLeft,
		RefreshCw,
		ExternalLink,
		Activity,
		Clock,
		BarChart3,
		AlertCircle,
		ShieldCheck,
		History
	} from 'lucide-svelte';
	import {
		getStatusColor,
		getFaviconUrl,
		getCurrentCheck,
		sampleChartHistory,
		type Server,
		type CheckResult
	} from '$lib/utils';
	import StatusChart from '$lib/StatusChart.svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	let server = $state<Server | null>(null);
	let error = $state('');
	let isLoading = $state(true);

	onMount(async () => {
		const id = page.params.id;
		const token = localStorage.getItem('token');
		if (!token) {
			goto(resolve('/login'));
			return;
		}

		try {
			const res = await fetch(`/api/dashboard/servers`, {
				headers: { Authorization: `Bearer ${token}` }
			});

			if (res.ok) {
				const data = (await res.json()) as Server[];
				const found = data.find((s) => s.id.toString() === id);

				if (found) {
					server = { ...found, history: [], history30d: [], incidents: [] };
					const resHist = await fetch(`/api/servers/${id}/results?hours=72`, {
						headers: { Authorization: `Bearer ${token}` }
					});
					if (resHist.ok) {
						const dataHist = (await resHist.json()) as CheckResult[];
						const s = server;
						if (s) {
							s.history = sampleChartHistory(dataHist, s.slow_threshold);
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
					error = 'Monitor not found';
				}
			} else {
				error = 'Failed to load monitors';
			}
		} catch {
			error = 'Connection error';
		} finally {
			isLoading = false;
		}
	});
</script>

<div class="container mx-auto max-w-5xl px-4 py-8 sm:py-12">
	<a
		href={resolve('/dashboard')}
		class="group mb-12 flex w-fit items-center gap-2 text-brand-light/40 transition-all hover:text-brand-primary"
	>
		<ArrowLeft class="h-4 w-4 transition-transform group-hover:-translate-x-1" />
		<span class="text-sm font-bold tracking-widest uppercase">All monitors</span>
	</a>

	{#if error}
		<div
			class="animate-in fade-in slide-in-from-top-4 mb-8 flex items-center gap-4 rounded-[2rem] border border-brand-accent/20 bg-brand-accent/10 p-8 text-brand-accent"
		>
			<AlertCircle class="h-8 w-8" />
			<div>
				<div class="text-xl font-bold">Could not load monitor</div>
				<div class="opacity-60">{error}</div>
			</div>
		</div>
	{:else if server}
		{@const uptime = server.uptime_30d ?? 0}
		{@const avgLatency = server.avg_latency_30d ?? 0}
		{@const current = getCurrentCheck(server)}
		{@const currentStatus = current?.status ?? 'unknown'}
		{@const currentLatency = current?.latency ?? 0}
		{@const isOnline = currentStatus.startsWith('2') || currentStatus === 'Connected'}

		<div class="mb-8 flex flex-col justify-between gap-6 sm:mb-12 lg:flex-row lg:items-center">
			<div class="flex items-start gap-4 sm:gap-6">
				<div class="relative flex-shrink-0">
					<div
						class="flex h-20 w-20 items-center justify-center overflow-hidden rounded-[2rem] border border-brand-light/10 bg-brand-light/5 shadow-2xl"
					>
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
									icon.innerHTML =
										s.check_type === 'http'
											? '<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-globe h-10 w-10 text-brand-primary/60"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>'
											: '<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-activity h-10 w-10 text-brand-primary/60"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>';
									parent.appendChild(icon);
								}
							}}
						/>
					</div>
					<div
						class="absolute -right-1 -bottom-1 flex h-7 w-7 items-center justify-center rounded-full border-4 border-brand-dark"
						style="background-color: {getStatusColor(
							currentStatus,
							currentLatency,
							server.slow_threshold
						)}"
					>
						{#if isOnline}
							<div class="h-2 w-2 animate-pulse rounded-full bg-brand-dark"></div>
						{/if}
					</div>
				</div>

				<div class="min-w-0">
					<div class="mb-2 flex flex-wrap items-center gap-3">
						<h1 class="text-3xl font-black tracking-tight break-words sm:text-4xl">
							{server.name}
						</h1>
						<span
							class="rounded-lg border border-brand-light/10 bg-brand-light/5 px-2.5 py-1 text-[10px] font-black tracking-widest text-brand-light/40 uppercase"
							>{server.check_type}</span
						>
					</div>
					<div class="flex min-w-0 items-start gap-2 text-brand-light/30">
						<p class="min-w-0 text-sm font-medium break-all sm:text-lg">{server.url}</p>
						<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
						<a
							href={server.url}
							target="_blank"
							rel="external noreferrer"
							class="rounded-full bg-brand-light/5 p-2 transition-colors hover:bg-brand-primary/10"
							><ExternalLink class="h-4 w-4" /></a
						>
					</div>
				</div>
			</div>

			<div class="glass-panel grid gap-4 rounded-[2rem] p-5 sm:flex sm:gap-8 sm:p-6">
				<div class="text-left sm:text-right">
					<div
						class="mb-1 flex items-center justify-end gap-1.5 text-[10px] font-bold tracking-widest text-brand-light/40 uppercase"
					>
						<Clock class="h-3 w-3" /> 30d Uptime
					</div>
					<div
						class="text-3xl font-black {uptime >= 99
							? 'text-brand-primary'
							: uptime >= 95
								? 'text-brand-soft'
								: 'text-brand-accent'}"
					>
						{uptime.toFixed(2)}%
					</div>
				</div>
				<div class="hidden h-10 w-px self-center bg-brand-light/10 sm:block"></div>
				<div class="min-w-0 text-left sm:min-w-[100px] sm:text-right">
					<div
						class="mb-1 flex items-center justify-end gap-1.5 text-[10px] font-bold tracking-widest text-brand-light/40 uppercase"
					>
						<BarChart3 class="h-3 w-3" /> 30d Avg
					</div>
					<div class="text-3xl font-black text-brand-light/80">
						{avgLatency}<span class="ml-0.5 text-xs font-bold text-brand-light/20">ms</span>
					</div>
				</div>
			</div>
		</div>

		<div class="grid gap-8">
			<!-- Chart Card -->
			<div class="glass-panel rounded-[2.5rem] p-1">
				<div class="rounded-[2.4rem] bg-brand-dark p-5 sm:p-8 lg:p-10">
					<div class="mb-10 flex flex-col justify-between gap-6 sm:flex-row sm:items-center">
						<div>
							<h3 class="mb-1 flex items-center gap-2 text-xl font-bold">
								<Activity class="h-5 w-5 text-brand-primary" />
								Response time
							</h3>
							<p class="text-sm text-brand-light/30">
								Latency and availability over the last 3 days.
							</p>
						</div>
						<div
							class="flex flex-wrap items-center gap-4 rounded-2xl border border-brand-light/10 bg-brand-light/5 px-4 py-2 text-[10px] font-black tracking-widest uppercase"
						>
							<span class="flex items-center gap-1.5"
								><span
									class="h-2 w-2 rounded-full bg-brand-primary shadow-[0_0_8px_rgba(115,226,167,0.5)]"
								></span> Healthy</span
							>
							<span class="flex items-center gap-1.5"
								><span class="h-2 w-2 rounded-full bg-[#E5B181]"></span> Slow &gt;
								{server.slow_threshold}ms</span
							>
							<span class="flex items-center gap-1.5"
								><span class="h-2 w-2 rounded-full bg-brand-accent"></span> Down</span
							>
						</div>
					</div>

					<div class="relative">
						<StatusChart
							history={server.history}
							height={500}
							slowThreshold={server.slow_threshold}
						/>
						<div
							class="absolute bottom-4 left-4 flex flex-wrap gap-2 text-[10px] font-bold tracking-widest text-brand-light/20 uppercase sm:gap-4"
						>
							<span>&larr; 3 days ago</span>
							<span
								>Peak: {Math.max(...server.history.map((r: CheckResult) => r.latency), 0)}ms</span
							>
						</div>
					</div>
				</div>
			</div>

			<!-- Logs Card -->
			<div class="glass-panel rounded-[2.5rem] p-5 sm:p-8 lg:p-10">
				<div
					class="mb-6 flex flex-col gap-3 sm:mb-8 sm:flex-row sm:items-center sm:justify-between"
				>
					<h3 class="flex items-center gap-2 text-xl font-bold">
						<History class="h-5 w-5 text-brand-primary" />
						Incidents
					</h3>
					<span
						class="rounded-full bg-brand-light/5 px-3 py-1 text-[10px] font-black tracking-widest text-brand-light/40 uppercase"
					>
						Last 50 incidents
					</span>
				</div>

				<div class="overflow-hidden rounded-3xl border border-brand-light/5 bg-brand-light/[0.01]">
					<div
						class="hidden grid-cols-12 border-b border-brand-light/10 bg-brand-light/[0.02] px-6 py-4 text-[10px] font-black tracking-widest text-brand-light/20 uppercase sm:grid"
					>
						<div class="col-span-1">Status</div>
						<div class="col-span-5 sm:col-span-6">Timestamp</div>
						<div class="col-span-3 sm:col-span-3">Response</div>
						<div class="col-span-3 text-right sm:col-span-2">Latency</div>
					</div>
					<div class="custom-scrollbar max-h-[600px] divide-y divide-brand-light/5 overflow-y-auto">
						{#if server.incidents && server.incidents.length > 0}
							{#each server.incidents as result (result.id)}
								<div
									class="group grid gap-3 px-4 py-4 transition-colors hover:bg-brand-light/[0.02] sm:grid-cols-12 sm:items-center sm:px-6"
								>
									<div class="flex items-center justify-between gap-3 sm:col-span-1 sm:block">
										<span
											class="text-[10px] font-black tracking-widest text-brand-light/20 uppercase sm:hidden"
											>Status</span
										>
										<div
											class="h-3 w-3 rounded-full border-2 border-brand-dark shadow-sm"
											style="background-color: {getStatusColor(
												result.status,
												result.latency,
												server.slow_threshold
											)}"
										></div>
									</div>
									<div class="flex items-center justify-between gap-3 sm:col-span-6 sm:block">
										<span
											class="text-[10px] font-black tracking-widest text-brand-light/20 uppercase sm:hidden"
											>Timestamp</span
										>
										<span class="text-sm font-medium text-brand-light/80"
											>{new Date(result.created_at).toLocaleString('en-US', {
												dateStyle: 'medium',
												timeStyle: 'short'
											})}</span
										>
									</div>
									<div class="flex items-center justify-between gap-3 sm:col-span-3 sm:block">
										<span
											class="text-[10px] font-black tracking-widest text-brand-light/20 uppercase sm:hidden"
											>Response</span
										>
										<span
											class="rounded-md border border-brand-light/10 bg-brand-light/5 px-2 py-0.5 text-[10px] font-black text-brand-light/40 uppercase transition-colors group-hover:text-brand-light/60"
										>
											{result.status === 'Connected' ? 'Online' : result.status}
										</span>
									</div>
									<div
										class="flex items-center justify-between gap-3 sm:col-span-2 sm:block sm:text-right"
									>
										<span
											class="text-[10px] font-black tracking-widest text-brand-light/20 uppercase sm:hidden"
											>Latency</span
										>
										<span
											class="font-mono text-sm font-black tracking-tight"
											style="color: {getStatusColor(
												result.status,
												result.latency,
												server.slow_threshold
											)}"
										>
											{result.latency}<span class="ml-0.5 text-[10px] opacity-40">ms</span>
										</span>
									</div>
								</div>
							{/each}
						{:else}
							<div
								class="flex flex-col items-center justify-center gap-4 py-20 text-brand-light/10"
							>
								<ShieldCheck class="h-12 w-12 opacity-20" />
								<p class="text-xs font-bold tracking-widest uppercase">No incidents yet</p>
							</div>
						{/if}
					</div>
				</div>
			</div>
		</div>
	{:else if isLoading}
		<div class="flex h-[60vh] animate-pulse flex-col items-center justify-center gap-6">
			<RefreshCw class="h-16 w-16 animate-spin text-brand-primary opacity-20" />
			<div class="text-center">
				<h3 class="text-lg font-black tracking-widest text-brand-light/20 uppercase">
					Loading monitor
				</h3>
				<p class="text-sm text-brand-light/10">Fetching latest checks...</p>
			</div>
		</div>
	{:else}
		<div class="flex h-[60vh] flex-col items-center justify-center gap-4 text-center">
			<AlertCircle class="h-12 w-12 text-brand-light/15" />
			<h3 class="text-lg font-black tracking-widest text-brand-light/25 uppercase">
				Monitor unavailable
			</h3>
			<p class="max-w-sm text-sm text-brand-light/20">
				Go back to your monitors and choose an active service.
			</p>
		</div>
	{/if}
</div>

<style>
	/* Custom animations */
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

	.animate-in {
		animation-duration: 400ms;
		animation-fill-mode: both;
	}
	.fade-in {
		animation-name: fade-in;
	}
	.slide-in-from-top-4 {
		animation-name: slide-in-from-top-4;
	}

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
