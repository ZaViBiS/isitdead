<script lang="ts">
	import { Activity, Clock, Bell, LineChart, Globe, Zap, ShieldCheck, ArrowRight, ExternalLink, History, BarChart3, History as HistoryIcon } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { getStatusColor } from '$lib/utils';

	let isLoggedIn = $state(false);

	onMount(() => {
		isLoggedIn = !!localStorage.getItem('token');
	});

	const features = [
		{
			title: 'Real-time Monitoring',
			description: 'We check your services every 30 seconds from multiple locations worldwide.',
			icon: Activity
		},
		{
			title: 'Instant Alerts',
			description: 'Get notified via Email, Slack, or Telegram as soon as something goes wrong.',
			icon: Bell
		},
		{
			title: 'Incidents Log',
			description: 'Focus on what matters. We filter out the noise and show only problematic checks.',
			icon: History
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
			title: 'Interval Presets',
			description: 'Flexible monitoring intervals from 30 seconds up to 24 hours.',
			icon: Clock
		}
	];
</script>

<div class="relative min-h-screen">
	<!-- Background Glow -->
	<div class="absolute top-0 left-1/2 -translate-x-1/2 w-full h-[800px] bg-gradient-to-b from-brand-primary/10 to-transparent blur-[120px] pointer-events-none -z-10"></div>

	<!-- Hero Section -->
	<section class="relative pt-24 pb-32">
		<div class="container mx-auto px-4 text-center">
			<div class="inline-flex items-center gap-2 rounded-full bg-brand-primary/10 border border-brand-primary/20 px-4 py-1.5 text-xs font-black uppercase tracking-widest text-brand-primary mb-8 animate-in fade-in slide-in-from-top-4">
				<span class="flex h-2 w-2 rounded-full bg-brand-primary shadow-[0_0_8px_rgba(115,226,167,0.5)]"></span>
				Infrastructure Intelligence
			</div>
			
			<h1 class="text-6xl md:text-8xl font-black tracking-tighter mb-8 leading-[0.9]">
				Monitor everything<br />
				<span class="text-brand-primary/80">without the headache</span>
			</h1>
			
			<p class="max-w-2xl mx-auto text-lg md:text-xl text-brand-light/40 font-medium mb-12">
				Professional-grade uptime monitoring and incidents tracking. 
				Simple to set up, impossible to live without.
			</p>

			<div class="flex flex-wrap justify-center gap-4">
				<a
					href={isLoggedIn ? '/dashboard' : '/register'}
					class="group relative rounded-2xl bg-brand-primary px-10 py-5 text-lg font-black text-brand-dark transition-all hover:scale-105 active:scale-95 shadow-2xl shadow-brand-primary/20"
				>
					<span class="flex items-center gap-2">
						{isLoggedIn ? 'Go to Dashboard' : 'Start Monitoring Free'}
						<ArrowRight class="h-5 w-5 group-hover:translate-x-1 transition-transform" />
					</span>
				</a>
				<a
					href="#demo"
					class="rounded-2xl border border-brand-light/10 bg-brand-dark/50 px-10 py-5 text-lg font-black text-brand-light/60 transition-all hover:bg-brand-light/5 hover:border-brand-light/20 backdrop-blur-md"
				>
					View Live Demo
				</a>
			</div>
		</div>

		<!-- Dashboard Preview -->
		<div class="container mx-auto px-4 mt-24">
			<div class="relative group">
				<div class="absolute -inset-1 bg-gradient-to-r from-brand-primary/20 to-brand-soft/20 rounded-[3rem] blur opacity-25 group-hover:opacity-40 transition duration-1000 group-hover:duration-200"></div>
				<div class="relative rounded-[2.8rem] border border-brand-light/10 bg-brand-dark/80 p-4 shadow-2xl backdrop-blur-xl">
					<div class="rounded-[2rem] border border-brand-light/5 bg-brand-light/[0.02] p-8">
						<div class="flex items-center justify-between mb-12">
							<div>
								<h3 class="text-2xl font-black tracking-tight mb-1">Infrastructure Overview</h3>
								<p class="text-xs font-bold text-brand-light/20 uppercase tracking-widest">Global Health Status</p>
							</div>
							<div class="flex items-center gap-2 px-4 py-2 bg-brand-primary/10 rounded-2xl border border-brand-primary/20">
								<span class="h-2 w-2 rounded-full bg-brand-primary shadow-[0_0_8px_rgba(115,226,167,0.5)]"></span>
								<span class="text-[10px] font-black uppercase tracking-widest text-brand-primary">All Systems Operational</span>
							</div>
						</div>

						<div class="grid gap-6">
							{#each ['Production API', 'Database Node', 'Auth Service'] as name, i}
								<div class="rounded-3xl border border-brand-light/10 bg-brand-dark/40 p-6 flex flex-col md:flex-row md:items-center justify-between gap-6 transition-all hover:border-brand-primary/30">
									<div class="flex items-center gap-4">
										<div class="h-12 w-12 rounded-2xl bg-brand-light/5 flex items-center justify-center border border-brand-light/10">
											<Globe class="h-6 w-6 text-brand-primary/60" />
										</div>
										<div>
											<div class="font-bold text-lg">{name}</div>
											<div class="text-xs text-brand-light/20 font-medium">s{i+1}.cluster.internal</div>
										</div>
									</div>
									<div class="flex gap-1.5 h-10 items-end flex-1 max-w-md">
										{#each Array(40) as _, j}
											<div 
												class="h-full w-full rounded-sm transition-all hover:scale-y-125" 
												style="background-color: {j === 25 ? '#D62246' : j > 30 ? '#E5B181' : '#73E2A7'}; opacity: {0.3 + (j/60)}"
											></div>
										{/each}
									</div>
									<div class="text-right">
										<div class="text-lg font-black text-brand-primary">99.98%</div>
										<div class="text-[10px] font-bold text-brand-light/20 uppercase tracking-widest">Uptime</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</div>
		</div>
	</section>

	<!-- Features Section -->
	<section class="py-32 bg-brand-light/[0.02] border-y border-brand-light/5">
		<div class="container mx-auto px-4">
			<div class="mb-20 text-center">
				<h2 class="text-4xl md:text-5xl font-black tracking-tight mb-4">Everything you need<br />to sleep better</h2>
				<p class="text-brand-light/40 font-medium">Professional monitoring features for modern dev teams.</p>
			</div>
			
			<div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
				{#each features as feature}
					{@const Icon = feature.icon}
					<div class="group rounded-[2.5rem] border border-brand-light/10 bg-brand-dark p-10 transition-all hover:shadow-2xl hover:border-brand-primary/30 hover:shadow-brand-primary/5">
						<div class="mb-6 inline-flex h-14 w-14 items-center justify-center rounded-2xl bg-brand-primary/10 text-brand-primary transition-all group-hover:scale-110 group-hover:bg-brand-primary group-hover:text-brand-dark">
							<Icon class="h-7 w-7" />
						</div>
						<h3 class="mb-3 text-2xl font-black tracking-tight">{feature.title}</h3>
						<p class="text-brand-light/40 leading-relaxed font-medium">{feature.description}</p>
					</div>
				{/each}
			</div>
		</div>
	</section>

	<!-- Demo Section -->
	<section id="demo" class="py-32 relative">
		<div class="absolute bottom-0 left-0 w-full h-[500px] bg-gradient-to-t from-brand-primary/5 to-transparent blur-[120px] pointer-events-none -z-10"></div>
		
		<div class="container mx-auto px-4">
			<div class="mb-20 text-center">
				<h2 class="text-4xl md:text-5xl font-black tracking-tight mb-4">Deep Insight</h2>
				<p class="text-brand-light/40 font-medium">Analyze every incident with precision.</p>
			</div>

			<div class="max-w-5xl mx-auto rounded-[3rem] border border-brand-light/10 bg-brand-dark/50 p-8 lg:p-12 shadow-2xl backdrop-blur-xl">
				<div class="flex flex-col lg:flex-row lg:items-center justify-between gap-8 mb-16">
					<div class="flex items-start gap-6">
						<div class="h-20 w-20 rounded-[2rem] bg-brand-light/5 flex items-center justify-center border border-brand-light/10">
							<Globe class="h-10 w-10 text-brand-primary/60" />
						</div>
						<div>
							<div class="flex items-center gap-3 mb-2">
								<h3 class="text-4xl font-black tracking-tight">Main API Cluster</h3>
								<span class="text-[10px] px-2.5 py-1 rounded-lg bg-brand-light/5 border border-brand-light/10 text-brand-light/40 uppercase font-black tracking-widest">HTTP</span>
							</div>
							<p class="text-lg font-medium text-brand-light/30">api.isitdead.io</p>
						</div>
					</div>
					<div class="flex gap-8 bg-brand-light/[0.02] border border-brand-light/5 rounded-[2rem] p-6">
						<div class="text-right">
							<div class="flex items-center justify-end gap-1.5 text-brand-light/40 text-[10px] font-bold uppercase tracking-widest mb-1">
								<Clock class="h-3 w-3" /> 30d Uptime
							</div>
							<div class="text-3xl font-black text-brand-primary">99.98%</div>
						</div>
						<div class="w-px h-10 bg-brand-light/10 self-center"></div>
						<div class="text-right">
							<div class="flex items-center justify-end gap-1.5 text-brand-light/40 text-[10px] font-bold uppercase tracking-widest mb-1">
								<BarChart3 class="h-3 w-3" /> Avg Latency
							</div>
							<div class="text-3xl font-black text-brand-light/80">42<span class="text-xs font-bold text-brand-light/20 ml-0.5">ms</span></div>
						</div>
					</div>
				</div>

				<!-- Incidents Log Demo -->
				<div class="rounded-[2.5rem] border border-brand-light/10 bg-brand-dark/80 p-8 lg:p-10">
					<div class="mb-8 flex items-center justify-between">
						<h3 class="text-xl font-bold flex items-center gap-2">
							<HistoryIcon class="h-5 w-5 text-brand-primary" />
							Incidents Log
						</h3>
						<span class="text-[10px] font-black uppercase tracking-widest px-3 py-1 bg-brand-light/5 rounded-full text-brand-light/40">
							Recent Issues
						</span>
					</div>

					<div class="overflow-hidden rounded-3xl border border-brand-light/5 bg-brand-light/[0.01]">
						<div class="grid grid-cols-12 px-6 py-4 border-b border-brand-light/10 text-[10px] font-black uppercase tracking-widest text-brand-light/20 bg-brand-light/[0.02]">
							<div class="col-span-1">Status</div>
							<div class="col-span-6">Timestamp</div>
							<div class="col-span-3">Check Response</div>
							<div class="col-span-2 text-right">Latency</div>
						</div>
						<div class="divide-y divide-brand-light/5">
							{#each [
								{ status: '500 Internal Server Error', time: '12:42 PM', latency: 450 },
								{ status: 'Connection Timeout', time: '10:15 AM', latency: 5000 },
								{ status: '404 Not Found', time: 'Yesterday, 8:20 PM', latency: 120 }
							] as incident}
								<div class="grid grid-cols-12 px-6 py-4 items-center">
									<div class="col-span-1">
										<div class="h-3 w-3 rounded-full bg-brand-accent shadow-sm"></div>
									</div>
									<div class="col-span-6">
										<span class="text-sm font-medium text-brand-light/80">{incident.time}</span>
									</div>
									<div class="col-span-3">
										<span class="text-[10px] font-black uppercase px-2 py-0.5 rounded-md bg-brand-light/5 text-brand-light/40 border border-brand-light/10">
											{incident.status}
										</span>
									</div>
									<div class="col-span-2 text-right">
										<span class="text-sm font-black font-mono tracking-tight text-brand-accent">
											{incident.latency}<span class="text-[10px] ml-0.5 opacity-40">ms</span>
										</span>
									</div>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</div>
		</div>
	</section>

	<!-- CTA Section -->
	<section class="py-32">
		<div class="container mx-auto px-4">
			<div class="rounded-[4rem] bg-brand-primary p-12 lg:p-24 text-center relative overflow-hidden group">
				<div class="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent pointer-events-none"></div>
				<h2 class="text-5xl md:text-7xl font-black text-brand-dark mb-8 tracking-tighter relative z-10">Ready to monitor?<br />It's free for beta users.</h2>
				<a 
					href="/register" 
					class="inline-flex items-center gap-2 rounded-2xl bg-brand-dark px-12 py-6 text-xl font-black text-brand-primary transition-all hover:scale-105 active:scale-95 shadow-2xl relative z-10"
				>
					Create Account
					<ArrowRight class="h-6 w-6" />
				</a>
			</div>
		</div>
	</section>

	<!-- Footer -->
	<footer class="py-12 border-t border-brand-light/5 text-center">
		<div class="flex items-center justify-center gap-2 font-black text-brand-light/20 uppercase tracking-[0.3em] text-xs">
			<Activity class="h-4 w-4" /> Isitdead &copy; 2026
		</div>
	</footer>
</div>

<style>
	@keyframes fade-in { from { opacity: 0; } to { opacity: 1; } }
	@keyframes slide-in-from-top-4 { from { transform: translateY(-1rem); } to { transform: translateY(0); } }

	.animate-in {
		animation-duration: 600ms;
		animation-fill-mode: both;
	}
	.fade-in { animation-name: fade-in; }
	.slide-in-from-top-4 { animation-name: slide-in-from-top-4; }
</style>
