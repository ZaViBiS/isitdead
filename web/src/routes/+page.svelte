<script lang="ts">
	import {
		Activity,
		ArrowRight,
		BarChart3,
		CheckCircle2,
		Clock3,
		Eye,
		Gauge,
		ListChecks,
		Server as ServerIcon,
		ShieldCheck,
		Wifi,
		XCircle
	} from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';

	let isLoggedIn = $state(false);

	onMount(() => {
		isLoggedIn = !!localStorage.getItem('token');
	});

	const features = [
		{
			title: 'Website and API monitoring',
			description:
				'Track public pages, API health endpoints, and critical services from one dashboard.',
			icon: Wifi
		},
		{
			title: 'Clear incident history',
			description: 'See failed checks, error responses, and outages without digging through noise.',
			icon: XCircle
		},
		{
			title: 'Flexible check intervals',
			description:
				'Monitor high-priority endpoints often and low-priority services less frequently.',
			icon: Clock3
		},
		{
			title: 'Response time tracking',
			description: 'Watch latency trends so slow endpoints are visible before they become outages.',
			icon: Gauge
		},
		{
			title: 'At-a-glance status',
			description: 'Quickly understand which services are healthy, degraded, or unavailable.',
			icon: Eye
		},
		{
			title: 'Private dashboard',
			description: 'Keep monitors and service history behind authenticated access for your team.',
			icon: ShieldCheck
		}
	];

	const monitors = [
		{
			name: 'Public Website',
			target: 'https://example.com',
			type: 'HTTP',
			status: '200 OK',
			uptime: '99.98%',
			latency: '42ms'
		},
		{
			name: 'API Gateway',
			target: 'https://api.example.com/health',
			type: 'HTTP',
			status: '200 OK',
			uptime: '99.94%',
			latency: '68ms'
		},
		{
			name: 'Database Port',
			target: 'db.internal:5432',
			type: 'TCP',
			status: 'Connected',
			uptime: '99.99%',
			latency: '12ms'
		}
	];

	const incidents = [
		{ status: '500 Internal Server Error', time: '12:42', latency: '450ms' },
		{ status: 'TCP Connection Error', time: '10:15', latency: '5000ms' },
		{ status: '404 Not Found', time: 'Yesterday', latency: '120ms' }
	];

	const workflowItems = [
		{ label: '1', value: 'Add the website, API endpoint, or TCP service you want to monitor.' },
		{
			label: '2',
			value: 'Choose the check type and interval that fits the importance of the service.'
		},
		{
			label: '3',
			value: 'Watch uptime, latency, current status, and recent failures in the dashboard.'
		},
		{ label: '4', value: 'Open the incident history when something fails and see what changed.' }
	];

	const chartBars = [
		44, 48, 54, 42, 60, 66, 72, 58, 64, 70, 78, 52, 46, 74, 80, 62, 57, 67, 73, 84, 76, 68, 59, 63,
		71, 77, 55, 49, 65, 82, 88, 70
	];
</script>

<svelte:head>
	<title>isitdead - uptime monitoring for small teams</title>
	<meta
		name="description"
		content="Uptime monitoring for websites, APIs, and TCP services with status dashboards, latency tracking, and incident history."
	/>
</svelte:head>

