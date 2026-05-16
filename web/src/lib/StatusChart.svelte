<script lang="ts">
	import { getStatusColor, formatDate, type CheckResult } from '$lib/utils';

	let {
		history,
		height = 100,
		slowThreshold = 300
	} = $props<{ history: CheckResult[]; height?: number; slowThreshold?: number }>();

	let hoveredResult = $state<CheckResult | null>(null);

	function getChartPoints(history: CheckResult[], width: number, h: number) {
		if (!history || history.length < 2) return [];
		const maxLatency = Math.max(...history.map((r) => r.latency), 100);
		const chartMax = maxLatency * 1.1;
		const startTime = new Date(history[0].created_at).getTime();
		const endTime = new Date(history[history.length - 1].created_at).getTime();
		const duration = Math.max(endTime - startTime, 1);
		return history.map((r) => ({
			x: ((new Date(r.created_at).getTime() - startTime) / duration) * width,
			y: h - (r.latency / chartMax) * h,
			result: r
		}));
	}

	function getPathD(points: { x: number; y: number }[]) {
		if (points.length < 2) return '';
		return points.map((p, i) => (i === 0 ? `M ${p.x},${p.y}` : `L ${p.x},${p.y}`)).join(' ');
	}

	function handleMouseMove(e: MouseEvent) {
		if (!history || history.length === 0) return;
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const x = e.clientX - rect.left;
		const ratio = x / rect.width;
		const startTime = new Date(history[0].created_at).getTime();
		const endTime = new Date(history[history.length - 1].created_at).getTime();
		const targetTime = startTime + ratio * (endTime - startTime);

		hoveredResult = history.reduce((closest: CheckResult, result: CheckResult) => {
			const closestDistance = Math.abs(new Date(closest.created_at).getTime() - targetTime);
			const nextDistance = Math.abs(new Date(result.created_at).getTime() - targetTime);
			return nextDistance < closestDistance ? result : closest;
		});
	}

	const gridLines = [0.25, 0.5, 0.75];
</script>

<div
	class="status-chart relative w-full cursor-crosshair overflow-hidden rounded-[2rem] border border-brand-light/5 bg-brand-light/[0.01]"
	style="--chart-height: {height}px"
	onmousemove={handleMouseMove}
	onmouseleave={() => (hoveredResult = null)}
	role="presentation"
