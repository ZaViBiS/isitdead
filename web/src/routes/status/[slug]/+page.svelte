<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import {
		Activity,
		AlertCircle,
		ArrowRight,
		BarChart3,
		CheckCircle2,
		Clock,
		ExternalLink,
		Globe2,
		History,
		RefreshCw,
		ShieldCheck,
		WifiOff
	} from 'lucide-svelte';
	import StatusChart from '$lib/StatusChart.svelte';
	import {
		calculateAvgLatency,
		calculateUptime,
		getFaviconUrl,
		getHourlyBuckets,
		getLatestCheck,
		getRecentHistory,
		getStatusColor,
		type CheckResult
	} from '$lib/utils';

	interface PublicMonitor {
		id: number;
		name: string;
		url: string;
		check_type: string;
		check_interval: number;
		public_slug: string;
		created_at: string;
		updated_at: string;
	}

	let monitor = $state<PublicMonitor | null>(null);
	let history30d = $state<CheckResult[]>([]);
	let history24h = $state<CheckResult[]>([]);
	let incidents = $state<CheckResult[]>([]);
	let isLoading = $state(true);
	let error = $state('');

	onMount(() => {
		void loadStatusPage();
	});

	async function loadStatusPage() {
		const slug = page.params.slug ?? '';
		isLoading = true;
		error = '';

		if (!slug) {
			error = 'This public monitor was not found or is no longer published.';
			isLoading = false;
			return;
		}

		try {
			const monitorRes = await fetch(`/api/public/monitors/${encodeURIComponent(slug)}`);
			if (!monitorRes.ok) {
				error = 'This public monitor was not found or is no longer published.';
				return;
			}

			monitor = (await monitorRes.json()) as PublicMonitor;

			const [historyRes, incidentsRes] = await Promise.all([
				fetch(`/api/public/monitors/${encodeURIComponent(slug)}/results?hours=720`),
				fetch(`/api/public/monitors/${encodeURIComponent(slug)}/results?incidents=true&limit=50`)
			]);

			if (historyRes.ok) {
				history30d = (await historyRes.json()) as CheckResult[];
				history24h = getRecentHistory(history30d, 24);
			}

			if (incidentsRes.ok) {
				incidents = (await incidentsRes.json()) as CheckResult[];
			}
		} catch {
			error = 'Connection error. Could not load this public status page.';
		} finally {
			isLoading = false;
		}
	}

	function isHealthyStatus(status: string) {
		return status.startsWith('2') || status === 'Connected';
	}

	function isUnknownStatus(status: string) {
		return !status || status === 'unknown';
	}

	function publicStatusColor(status: string, latency: number) {
		if (isUnknownStatus(status)) return '#1f332f';
		return getStatusColor(status, latency);
	}

	function currentLabel(status: string) {
		if (isUnknownStatus(status)) return 'Waiting for first check';
		return isHealthyStatus(status) ? 'Operational' : 'Incident detected';
	}

	function compactStatus(status: string) {
		if (isUnknownStatus(status)) return 'Unknown';
		if (status === 'Connected') return 'Connected';
		return status.length > 34 ? `${status.slice(0, 31)}...` : status;
	}

	function formatDateTime(value: string) {
		return new Date(value).toLocaleString('en-US', {
			dateStyle: 'medium',
			timeStyle: 'short'
		});
	}

	function formatInterval(seconds: number) {
		if (seconds < 60) return `${seconds}s`;
		const minutes = Math.floor(seconds / 60);
		const remainingSeconds = seconds % 60;
		if (minutes < 60)
			return remainingSeconds > 0 ? `${minutes}m ${remainingSeconds}s` : `${minutes}m`;
		const hours = Math.floor(minutes / 60);
		const remainingMinutes = minutes % 60;
		return remainingMinutes > 0 ? `${hours}h ${remainingMinutes}m` : `${hours}h`;
	}

	function latestCheck() {
		return getLatestCheck(history30d);
	}

	function lastUpdatedLabel() {
		const latest = latestCheck();
		if (latest) return formatDateTime(latest.created_at);
		if (monitor?.updated_at) return formatDateTime(monitor.updated_at);
		return 'Waiting for data';
	}

	function maxLatency(history: CheckResult[]) {
		return Math.max(...history.map((result) => result.latency), latestCheck()?.latency ?? 0, 0);
	}

	function targetHref(url: string, checkType: string) {
		if (checkType === 'ping') return '';
		return url.startsWith('http://') || url.startsWith('https://') ? url : `https://${url}`;
	}