<div class="relative isolate overflow-hidden">
	<div class="pointer-events-none absolute inset-0 -z-10">
		<div
			class="absolute top-[-10rem] left-1/2 h-[36rem] w-[36rem] -translate-x-1/2 rounded-full bg-brand-primary/24 blur-[150px]"
		></div>
		<div
			class="absolute top-[28rem] right-[-10rem] h-[30rem] w-[30rem] rounded-full bg-brand-soft/14 blur-[120px]"
		></div>
	</div>

	<section class="px-4 py-14 sm:px-6 sm:py-20 lg:py-28">
		<div
			class="container mx-auto grid max-w-7xl gap-14 lg:grid-cols-[0.95fr_1.05fr] lg:items-center"
		>
			<div class="animate-rise">
				<div class="signal-pill mb-8 gap-3">
					<span class="relative flex h-2.5 w-2.5">
						<span
							class="absolute inline-flex h-full w-full animate-ping rounded-full bg-brand-primary opacity-40"
						></span>
						<span class="relative inline-flex h-2.5 w-2.5 rounded-full bg-brand-primary"></span>
					</span>
					Uptime monitoring service
				</div>

				<h1
					class="display-gradient max-w-5xl text-4xl leading-[0.92] font-black tracking-[-0.07em] sm:text-7xl lg:text-8xl"
				>
					Know what is down before users do.
				</h1>

				<p class="mt-8 max-w-2xl text-lg leading-8 font-medium text-brand-light/62 sm:text-xl">
					isitdead watches your websites, APIs, and critical ports, then turns uptime, latency, and
					failed checks into a dashboard your team can understand quickly.
				</p>

				<div class="mt-10 flex flex-col gap-4 sm:flex-row">
					<a
						href={isLoggedIn ? resolve('/dashboard') : resolve('/register')}
						class="group inline-flex items-center justify-center rounded-2xl bg-brand-primary px-8 py-4 text-base font-black text-brand-dark shadow-2xl shadow-brand-primary/20 transition hover:-translate-y-0.5 hover:bg-brand-primary/90 active:translate-y-0"
					>
						{isLoggedIn ? 'Open dashboard' : 'Create monitor'}
						<ArrowRight class="ml-2 h-5 w-5 transition group-hover:translate-x-1" />
					</a>
					<a
						href={resolve('/#features')}
						class="inline-flex items-center justify-center rounded-2xl border border-brand-light/10 bg-brand-light/[0.03] px-8 py-4 text-base font-black text-brand-light/75 backdrop-blur transition hover:border-brand-primary/30 hover:text-brand-light"
					>
						Explore features
					</a>
				</div>

				<div class="mt-10 grid max-w-xl gap-3 sm:grid-cols-3">
					<div class="soft-panel rounded-3xl p-4">
						<div class="text-2xl font-black text-brand-primary">HTTP</div>
						<div class="micro-label mt-1">checks</div>
					</div>
					<div class="soft-panel rounded-3xl p-4">
						<div class="text-2xl font-black text-brand-primary">TCP</div>
						<div class="micro-label mt-1">ports</div>
					</div>
					<div class="soft-panel rounded-3xl p-4">
						<div class="text-2xl font-black text-brand-primary">30d</div>
						<div class="micro-label mt-1">history</div>
					</div>
				</div>
			</div>

			<div id="demo" class="animate-rise animate-rise-delay relative">
				<div
					class="absolute -inset-4 rounded-[3.25rem] bg-gradient-to-br from-brand-primary/20 via-brand-soft/10 to-transparent blur-2xl"
				></div>
				<div class="glass-panel relative overflow-hidden rounded-[2.75rem] p-3">
					<div class="rounded-[2.25rem] border border-brand-light/10 bg-brand-dark/80">
						<div class="flex items-center justify-between border-b border-brand-light/10 px-5 py-4">
							<div class="flex items-center gap-2">
								<span class="h-3 w-3 rounded-full bg-brand-accent"></span>
								<span class="h-3 w-3 rounded-full bg-[#E5B181]"></span>
								<span class="h-3 w-3 rounded-full bg-brand-primary"></span>
							</div>
							<div
								class="rounded-full bg-brand-primary/10 px-3 py-1 text-[10px] font-black tracking-widest text-brand-primary uppercase"
							>
								Live dashboard
							</div>
						</div>

						<div class="p-5 sm:p-7">
							<div class="grid gap-5 lg:grid-cols-[1fr_13rem]">
								<div class="soft-panel rounded-[2rem] p-5">
									<div class="mb-6 flex items-start justify-between gap-4">
										<div>
											<p class="micro-label">Infrastructure health</p>
											<h2 class="mt-2 text-2xl font-black tracking-tight">
												All systems operational
											</h2>
										</div>
										<div class="rounded-2xl bg-brand-primary/10 p-3 text-brand-primary">
											<Activity class="h-6 w-6" />
										</div>
									</div>

									<div class="flex h-32 items-end gap-1.5">
										{#each chartBars as height, index (index)}
											<div
												class="flex-1 rounded-t-md transition hover:opacity-100"
												class:bg-brand-accent={index === 12}
												class:bg-[#E5B181]={index === 26}
												class:bg-brand-primary={index !== 12 && index !== 26}
												style={`height: ${height}%; opacity: ${index === 12 ? 0.85 : 0.28 + index / 62}`}
											></div>
										{/each}
									</div>
								</div>

								<div class="grid gap-4">
									<div
										class="rounded-[2rem] border border-brand-primary/20 bg-brand-primary/10 p-5"
									>
										<div
											class="flex items-center gap-2 text-xs font-black tracking-widest text-brand-primary uppercase"
										>
											<CheckCircle2 class="h-4 w-4" />
											Uptime
										</div>
										<div class="mt-3 text-4xl font-black text-brand-primary">99.98%</div>
									</div>
									<div class="soft-panel rounded-[2rem] p-5">
										<div
											class="flex items-center gap-2 text-xs font-black tracking-widest text-brand-light/35 uppercase"
										>
											<BarChart3 class="h-4 w-4" />
											Average
										</div>
										<div class="mt-3 text-4xl font-black">
											42<span class="ml-1 text-sm text-brand-light/25">ms</span>
										</div>
									</div>
								</div>
							</div>

							<div class="mt-5 space-y-3">
								{#each monitors as monitor (monitor.name)}
									<div
										class="rounded-3xl border border-brand-light/10 bg-brand-dark/70 p-4 transition hover:border-brand-primary/25"
									>
										<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
											<div class="flex items-center gap-4">
												<div
													class="flex h-12 w-12 items-center justify-center rounded-2xl border border-brand-light/10 bg-brand-light/[0.04] text-brand-primary"
												>
													<ServerIcon class="h-6 w-6" />
												</div>
												<div>
													<div class="flex flex-wrap items-center gap-2">
														<h3 class="font-black">{monitor.name}</h3>
														<span
															class="rounded-lg border border-brand-light/10 bg-brand-light/[0.04] px-2 py-0.5 text-[10px] font-black tracking-widest text-brand-light/35 uppercase"
														>
															{monitor.type}
														</span>
													</div>
													<p
														class="mt-1 max-w-[16rem] truncate text-sm font-medium text-brand-light/35 sm:max-w-xs"
													>
														{monitor.target}
													</p>
												</div>
											</div>
											<div class="grid grid-cols-3 gap-3 text-right sm:min-w-64">
												<div>
													<div class="text-sm font-black text-brand-primary">{monitor.status}</div>
													<div
														class="text-[10px] font-bold tracking-widest text-brand-light/25 uppercase"
													>
														status
													</div>
												</div>
												<div>
													<div class="text-sm font-black">{monitor.uptime}</div>
													<div
														class="text-[10px] font-bold tracking-widest text-brand-light/25 uppercase"
													>
														uptime
													</div>
												</div>
												<div>
													<div class="text-sm font-black text-[#E5B181]">{monitor.latency}</div>
													<div
														class="text-[10px] font-bold tracking-widest text-brand-light/25 uppercase"
													>
														latency
													</div>
												</div>
											</div>
										</div>
									</div>
								{/each}
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</section>

	<section id="features" class="section-rule px-4 py-16 sm:px-6 sm:py-24">
		<div class="container mx-auto max-w-7xl">
			<div class="mb-14 max-w-3xl">
				<p class="signal-kicker">What is included</p>
				<h2 class="mt-4 text-3xl font-black tracking-[-0.05em] sm:text-6xl">
					The monitoring signals teams actually look at.
				</h2>
			</div>

			<div class="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
				{#each features as feature (feature.title)}
					{@const Icon = feature.icon}
					<article
						class="glass-panel group rounded-[2.25rem] p-7 transition hover:-translate-y-1 hover:border-brand-primary/30 hover:bg-brand-panel-soft"
					>
						<div
							class="mb-7 flex h-14 w-14 items-center justify-center rounded-2xl bg-brand-primary/10 text-brand-primary transition group-hover:bg-brand-primary group-hover:text-brand-dark"
						>
							<Icon class="h-7 w-7" />
						</div>
						<h3 class="text-2xl font-black tracking-tight">{feature.title}</h3>
						<p class="mt-3 leading-7 text-brand-light/45">{feature.description}</p>
					</article>
				{/each}
			</div>
		</div>
	</section>

	<section class="px-4 py-16 sm:px-6 sm:py-24">
		<div class="container mx-auto grid max-w-7xl gap-8 lg:grid-cols-[0.8fr_1.2fr] lg:items-start">
			<div>
				<p class="signal-kicker">Incident review</p>
				<h2 class="mt-4 text-3xl font-black tracking-[-0.05em] sm:text-6xl">
					A timeline built for quick triage.
				</h2>
				<p class="mt-6 text-lg leading-8 text-brand-light/50">
					The dashboard keeps recent checks visible and separates failing responses, so you can
					understand what broke and when without reading server logs first.
				</p>
			</div>

			<div class="glass-panel rounded-[2.5rem] p-5">
				<div class="mb-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
					<div class="flex items-center gap-3">
						<div class="rounded-2xl bg-brand-accent/10 p-3 text-brand-accent">
							<XCircle class="h-6 w-6" />
						</div>
						<div>
							<h3 class="text-2xl font-black tracking-tight">Recent incidents</h3>
							<p class="text-sm text-brand-light/35">Failed and degraded checks only</p>
						</div>
					</div>
					<span
						class="rounded-full border border-brand-light/10 bg-brand-light/[0.03] px-4 py-2 text-xs font-black tracking-widest text-brand-light/35 uppercase"
					>
						Last 24 hours
					</span>
				</div>

				<div class="overflow-hidden rounded-[1.75rem] border border-brand-light/10">
					{#each incidents as incident (incident.status)}
						<div
							class="relative grid gap-3 border-b border-brand-light/10 bg-brand-light/[0.015] p-4 pl-8 before:absolute before:top-0 before:bottom-0 before:left-4 before:w-px before:bg-brand-light/10 after:absolute after:top-6 after:left-[0.875rem] after:h-2 after:w-2 after:rounded-full after:bg-brand-accent last:border-b-0 sm:grid-cols-[1fr_7rem_6rem] sm:items-center"
						>
							<div>
								<div class="text-sm font-black text-brand-light">{incident.status}</div>
								<div class="mt-1 text-xs font-bold tracking-widest text-brand-light/25 uppercase">
									check response
								</div>
							</div>
							<div class="text-sm font-bold text-brand-light/60">{incident.time}</div>
							<div class="text-sm font-black text-brand-accent sm:text-right">
								{incident.latency}
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</section>

	<section id="how-it-works" class="px-4 pb-16 sm:px-6 sm:pb-24">
		<div class="container mx-auto max-w-7xl">
			<div
				class="overflow-hidden rounded-[3rem] border border-brand-primary/20 bg-[linear-gradient(135deg,#73e2a7_0%,#b8f0bf_55%,#def4c6_100%)] text-brand-dark shadow-2xl shadow-brand-primary/10"
			>
				<div class="grid gap-0 lg:grid-cols-[0.9fr_1.1fr]">
					<div class="p-8 sm:p-12 lg:p-14">
						<div
							class="mb-6 inline-flex items-center gap-2 rounded-full bg-brand-dark/10 px-4 py-2 text-xs font-black tracking-widest uppercase"
						>
							<ListChecks class="h-4 w-4" />
							How it works
						</div>
						<h2 class="text-3xl leading-none font-black tracking-[-0.06em] sm:text-6xl">
							Start watching a service in minutes.
						</h2>
						<p class="mt-6 max-w-xl text-lg leading-8 font-semibold text-brand-dark/70">
							Add a monitor, choose how often it should be checked, and use the dashboard to track
							current status, uptime, latency, and recent incidents.
						</p>
						<a
							href={isLoggedIn ? resolve('/dashboard') : resolve('/register')}
							class="mt-8 inline-flex items-center rounded-2xl bg-brand-dark px-7 py-4 font-black text-brand-primary transition hover:-translate-y-0.5"
						>
							{isLoggedIn ? 'Open monitors' : 'Create account'}
							<ArrowRight class="ml-2 h-5 w-5" />
						</a>
					</div>

					<div class="bg-brand-dark/95 p-5 sm:p-8 lg:p-10">
						<div class="rounded-[2rem] border border-brand-light/10 bg-brand-light/[0.03] p-5">
							<div class="mb-5 flex items-center gap-3 text-brand-light">
								<CheckCircle2 class="h-5 w-5 text-brand-primary" />
								<span class="font-black">Monitoring flow</span>
							</div>
							<div class="space-y-3">
								{#each workflowItems as item (item.label)}
									<div class="rounded-2xl border border-brand-light/10 bg-brand-dark/80 p-4">
										<div class="text-xs font-black tracking-widest text-brand-primary uppercase">
											Step {item.label}
										</div>
										<div class="mt-1 text-sm leading-6 text-brand-light/65">
											{item.value}
										</div>
									</div>
								{/each}
							</div>
							<p class="mt-5 text-sm leading-6 text-brand-light/40">
								Designed for everyday operations: quick checks, clear state, and enough history to
								understand what happened.
							</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	</section>
</div>

<style>
	@keyframes rise {
		from {
			opacity: 0;
			transform: translateY(1.25rem);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.animate-rise {
		animation: rise 700ms ease-out both;
	}

	.animate-rise-delay {
		animation-delay: 140ms;
	}
</style>
