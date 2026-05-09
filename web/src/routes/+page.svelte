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
			<div class="inline-flex items-center rounded-full bg-indigo-50 px-3 py-1 text-sm font-medium text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-300">
				<span class="flex h-2 w-2 rounded-full bg-indigo-600 mr-2"></span>
				Now in Beta
			</div>
			<h1 class="mt-6 max-w-4xl text-5xl font-extrabold tracking-tight sm:text-7xl">
				Monitor your servers <br />
				<span class="bg-gradient-to-r from-indigo-600 to-violet-600 bg-clip-text text-transparent">without the headache</span>
			</h1>
			<p class="mt-6 max-w-2xl text-lg text-slate-600 dark:text-slate-400">
				Isitdead provides professional-grade uptime monitoring and status pages. Simple to set up, impossible to live without.
			</p>
			<div class="mt-10 flex flex-wrap justify-center gap-4">
				<a
					href="/register"
					class="rounded-full bg-indigo-600 px-8 py-4 text-lg font-semibold text-white shadow-lg transition-all hover:bg-indigo-700 hover:shadow-indigo-200 dark:hover:shadow-none"
				>
					Start Monitoring Free
				</a>
				<a
					href="#demo"
					class="rounded-full bg-white px-8 py-4 text-lg font-semibold text-slate-900 shadow-md ring-1 ring-slate-200 transition-all hover:bg-slate-50 dark:bg-slate-900 dark:text-white dark:ring-slate-800 dark:hover:bg-slate-800"
				>
					View Live Demo
				</a>
			</div>
		</div>

		<!-- Mock Dashboard -->
		<div class="mt-20 overflow-hidden rounded-2xl border border-slate-200 bg-white p-2 shadow-2xl dark:border-slate-800 dark:bg-slate-900">
			<div class="rounded-xl border border-slate-100 bg-slate-50/50 p-6 dark:border-slate-800 dark:bg-slate-950">
				<div class="flex items-center justify-between mb-8">
					<h3 class="text-xl font-bold">Monitor Overview</h3>
					<span class="text-sm text-green-500 font-medium flex items-center gap-1">
						<span class="h-2 w-2 rounded-full bg-green-500"></span>
						All Systems Operational
					</span>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
					{#each [1, 2, 3] as i}
						<div class="rounded-lg border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-800 dark:bg-slate-900">
							<div class="flex items-center justify-between mb-2">
								<span class="font-medium">API Server {i}</span>
								<span class="text-xs text-green-500 px-2 py-0.5 rounded-full bg-green-50 dark:bg-green-900/20">Up</span>
							</div>
							<div class="flex gap-1">
								{#each Array(20) as _}
									<div class="h-8 w-full rounded-sm bg-green-500 opacity-80"></div>
								{/each}
							</div>
							<div class="mt-2 flex items-center justify-between text-[10px] text-slate-400">
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
<section class="bg-slate-100 py-24 dark:bg-slate-900/50">
	<div class="container mx-auto px-4 sm:px-6">
		<div class="mb-16 text-center">
			<h2 class="text-3xl font-bold tracking-tight sm:text-4xl">Everything you need to sleep better</h2>
			<p class="mt-4 text-slate-600 dark:text-slate-400">Professional features for modern dev teams.</p>
		</div>
		<div class="grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3">
			{#each features as feature}
				{@const Icon = feature.icon}
				<div class="group rounded-2xl border border-slate-200 bg-white p-8 transition-all hover:shadow-xl dark:border-slate-800 dark:bg-slate-950">
					<div class="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 transition-colors group-hover:bg-indigo-600 group-hover:text-white dark:bg-indigo-900/20 dark:text-indigo-400">
						<Icon class="h-6 w-6" />
					</div>
					<h3 class="mb-2 text-xl font-bold">{feature.title}</h3>
					<p class="text-slate-600 dark:text-slate-400">{feature.description}</p>
				</div>
			{/each}
		</div>
	</div>
</section>