</script>

<svelte:head>
	<title>{monitor ? `${monitor.name} status - isitdead` : 'Public status - isitdead'}</title>
	<meta
		name="description"
		content={monitor
			? `Live public status page for ${monitor.name}. View uptime, latency, recent checks, and incidents.`
			: 'Live public status pages powered by isitdead.'}
	/>
</svelte:head>

<div class="relative isolate min-h-[calc(100vh-4rem)] overflow-hidden">
	<div class="pointer-events-none absolute inset-0 -z-10">
		<div
			class="absolute top-[-10rem] left-1/2 h-[34rem] w-[34rem] -translate-x-1/2 rounded-full bg-brand-primary/15 blur-[140px]"
		></div>
		<div
			class="absolute right-[-14rem] bottom-[12rem] h-[30rem] w-[30rem] rounded-full bg-brand-soft/10 blur-[120px]"
		></div>
		<div
			class="absolute inset-0 bg-[linear-gradient(to_right,rgba(222,244,198,0.035)_1px,transparent_1px),linear-gradient(to_bottom,rgba(222,244,198,0.035)_1px,transparent_1px)] [mask-image:linear-gradient(to_bottom,black,transparent_86%)] bg-[size:72px_72px]"
		></div>
	</div>

	<div class="container mx-auto max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:py-14">
		{#if isLoading}
			<div
				class="flex min-h-[60vh] flex-col items-center justify-center gap-5 rounded-[2rem] border border-brand-light/10 bg-brand-light/[0.025]"
			>
				<RefreshCw class="h-10 w-10 animate-spin text-brand-primary" />
				<div class="text-center">
					<h1 class="text-lg font-black tracking-widest text-brand-light/25 uppercase">
						Loading public status
					</h1>
					<p class="mt-2 text-sm text-brand-light/20">Fetching latest monitor checks...</p>
				</div>
			</div>
		{:else if error || !monitor}
			<section
				class="grid min-h-[60vh] gap-8 rounded-[2.5rem] border border-brand-light/10 bg-[#111f1c]/80 p-6 shadow-2xl shadow-black/20 sm:p-10 lg:grid-cols-[minmax(0,1fr)_22rem] lg:items-center"
			>
				<div>
					<div
						class="mb-6 inline-flex items-center gap-2 rounded-full border border-brand-accent/20 bg-brand-accent/10 px-3 py-1.5 text-xs font-black tracking-widest text-brand-accent uppercase"
					>
						<AlertCircle class="h-4 w-4" />
						Status page unavailable
					</div>
					<h1
						class="max-w-3xl text-4xl leading-[0.92] font-black tracking-[-0.05em] text-brand-light sm:text-6xl"
					>
						This monitor is not public.
					</h1>
					<p class="mt-6 max-w-2xl text-base leading-7 font-medium text-brand-light/50 sm:text-lg">
						{error || 'The route does not exist, moved, or was never published.'}
					</p>
					<div class="mt-8 flex flex-col gap-3 sm:flex-row">
						<a
							href={resolve('/')}
							class="inline-flex items-center justify-center rounded-2xl bg-brand-primary px-6 py-3 font-black text-brand-dark shadow-2xl shadow-brand-primary/15 transition hover:-translate-y-0.5 hover:bg-brand-primary/90 active:translate-y-0"
						>
							Go home
						</a>
						<a
							href={resolve('/register')}
							class="inline-flex items-center justify-center rounded-2xl border border-brand-light/10 bg-brand-light/[0.03] px-6 py-3 font-black text-brand-light/75 transition hover:border-brand-primary/30 hover:text-brand-light"
						>
							Create monitor
						</a>
					</div>
				</div>
				<aside class="rounded-[2rem] border border-brand-light/10 bg-brand-dark/50 p-6">
					<div class="mb-4 rounded-2xl bg-brand-accent/10 p-3 text-brand-accent">
						<WifiOff class="h-8 w-8" />
					</div>
					<div class="text-xs font-black tracking-widest text-brand-light/35 uppercase">
						Requested slug
					</div>
					<div class="mt-2 text-2xl font-black break-words text-brand-light/80">
						{page.params.slug}
					</div>
				</aside>
			</section>
		{:else}
			{@const uptime = calculateUptime(history30d)}
			{@const avgLatency = calculateAvgLatency(history30d)}
			{@const current = latestCheck()}
			{@const currentStatus = current?.status ?? 'unknown'}
			{@const currentLatency = current?.latency ?? 0}
			{@const healthy = isHealthyStatus(currentStatus)}
			{@const unknown = isUnknownStatus(currentStatus)}
			{@const statusLabel = currentLabel(currentStatus)}
			{@const color = publicStatusColor(currentStatus, currentLatency)}
			{@const href = targetHref(monitor.url, monitor.check_type)}

			<section class="mb-8 grid gap-8 lg:grid-cols-[minmax(0,1fr)_25rem] lg:items-start">
				<div class="animate-rise">
					<div
						class="mb-7 inline-flex items-center gap-3 rounded-full border px-4 py-2 text-xs font-black tracking-[0.24em] uppercase shadow-2xl {healthy
							? 'border-brand-primary/20 bg-brand-primary/10 text-brand-primary shadow-brand-primary/10'
							: unknown
								? 'border-brand-light/10 bg-brand-light/[0.04] text-brand-light/45'
								: 'border-brand-accent/20 bg-brand-accent/10 text-brand-accent shadow-brand-accent/10'}"
					>
						<span class="relative flex h-2.5 w-2.5">
							{#if healthy}
								<span
									class="absolute inline-flex h-full w-full animate-ping rounded-full bg-brand-primary opacity-40"
								></span>
							{/if}
							<span
								class="relative inline-flex h-2.5 w-2.5 rounded-full"
								style="background-color: {color}"
							></span>
						</span>
						{statusLabel}
					</div>

					<div class="flex items-start gap-5">
						<div class="relative hidden shrink-0 sm:block">
							<div
								class="flex h-20 w-20 items-center justify-center overflow-hidden rounded-[2rem] border border-brand-light/10 bg-brand-light/[0.04] text-brand-primary/70 shadow-2xl"
							>
								{#if monitor.check_type === 'http' || monitor.check_type === 'links'}
									<Globe2 class="h-10 w-10" />
									<img
										src={getFaviconUrl(monitor.url)}
										alt=""
										class="absolute h-10 w-10 rounded-lg bg-[#111f1c] object-contain"
										onerror={(event) => {
											const target = event.currentTarget as HTMLImageElement;
											target.style.display = 'none';
										}}
									/>
								{:else}
									<Activity class="h-10 w-10" />
								{/if}
							</div>
							<div
								class="absolute -right-1 -bottom-1 h-7 w-7 rounded-full border-4 border-brand-dark"
								style="background-color: {color}"
							></div>
						</div>
						<div class="min-w-0">
							<h1
								class="max-w-5xl text-4xl leading-[0.92] font-black tracking-[-0.07em] break-words text-brand-light sm:text-7xl lg:text-8xl"
							>
								{monitor.name} status
							</h1>
							<p class="mt-6 max-w-3xl text-lg leading-8 font-medium text-brand-light/55">
								Live public monitoring for {monitor.name}. Track current availability, response
								time, uptime, and recent incidents without signing in.
							</p>
						</div>
					</div>

					<div class="mt-8 flex flex-col gap-3 sm:flex-row">
						{#if href}
							<a
								{href}
								target="_blank"
								rel="external noreferrer"
								class="group inline-flex items-center justify-center rounded-2xl bg-brand-primary px-7 py-4 font-black text-brand-dark shadow-2xl shadow-brand-primary/20 transition hover:-translate-y-0.5 hover:bg-brand-primary/90 active:translate-y-0"
							>
								Open monitored target
								<ExternalLink class="ml-2 h-4 w-4 transition group-hover:translate-x-0.5" />
							</a>
						{/if}
						<a
							href={resolve('/register')}
							class="inline-flex items-center justify-center rounded-2xl border border-brand-light/10 bg-brand-light/[0.03] px-7 py-4 font-black text-brand-light/75 backdrop-blur transition hover:border-brand-primary/30 hover:text-brand-light"
						>
							Monitor your own service
							<ArrowRight class="ml-2 h-4 w-4" />
						</a>
					</div>
				</div>

				<aside
					class="animate-rise animate-rise-delay rounded-[2.25rem] border border-brand-light/10 bg-[#111f1c]/90 p-5 shadow-2xl shadow-black/20 backdrop-blur-xl"
				>
					<div
						class="rounded-[1.75rem] border p-5 {healthy
							? 'border-brand-primary/20 bg-brand-primary/10'
							: unknown
								? 'border-brand-light/10 bg-brand-light/[0.03]'
								: 'border-brand-accent/20 bg-brand-accent/10'}"
					>
						<div class="flex items-center justify-between gap-4">
							<div>
								<div class="text-xs font-black tracking-widest text-brand-light/45 uppercase">
									Current status
								</div>
								<div
									class="mt-2 text-3xl leading-none font-black"
									style="color: {healthy
										? '#73E2A7'
										: unknown
											? 'rgba(222,244,198,.72)'
											: '#D62246'}"
								>
									{statusLabel}
								</div>
							</div>
							<div class="rounded-2xl bg-brand-dark/45 p-3" style="color: {color}">
								{#if healthy}
									<CheckCircle2 class="h-8 w-8" />
								{:else if unknown}
									<Activity class="h-8 w-8" />
								{:else}
									<AlertCircle class="h-8 w-8" />
								{/if}
							</div>
						</div>

						<div class="mt-5 truncate text-sm font-medium text-brand-light/45" title={monitor.url}>
							{monitor.url}
						</div>
					</div>

					<div class="mt-4 grid gap-3 sm:grid-cols-2">
						<div class="rounded-3xl border border-brand-light/10 bg-brand-light/[0.035] p-4">
							<div
								class="mb-2 flex items-center gap-2 text-xs font-bold text-brand-light/35 uppercase"
							>
								<Clock class="h-3.5 w-3.5 text-brand-primary" />
								Updated
							</div>
							<div class="text-sm font-black text-brand-light/80">{lastUpdatedLabel()}</div>
						</div>
						<div class="rounded-3xl border border-brand-light/10 bg-brand-light/[0.035] p-4">
							<div
								class="mb-2 flex items-center gap-2 text-xs font-bold text-brand-light/35 uppercase"
							>
								<History class="h-3.5 w-3.5 text-brand-primary" />
								Incidents
							</div>
							<div class="text-2xl font-black text-brand-light/85">{incidents.length}</div>
						</div>
					</div>

					<div class="mt-4 rounded-3xl border border-brand-light/10 bg-brand-light/[0.025] p-4">
						<div class="mb-3 text-xs font-black tracking-widest text-brand-light/35 uppercase">
							Recent checks
						</div>
						<div class="grid gap-3">
							{#if history30d.length > 0}
								{#each history30d.slice(-4).reverse() as result (result.id)}
									<div class="grid grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-3">
										<span
											class="h-2.5 w-2.5 rounded-full"
											style="background-color: {publicStatusColor(result.status, result.latency)}"
										></span>
										<span class="truncate text-sm font-bold text-brand-light/75">
											{compactStatus(result.status)}
										</span>
										<span class="font-mono text-xs font-black text-brand-light/35">
											{result.latency}ms
										</span>
									</div>
								{/each}
							{:else}
								<div class="text-sm font-medium text-brand-light/30">Waiting for first check.</div>
							{/if}
						</div>
					</div>
				</aside>
			</section>

			<section class="mb-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
				<div class="rounded-3xl border border-brand-light/10 bg-brand-light/[0.035] p-5">
					<div class="mb-4 flex items-center justify-between gap-3 text-brand-light/35">
						<span class="text-xs font-bold uppercase">30d uptime</span>
						<ShieldCheck class="h-4 w-4 text-brand-primary" />
					</div>
					<div
						class="text-3xl font-black {uptime >= 99
							? 'text-brand-primary'
							: uptime >= 95
								? 'text-brand-soft'
								: 'text-brand-accent'}"
					>
						{history30d.length > 0 ? `${uptime.toFixed(2)}%` : 'No data'}
					</div>
				</div>
				<div class="rounded-3xl border border-brand-light/10 bg-brand-light/[0.035] p-5">
					<div class="mb-4 flex items-center justify-between gap-3 text-brand-light/35">
						<span class="text-xs font-bold uppercase">Avg response</span>
						<BarChart3 class="h-4 w-4 text-brand-primary" />
					</div>
					<div class="text-3xl font-black text-brand-light/85">
						{avgLatency}<span class="ml-1 text-sm text-brand-light/30">ms</span>
					</div>
				</div>
				<div class="rounded-3xl border border-brand-light/10 bg-brand-light/[0.035] p-5">
					<div class="mb-4 flex items-center justify-between gap-3 text-brand-light/35">
						<span class="text-xs font-bold uppercase">Last response</span>
						<Activity class="h-4 w-4 text-brand-primary" />
					</div>
					<div class="text-3xl font-black text-brand-light/85">
						{current ? currentLatency : 'No data'}{#if current}<span
								class="ml-1 text-sm text-brand-light/30">ms</span
							>{/if}
					</div>
				</div>
				<div class="rounded-3xl border border-brand-light/10 bg-brand-light/[0.035] p-5">
					<div class="mb-4 flex items-center justify-between gap-3 text-brand-light/35">
						<span class="text-xs font-bold uppercase">Interval</span>
						<Clock class="h-4 w-4 text-brand-primary" />
					</div>
					<div class="text-3xl font-black text-brand-light/85">
						{formatInterval(monitor.check_interval)}
					</div>
				</div>
			</section>

			<section class="grid gap-6 lg:grid-cols-[minmax(0,1.45fr)_minmax(20rem,0.55fr)]">
				<div
					class="rounded-[2.5rem] border border-brand-light/10 bg-gradient-to-b from-brand-light/[0.03] to-transparent p-1 shadow-2xl shadow-black/20"
				>
					<div class="rounded-[2.4rem] bg-brand-dark p-5 sm:p-8 lg:p-10">
						<div class="mb-8 flex flex-col justify-between gap-5 sm:flex-row sm:items-center">
							<div>
								<h2 class="mb-1 flex items-center gap-2 text-xl font-black">
									<Activity class="h-5 w-5 text-brand-primary" />
									Response time
								</h2>
								<p class="text-sm text-brand-light/35">
									Latency and availability over the last 24 hours.
								</p>
							</div>
							<div
								class="flex flex-wrap items-center gap-4 rounded-2xl border border-brand-light/10 bg-brand-light/5 px-4 py-2 text-[10px] font-black tracking-widest uppercase"
							>
								<span class="flex items-center gap-1.5">
									<span
										class="h-2 w-2 rounded-full bg-brand-primary shadow-[0_0_8px_rgba(115,226,167,0.5)]"
									></span>
									Healthy
								</span>
								<span class="flex items-center gap-1.5">
									<span class="h-2 w-2 rounded-full bg-[#E5B181]"></span>
									Slow
								</span>
								<span class="flex items-center gap-1.5">
									<span class="h-2 w-2 rounded-full bg-brand-accent"></span>
									Down
								</span>
							</div>
						</div>

						<div class="relative">
							<StatusChart history={history24h} height={430} />
							<div
								class="pointer-events-none absolute right-5 bottom-5 rounded-full border border-brand-light/10 bg-brand-dark/70 px-3 py-1 text-[10px] font-black tracking-widest text-brand-light/30 uppercase backdrop-blur"
							>
								Peak {maxLatency(history24h)}ms
							</div>
						</div>
					</div>
				</div>

				<div class="grid gap-6">
					<section class="rounded-[2rem] border border-brand-light/10 bg-[#111f1c]/90 p-5 sm:p-6">
						<div class="mb-5 flex items-center justify-between gap-4">
							<div>
								<div class="text-xs font-black tracking-widest text-brand-light/35 uppercase">
									Availability
								</div>
								<div class="mt-1 text-2xl font-black text-brand-primary sm:text-3xl">
									{history30d.length > 0 ? `${uptime.toFixed(2)}%` : 'No data'}
								</div>
							</div>
							<div class="rounded-2xl bg-brand-primary/10 p-3 text-brand-primary">
								<ShieldCheck class="h-7 w-7" />
							</div>
						</div>
						<div class="flex h-12 w-full items-end gap-1">
							{#each getHourlyBuckets(history24h) as bucketColor, index (index)}
								<div
									class="group relative flex-1 cursor-help rounded-sm opacity-80 transition hover:opacity-100"
									style="background-color: {bucketColor}; height: {bucketColor === '#1f332f'
										? '38%'
										: '100%'}"
								>
									<div
										class="pointer-events-none absolute bottom-full left-1/2 z-50 mb-3 hidden -translate-x-1/2 group-hover:block"
									>
										<div
											class="rounded-xl border border-brand-light/10 bg-brand-dark/95 px-3 py-2 text-[11px] whitespace-nowrap shadow-2xl ring-1 ring-white/5 backdrop-blur-xl"
										>
											<div class="mb-1 font-black text-brand-light/40">{23 - index}h ago</div>
											{#if bucketColor !== '#1f332f'}
												<div class="flex items-center gap-2">
													<div
														class="h-2 w-2 rounded-full"
														style="background-color: {bucketColor}"
													></div>
													<span class="font-bold">
														{bucketColor === '#73E2A7'
															? 'Healthy'
															: bucketColor === '#E5B181'
																? 'Slow response'
																: 'Down'}
													</span>
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
					</section>

					<section class="rounded-[2rem] border border-brand-light/10 bg-[#111f1c]/90 p-5 sm:p-6">
						<div class="mb-5 flex items-center justify-between">
							<h2 class="flex items-center gap-2 text-xl font-black">
								<History class="h-5 w-5 text-brand-primary" />
								Incidents
							</h2>
							<span
								class="rounded-full bg-brand-light/5 px-3 py-1 text-[10px] font-black tracking-widest text-brand-light/40 uppercase"
							>
								Last 50
							</span>
						</div>
						<div class="max-h-[22rem] divide-y divide-brand-light/5 overflow-y-auto pr-1">
							{#if incidents.length > 0}
								{#each incidents as incident (incident.id)}
									<div class="grid grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-3 py-3">
										<div
											class="h-3 w-3 rounded-full"
											style="background-color: {publicStatusColor(
												incident.status,
												incident.latency
											)}"
										></div>
										<div class="min-w-0">
											<div class="truncate text-sm font-bold text-brand-light/75">
												{compactStatus(incident.status)}
											</div>
											<div class="mt-1 text-xs text-brand-light/30">
												{formatDateTime(incident.created_at)}
											</div>
										</div>
										<div class="font-mono text-sm font-black text-brand-accent">
											{incident.latency}<span class="text-[10px] opacity-50">ms</span>
										</div>
									</div>
								{/each}
							{:else}
								<div class="flex flex-col items-center justify-center gap-4 py-14 text-center">
									<div
										class="rounded-3xl border border-brand-primary/15 bg-brand-primary/10 p-4 text-brand-primary"
									>
										<CheckCircle2 class="h-9 w-9" />
									</div>
									<div>
										<div class="font-black text-brand-light/80">No recent incidents</div>
										<p class="mt-1 text-sm text-brand-light/35">
											Failed checks will appear here when they happen.
										</p>
									</div>
								</div>
							{/if}
						</div>
					</section>
				</div>
			</section>

			<section
				class="mt-6 grid gap-5 rounded-[2rem] border border-brand-primary/20 bg-gradient-to-br from-brand-primary/12 via-brand-light/[0.035] to-brand-soft/10 p-6 shadow-2xl shadow-brand-primary/5 sm:p-8 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center"
			>
				<div>
					<h2 class="text-2xl font-black tracking-tight sm:text-3xl">
						Need a public status page for your own service?
					</h2>
					<p class="mt-3 max-w-2xl text-sm leading-6 text-brand-light/50 sm:text-base">
						isitdead turns uptime checks, response time, and incidents into a monitor that is easy
						to share with your users.
					</p>
				</div>
				<a
					href={resolve('/register')}
					class="inline-flex items-center justify-center rounded-2xl bg-brand-primary px-7 py-4 font-black text-brand-dark shadow-2xl shadow-brand-primary/20 transition hover:-translate-y-0.5 hover:bg-brand-primary/90 active:translate-y-0"
				>
					Start monitoring
					<ArrowRight class="ml-2 h-4 w-4" />
				</a>
			</section>
		{/if}
	</div>
</div>

<style>
	@keyframes rise {
		from {
			opacity: 0;
			transform: translateY(18px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.animate-rise {
		animation: rise 650ms cubic-bezier(0.22, 1, 0.36, 1) both;
	}

	.animate-rise-delay {
		animation-delay: 120ms;
	}
</style>