>
	{#if history && history.length >= 2}
		{@const points = getChartPoints(history, 1000, 300)}
		<svg class="h-full w-full" preserveAspectRatio="none" viewBox="0 0 1000 300">
			<defs>
				<linearGradient id="grad-large" x1="0%" y1="0%" x2="0%" y2="100%">
					<stop offset="0%" stop-color="#73E2A7" stop-opacity="0.15" />
					<stop offset="100%" stop-color="#73E2A7" stop-opacity="0" />
				</linearGradient>
				<linearGradient id="line-grad-large" x1="0%" y1="0%" x2="100%" y2="0%">
					{#each points as p, i (p.result.id)}
						<stop
							offset="{(i / (points.length - 1)) * 100}%"
							stop-color={getStatusColor(p.result.status, p.result.latency, slowThreshold)}
						/>
					{/each}
				</linearGradient>
				<filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
					<feGaussianBlur stdDeviation="3" result="blur" />
					<feComposite in="SourceGraphic" in2="blur" operator="over" />
				</filter>
			</defs>

			<!-- Grid Lines -->
			{#each gridLines as line (line)}
				<line
					x1="0"
					y1={300 * line}
					x2="1000"
					y2={300 * line}
					stroke="white"
					stroke-opacity="0.03"
					stroke-width="1"
				/>
			{/each}

			<path
				d={`M ${points[0].x},300 ` +
					points.map((p) => `L ${p.x},${p.y}`).join(' ') +
					` L ${points[points.length - 1].x},300 Z`}
				fill="url(#grad-large)"
			/>
			<path
				d={getPathD(points)}
				fill="none"
				stroke="url(#line-grad-large)"
				stroke-width="3"
				stroke-linejoin="round"
				filter="url(#glow)"
				class="opacity-90"
			/>

			{#if hoveredResult}
				{@const startTime = new Date(history[0].created_at).getTime()}
				{@const endTime = new Date(history[history.length - 1].created_at).getTime()}
				{@const hX =
					((new Date(hoveredResult.created_at).getTime() - startTime) /
						Math.max(endTime - startTime, 1)) *
					1000}
				<line
					x1={hX}
					y1="0"
					x2={hX}
					y2="300"
					stroke="#DEF4C6"
					stroke-width="1"
					stroke-dasharray="4,4"
					opacity="0.2"
				/>
			{/if}
		</svg>

		{#if hoveredResult}
			{@const maxLatency = Math.max(...history.map((r: CheckResult) => r.latency), 100) * 1.1}
			{@const startTime = new Date(history[0].created_at).getTime()}
			{@const endTime = new Date(history[history.length - 1].created_at).getTime()}
			{@const hX_p =
				((new Date(hoveredResult.created_at).getTime() - startTime) /
					Math.max(endTime - startTime, 1)) *
				100}
			{@const hY_p = (1 - hoveredResult.latency / maxLatency) * 100}

			<div
				class="pointer-events-none absolute z-10 h-4 w-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-brand-dark shadow-[0_0_15px_rgba(115,226,167,0.5)] transition-all duration-75"
				style="left: {hX_p}%; top: {hY_p}%; background-color: {getStatusColor(
					hoveredResult.status,
					hoveredResult.latency,
					slowThreshold
				)}"
			></div>

			<div
				class="pointer-events-none absolute z-10 flex min-w-[180px] flex-col gap-1 rounded-2xl border border-brand-light/10 bg-brand-dark/95 p-4 shadow-2xl backdrop-blur-xl transition-all duration-75"
				style="left: {hX_p > 80 ? 'auto' : `${hX_p}%`}; right: {hX_p > 80
					? '20px'
					: 'auto'}; top: {hY_p > 50 ? 'auto' : `${hY_p + 5}%`}; bottom: {hY_p > 50
					? `${100 - hY_p + 5}%`
					: 'auto'}; margin-left: {hX_p > 80 ? '0' : '20px'}"
			>
				<div class="text-[10px] font-black tracking-widest text-brand-light/40 uppercase">
					{formatDate(hoveredResult.created_at)}
				</div>
				<div class="mt-1 flex items-end justify-between gap-4">
					<span
						class="text-2xl leading-none font-black"
						style="color: {getStatusColor(
							hoveredResult.status,
							hoveredResult.latency,
							slowThreshold
						)}"
					>
						{hoveredResult.latency}<span class="ml-0.5 text-xs opacity-50">ms</span>
					</span>
					<span
						class="rounded-md border border-brand-light/10 bg-brand-light/5 px-2 py-0.5 text-[10px] font-black text-brand-light/60 uppercase"
					>
						{hoveredResult.status === 'Connected' ? 'Online' : hoveredResult.status}
					</span>
				</div>
				{#if hoveredResult.latency > slowThreshold}
					<div class="mt-2 flex items-center gap-1 text-[9px] font-bold text-brand-soft">
						<div class="h-1 w-1 animate-pulse rounded-full bg-brand-soft"></div>
						Latency threshold exceeded
					</div>
				{/if}
			</div>
		{/if}
	{:else}
		<div
			class="flex h-full items-center justify-center text-xs font-bold tracking-widest text-brand-light/20 uppercase italic"
		>
			Collecting telemetry data...
		</div>
	{/if}
</div>

<style>
	.status-chart {
		height: var(--chart-height);
	}

	@media (max-width: 640px) {
		.status-chart {
			height: min(var(--chart-height), 20rem);
		}
	}
</style>
