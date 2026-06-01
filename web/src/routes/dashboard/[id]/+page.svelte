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
		History,
		Globe2,
		LockKeyhole
	} from 'lucide-svelte';
	import {
		calculateAvgLatency,
		calculateUptime,
		getStatusColor,
		getFaviconUrl,
		getCurrentCheck,
		getEffectiveSlowThreshold,
		formatDateTime,
		supportsSlowThreshold,
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
	let regionalHistory = $state<Record<string, CheckResult[]>>({});
	let regionalIncidents = $state<Record<string, CheckResult[]>>({});
	let selectedRegion = $state('global');
	let faviconLoadFailed = $state(false);

	function safeExternalHref(rawUrl: string) {
		try {
			const parsed = new URL(rawUrl);
			return parsed.protocol === 'http:' || parsed.protocol === 'https:' ? parsed.href : '';
		} catch {
			return '';
		}
	}

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
					faviconLoadFailed = false;
					server = { ...found, history: [], history30d: [], incidents: [] };
					const [resHist, resRegions] = await Promise.all([
						fetch(`/api/servers/${id}/results?hours=72`, {
							headers: { Authorization: `Bearer ${token}` }
						}),
						fetch(`/api/servers/${id}/results?hours=72&region=all`, {
							headers: { Authorization: `Bearer ${token}` }
						})
					]);
					if (resHist.ok || resRegions.ok) {
						const dataHist = resHist.ok ? ((await resHist.json()) as CheckResult[]) : [];
						const dataRegions = resRegions.ok ? ((await resRegions.json()) as CheckResult[]) : [];
						regionalHistory = groupResultsByRegion([...dataHist, ...dataRegions]);
						const s = server;
						if (s) {
							s.history = sampleChartHistory(
								regionalHistory.global ?? [],
								getEffectiveSlowThreshold(s.check_type, s.slow_threshold)
							);
						}
					}

					// Fetch last 50 incidents
					const [resIncidents, resRegionIncidents] = await Promise.all([
						fetch(`/api/servers/${id}/results?incidents=true&limit=50`, {
							headers: { Authorization: `Bearer ${token}` }
						}),
						fetch(`/api/servers/${id}/results?incidents=true&limit=50&region=all`, {
							headers: { Authorization: `Bearer ${token}` }
						})
					]);
					if (resIncidents.ok || resRegionIncidents.ok) {
						const dataIncidents = resIncidents.ok
							? ((await resIncidents.json()) as CheckResult[])
							: [];
						const dataRegionIncidents = resRegionIncidents.ok
							? ((await resRegionIncidents.json()) as CheckResult[])
							: [];
						regionalIncidents = groupResultsByRegion(
							[...dataIncidents, ...dataRegionIncidents],
							'desc'
						);
						if (server) {
							server.incidents = regionalIncidents.global ?? [];
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

	function resultRegion(result: CheckResult) {
		return result.region?.trim() || 'global';
	}

	function groupResultsByRegion(results: CheckResult[], order: 'asc' | 'desc' = 'asc') {
		const grouped: Record<string, CheckResult[]> = {};
		for (const result of results) {
			const region = resultRegion(result);
			if (!grouped[region]) grouped[region] = [];
			grouped[region].push(result);
		}
		for (const region of Object.keys(grouped)) {
			grouped[region].sort(
				(a, b) =>
					(new Date(a.created_at).getTime() - new Date(b.created_at).getTime()) *
					(order === 'asc' ? 1 : -1)
			);
		}
		return grouped;
	}

	function getRegionNames() {
		const names = Object.keys(regionalHistory).filter((region) => region !== 'global');
		return names.sort((a, b) => {
			return a.localeCompare(b);
		});
	}

	function getViewNames() {
		const names = getRegionNames();
		return regionalHistory.global ? ['global', ...names] : names;
	}

	function getActiveRegion() {
		const names = getViewNames();
		if (names.includes(selectedRegion)) return selectedRegion;
		return names[0] ?? selectedRegion;
	}

	function displayRegion(region: string) {
		if (region === 'global') return 'Overall';
		return region.toUpperCase();
	}

	function getRegionHistory(region: string) {
		return regionalHistory[region] ?? [];
	}

	function getSelectedHistory(region: string) {
		return getRegionHistory(region);
	}

	function getSelectedChartHistory(region: string, slowThreshold: number) {
		return sampleChartHistory(getSelectedHistory(region), slowThreshold);
	}

	function getSelectedIncidents(region: string) {
		return (regionalIncidents[region] ?? []).slice(0, 50);
	}

	function isHealthyStatus(status: string) {
		return status.startsWith('2') || status === 'Connected';
	}

	function regionSummary(region: string) {
		const history = getRegionHistory(region);
		const current = history.at(-1);
		return {
			name: region,
			current,
			uptime: calculateUptime(history),
			avgLatency: calculateAvgLatency(history),
			count: history.length,
			online: current ? isHealthyStatus(current.status) : false
		};
	}
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
		{@const effectiveSlowThreshold = getEffectiveSlowThreshold(
			server.check_type,
			server.slow_threshold
		)}
		{@const regionNames = getRegionNames()}
		{@const viewNames = getViewNames()}
		{@const activeRegion = getActiveRegion()}
		{@const activeSummary = regionSummary(activeRegion)}
		{@const selectedHistory = getSelectedChartHistory(activeRegion, effectiveSlowThreshold)}
		{@const selectedIncidents = getSelectedIncidents(activeRegion)}

		<div class="mb-8 flex flex-col justify-between gap-6 sm:mb-12 lg:flex-row lg:items-center">
			<div class="flex items-start gap-4 sm:gap-6">
				<div class="relative flex-shrink-0">
					<div
						class="flex h-20 w-20 items-center justify-center overflow-hidden rounded-[2rem] border border-brand-light/10 bg-brand-light/5 shadow-2xl"
					>
						{#if faviconLoadFailed}
							{#if server.check_type === 'ssl'}
								<LockKeyhole class="h-10 w-10 text-brand-primary/60" />
							{:else if server.check_type === 'http' || server.check_type === 'links'}
								<Globe2 class="h-10 w-10 text-brand-primary/60" />
							{:else}
								<Activity class="h-10 w-10 text-brand-primary/60" />
							{/if}
						{:else}
							<img
								src={getFaviconUrl(server.url)}
								alt={server.name}
								class="h-10 w-10 object-contain"
								onerror={() => (faviconLoadFailed = true)}
							/>
						{/if}
					</div>
					<div
						class="absolute -right-1 -bottom-1 flex h-7 w-7 items-center justify-center rounded-full border-4 border-brand-dark"
						style="background-color: {getStatusColor(
							currentStatus,
							currentLatency,
							effectiveSlowThreshold
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
						{#if safeExternalHref(server.url)}
							<a
								href={safeExternalHref(server.url)}
								target="_blank"
								rel="external noreferrer"
								class="rounded-full bg-brand-light/5 p-2 transition-colors hover:bg-brand-primary/10"
								><ExternalLink class="h-4 w-4" /></a
							>
						{/if}
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
			{#if viewNames.length > 0}
				<section class="grid gap-4 lg:grid-cols-[minmax(0,0.85fr)_minmax(0,1.4fr)]">
					{#if regionalHistory.global}
						{@const summary = regionSummary('global')}
						{@const overallStatus = summary.current?.status ?? 'unknown'}
						{@const overallLatency = summary.current?.latency ?? 0}
						<button
							type="button"
							aria-pressed={activeRegion === 'global'}
							onclick={() => (selectedRegion = 'global')}
							class="glass-panel rounded-[2rem] p-5 text-left transition sm:p-6 {activeRegion ===
							'global'
								? 'border-brand-primary/35 bg-brand-primary/10 shadow-2xl shadow-brand-primary/5'
								: 'hover:border-brand-primary/25'}"
						>
							<div class="mb-8 flex items-start justify-between gap-4">
								<div>
									<div class="mb-2 flex items-center gap-2 text-xl font-black">
										<Globe2 class="h-5 w-5 text-brand-primary" />
										Overall
									</div>
									<div class="text-sm font-medium text-brand-light/35">
										{summary.current ? formatDateTime(summary.current.created_at) : 'No checks yet'}
									</div>
								</div>
								<span
									class="h-4 w-4 shrink-0 rounded-full border-2 border-brand-dark"
									style="background-color: {getStatusColor(
										overallStatus,
										overallLatency,
										effectiveSlowThreshold
									)}"
								></span>
							</div>

							<div class="mb-7">
								<div
									class="text-3xl font-black {summary.online
										? 'text-brand-primary'
										: summary.current
											? 'text-brand-accent'
											: 'text-brand-light/35'}"
								>
									{summary.online ? 'Operational' : summary.current ? 'Incident' : 'No data'}
								</div>
								<div class="mt-1 text-xs font-bold tracking-widest text-brand-light/25 uppercase">
									{summary.count} checks
								</div>
							</div>

							<div class="grid grid-cols-2 gap-3">
								<div class="rounded-2xl border border-brand-light/5 bg-brand-light/[0.025] p-3">
									<div class="font-mono text-lg font-black text-brand-light/85">
										{summary.avgLatency}<span class="ml-0.5 text-[10px] text-brand-light/30"
											>ms</span
										>
									</div>
									<div class="text-[10px] font-bold text-brand-light/25 uppercase">avg</div>
								</div>
								<div class="rounded-2xl border border-brand-light/5 bg-brand-light/[0.025] p-3">
									<div class="text-lg font-black text-brand-light/85">
										{summary.uptime.toFixed(1)}<span class="ml-0.5 text-[10px] text-brand-light/30"
											>%</span
										>
									</div>
									<div class="text-[10px] font-bold text-brand-light/25 uppercase">uptime</div>
								</div>
							</div>
						</button>
					{/if}

					{#if regionNames.length > 0}
						<div class="glass-panel rounded-[2rem] p-5 sm:p-6">
							<div class="mb-5 flex items-center justify-between gap-4">
								<h2 class="flex items-center gap-2 text-xl font-black">
									<Activity class="h-5 w-5 text-brand-primary" />
									Probes
								</h2>
								<span
									class="rounded-full bg-brand-light/5 px-3 py-1 text-[10px] font-black tracking-widest text-brand-light/40 uppercase"
								>
									{regionNames.length} regions
								</span>
							</div>

							<div class="grid gap-3 sm:grid-cols-2">
								{#each regionNames as region (region)}
									{@const summary = regionSummary(region)}
									{@const regionStatus = summary.current?.status ?? 'unknown'}
									{@const regionLatency = summary.current?.latency ?? 0}
									<button
										type="button"
										aria-pressed={activeRegion === region}
										onclick={() => (selectedRegion = region)}
										class="rounded-3xl border p-4 text-left transition {activeRegion === region
											? 'border-brand-primary/35 bg-brand-primary/10 shadow-2xl shadow-brand-primary/5'
											: 'border-brand-light/10 bg-brand-light/[0.025] hover:border-brand-primary/25'}"
									>
										<div class="mb-4 flex items-start justify-between gap-3">
											<div class="min-w-0">
												<div class="truncate text-lg font-black">{displayRegion(region)}</div>
												<div
													class="mt-1 text-[10px] font-bold tracking-widest text-brand-light/30 uppercase"
												>
													{summary.current ? formatDateTime(summary.current.created_at) : 'No data'}
												</div>
											</div>
											<span
												class="h-3 w-3 shrink-0 rounded-full"
												style="background-color: {getStatusColor(
													regionStatus,
													regionLatency,
													effectiveSlowThreshold
												)}"
											></span>
										</div>
										<div class="grid grid-cols-3 gap-3">
											<div>
												<div
													class="text-sm font-black {summary.online
														? 'text-brand-primary'
														: summary.current
															? 'text-brand-accent'
															: 'text-brand-light/35'}"
												>
													{summary.online ? 'Online' : summary.current ? 'Down' : 'No data'}
												</div>
												<div class="text-[10px] font-bold text-brand-light/25 uppercase">
													status
												</div>
											</div>
											<div>
												<div class="font-mono text-sm font-black text-brand-light/80">
													{summary.current ? `${summary.avgLatency}ms` : '-'}
												</div>
												<div class="text-[10px] font-bold text-brand-light/25 uppercase">avg</div>
											</div>
											<div class="text-right">
												<div class="text-sm font-black text-brand-light/80">
													{summary.count}
												</div>
												<div class="text-[10px] font-bold text-brand-light/25 uppercase">
													checks
												</div>
											</div>
										</div>
										<div class="mt-4 h-1.5 overflow-hidden rounded-full bg-brand-light/5">
											<div
												class="h-full rounded-full {summary.uptime >= 99
													? 'bg-brand-primary'
													: summary.uptime >= 95
														? 'bg-brand-soft'
														: 'bg-brand-accent'}"
												style="width: {summary.count > 0 ? summary.uptime : 0}%"
											></div>
										</div>
									</button>
								{/each}
							</div>
						</div>
					{/if}
				</section>
			{/if}

			<!-- Chart Card -->
			<div class="glass-panel rounded-[2.5rem] p-1">
				<div class="rounded-[2.4rem] bg-brand-dark p-5 sm:p-8 lg:p-10">
					<div class="mb-10 flex flex-col justify-between gap-6 lg:flex-row lg:items-start">
						<div>
							<div class="mb-1 flex flex-wrap items-center gap-2">
								<h3 class="flex items-center gap-2 text-xl font-bold">
									<Activity class="h-5 w-5 text-brand-primary" />
									Response time
								</h3>
								<span
									class="rounded-full border border-brand-primary/20 bg-brand-primary/10 px-2.5 py-1 text-[10px] font-black tracking-widest text-brand-primary uppercase"
								>
									{displayRegion(activeRegion)}
								</span>
							</div>
							<p class="text-sm text-brand-light/30">
								{activeSummary.current
									? `${activeSummary.avgLatency}ms average · ${activeSummary.uptime.toFixed(1)}% uptime`
									: 'Collecting telemetry data'}
							</p>
						</div>
						<div class="flex flex-col gap-3 lg:items-end">
							{#if viewNames.length > 1}
								<div
									class="inline-flex w-fit flex-wrap gap-1 rounded-2xl border border-brand-light/10 bg-brand-light/[0.035] p-1"
								>
									{#each viewNames as region (region)}
										<button
											type="button"
											aria-pressed={activeRegion === region}
											onclick={() => (selectedRegion = region)}
											class="inline-flex h-9 items-center gap-2 rounded-xl px-3 text-[10px] font-black tracking-widest uppercase transition {activeRegion ===
											region
												? 'bg-brand-primary text-brand-dark shadow-lg shadow-brand-primary/10'
												: 'text-brand-light/40 hover:bg-brand-light/5 hover:text-brand-light/75'}"
										>
											{#if region === 'global'}
												<Globe2 class="h-3.5 w-3.5" />
											{:else}
												<Activity class="h-3.5 w-3.5" />
											{/if}
											{displayRegion(region)}
										</button>
									{/each}
								</div>
							{/if}
							<div
								class="flex flex-wrap items-center gap-4 rounded-2xl border border-brand-light/10 bg-brand-light/5 px-4 py-2 text-[10px] font-black tracking-widest uppercase"
							>
								<span class="flex items-center gap-1.5"
									><span
										class="h-2 w-2 rounded-full bg-[#50FA7B] shadow-[0_0_8px_rgba(80,250,123,0.5)]"
									></span> Healthy</span
								>
								{#if supportsSlowThreshold(server.check_type)}
									<span class="flex items-center gap-1.5"
										><span class="h-2 w-2 rounded-full bg-[#F0AD4E]"></span> Slow &gt;
										{server.slow_threshold}ms</span
									>
								{/if}
								<span class="flex items-center gap-1.5"
									><span class="h-2 w-2 rounded-full bg-brand-accent"></span> Down</span
								>
							</div>
						</div>
					</div>

					<div class="relative">
						<StatusChart
							history={selectedHistory}
							height={500}
							slowThreshold={effectiveSlowThreshold}
						/>
						<div
							class="absolute bottom-4 left-4 flex flex-wrap gap-2 text-[10px] font-bold tracking-widest text-brand-light/20 uppercase sm:gap-4"
						>
							<span>&larr; 3 days ago</span>
							<span
								>Peak: {Math.max(...selectedHistory.map((r: CheckResult) => r.latency), 0)}ms</span
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
						Incidents · {displayRegion(activeRegion)}
					</h3>
					<span
						class="rounded-full bg-brand-light/5 px-3 py-1 text-[10px] font-black tracking-widest text-brand-light/40 uppercase"
					>
						Last {selectedIncidents.length} incidents
					</span>
				</div>

				<div class="overflow-hidden rounded-3xl border border-brand-light/5 bg-brand-light/[0.01]">
					<div
						class="hidden grid-cols-12 border-b border-brand-light/10 bg-brand-light/[0.02] px-6 py-4 text-[10px] font-black tracking-widest text-brand-light/20 uppercase sm:grid"
					>
						<div class="col-span-1">Status</div>
						<div class="col-span-2">Region</div>
						<div class="col-span-4">Timestamp</div>
						<div class="col-span-3">Response</div>
						<div class="col-span-3 text-right sm:col-span-2">Latency</div>
					</div>
					<div class="custom-scrollbar max-h-[600px] divide-y divide-brand-light/5 overflow-y-auto">
						{#if selectedIncidents.length > 0}
							{#each selectedIncidents as result (result.id)}
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
												effectiveSlowThreshold
											)}"
										></div>
									</div>
									<div class="flex items-center justify-between gap-3 sm:col-span-2 sm:block">
										<span
											class="text-[10px] font-black tracking-widest text-brand-light/20 uppercase sm:hidden"
											>Region</span
										>
										<span
											class="rounded-md border border-brand-light/10 bg-brand-light/5 px-2 py-0.5 text-[10px] font-black text-brand-light/50 uppercase"
										>
											{displayRegion(resultRegion(result))}
										</span>
									</div>
									<div class="flex items-center justify-between gap-3 sm:col-span-4 sm:block">
										<span
											class="text-[10px] font-black tracking-widest text-brand-light/20 uppercase sm:hidden"
											>Timestamp</span
										>
										<span class="text-sm font-medium text-brand-light/80"
											>{formatDateTime(result.created_at)}</span
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
												effectiveSlowThreshold
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
