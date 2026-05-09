<script lang="ts">
	import { Activity, Clock, Bell, LineChart, Globe, Zap } from 'lucide-svelte';

	const features = [
		{
			title: 'Real-time Monitoring',
			description: 'We check your services every 60 seconds from multiple locations worldwide.',
			icon: Activity
		},
		{
			title: 'Instant Alerts',
			description: 'Get notified via Email, Slack, or Telegram as soon as something goes wrong.',
			icon: Bell
		},
		{
			title: 'Detailed History',
			description: 'Analyze uptime trends and performance bottlenecks with deep insights.',
			icon: LineChart
		},
		{
			title: 'Global Coverage',
			description: 'Global checking stations ensure your site is reachable everywhere.',
			icon: Globe
		},
		{
			title: 'Speed Analysis',
			description: 'Monitor not just uptime, but also response times and latency.',
			icon: Zap
		},
		{
			title: 'SLA Reporting',
			description: 'Generate professional reports for your clients or stakeholders.',
			icon: Clock
		}
	];
</script>

<!-- Hero Section -->
<section class="relative overflow-hidden py-24 lg:py-32">
	<div class="container relative mx-auto px-4 sm:px-6">
		<div class="flex flex-col items-center text-center">
			<div class="inline-flex items-center rounded-full bg-brand-primary/10 px-3 py-1 text-sm font-medium text-brand-primary">
				<span class="flex h-2 w-2 rounded-full bg-brand-primary mr-2"></span>
				Now in Beta
			</div>
			<h1 class="mt-6 max-w-4xl text-5xl font-extrabold tracking-tight sm:text-7xl">
				Monitor your servers <br />
				<span class="bg-gradient-to-r from-brand-primary to-brand-soft bg-clip-text text-transparent">without the headache</span>
			</h1>
			<p class="mt-6 max-w-2xl text-lg text-brand-light/70">
				Isitdead provides professional-grade uptime monitoring and status pages. Simple to set up, impossible to live without.
			</p>
			<div class="mt-10 flex flex-wrap justify-center gap-4">
				<a
					href="/register"
					class="rounded-full bg-brand-primary px-8 py-4 text-lg font-semibold text-brand-dark shadow-lg transition-all hover:bg-brand-primary/90 hover:shadow-brand-primary/20"
				>
					Start Monitoring Free
				</a>
				<a
					href="#demo"
					class="rounded-full bg-brand-dark px-8 py-4 text-lg font-semibold text-brand-light shadow-md ring-1 ring-brand-light/20 transition-all hover:bg-brand-light/5"
				>
					View Live Demo
				</a>
			</div>
		</div>

		<!-- Mock Dashboard -->
		<div class="mt-20 overflow-hidden rounded-2xl border border-brand-light/10 bg-brand-dark p-2 shadow-2xl">
			<div class="rounded-xl border border-brand-light/5 bg-brand-light/5 p-6">
				<div class="flex items-center justify-between mb-8">
					<h3 class="text-xl font-bold">Monitor Overview</h3>
					<span class="text-sm text-brand-primary font-medium flex items-center gap-1">
						<span class="h-2 w-2 rounded-full bg-brand-primary"></span>
						All Systems Operational
					</span>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
					{#each [1, 2, 3] as i}
						<div class="rounded-lg border border-brand-light/10 bg-brand-dark p-4 shadow-sm">
							<div class="flex items-center justify-between mb-2">
								<span class="font-medium">API Server {i}</span>
								<span class="text-xs text-brand-primary px-2 py-0.5 rounded-full bg-brand-primary/10">Up</span>
							</div>
							<div class="flex gap-1">
								{#each Array(20) as _}
									<div class="h-8 w-full rounded-sm bg-brand-primary opacity-80"></div>
								{/each}
							</div>
							<div class="mt-2 flex items-center justify-between text-[10px] text-brand-light/40">
								<span>99.99% Uptime</span>
								<span>32ms avg</span>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</div>
</section>

<!-- Features Section -->
<section class="bg-brand-light/5 py-24">
	<div class="container mx-auto px-4 sm:px-6">
		<div class="mb-16 text-center">
			<h2 class="text-3xl font-bold tracking-tight sm:text-4xl">Everything you need to sleep better</h2>
			<p class="mt-4 text-brand-light/60">Professional features for modern dev teams.</p>
		</div>
		<div class="grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3">
			{#each features as feature}
				{@const Icon = feature.icon}
				<div class="group rounded-2xl border border-brand-light/10 bg-brand-dark p-8 transition-all hover:shadow-xl hover:border-brand-primary/30">
					<div class="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-xl bg-brand-primary/10 text-brand-primary transition-colors group-hover:bg-brand-primary group-hover:text-brand-dark">
						<Icon class="h-6 w-6" />
					</div>
					<h3 class="mb-2 text-xl font-bold">{feature.title}</h3>
					<p class="text-brand-light/60">{feature.description}</p>
				</div>
			{/each}
		</div>
	</div>
</section>
