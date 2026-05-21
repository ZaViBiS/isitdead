<script lang="ts">
	import { getStatusColor, formatDate, type CheckResult } from '$lib/utils';

	let {
		history,
		height = 100,
		slowThreshold = 300
	} = $props<{ history: CheckResult[]; height?: number; slowThreshold?: number }>();

	let hoveredResult = $state<CheckResult | null>(null);

	type ChartPoint = {
		x: number;
		latencyY: number;
		availabilityY: number;
		color: string;
		result: CheckResult;
	};

	const chartWidth = 1000;
	const chartHeight = 300;
	const latencyTop = 24;
	const latencyHeight = 190;
	const availabilityTop = 242;
	const availabilityHeight = 34;
	const gridLines = [0.25, 0.5, 0.75];

	function getAvailabilityY(result: CheckResult) {
		const color = getStatusColor(result.status, result.latency, slowThreshold);
		if (color === '#D62246') return availabilityTop + availabilityHeight;
		if (color === '#E5B181') return availabilityTop + availabilityHeight / 2;
		return availabilityTop;
	}

	function getChartPoints(history: CheckResult[], width: number) {
		if (!history || history.length < 2) return [];
		const maxLatency = Math.max(...history.map((r) => r.latency), 100);
		const chartMax = maxLatency * 1.1;
		const startTime = new Date(history[0].created_at).getTime();
		const endTime = new Date(history[history.length - 1].created_at).getTime();
		const duration = Math.max(endTime - startTime, 1);
		return history.map((r) => ({
			x: ((new Date(r.created_at).getTime() - startTime) / duration) * width,
			latencyY: latencyTop + (1 - r.latency / chartMax) * latencyHeight,
			availabilityY: getAvailabilityY(r),
			color: getStatusColor(r.status, r.latency, slowThreshold),
			result: r
		}));
	}

	function getLatencyPathD(points: ChartPoint[]) {
		if (points.length < 2) return '';
		return points
			.map((p, i) => (i === 0 ? `M ${p.x},${p.latencyY}` : `L ${p.x},${p.latencyY}`))
			.join(' ');
	}

	function getAvailabilitySegments(points: ChartPoint[]) {
		const segments: { d: string; color: string; key: string }[] = [];
		for (let i = 1; i < points.length; i += 1) {
			const previous = points[i - 1];
			const current = points[i];
			segments.push({
				d: `M ${previous.x},${previous.availabilityY} L ${current.x},${previous.availabilityY} L ${current.x},${current.availabilityY}`,
				color: current.color,
				key: `${previous.result.id}-${current.result.id}-${i}`
			});
		}
		return segments;
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

</script>

<div
	class="status-chart relative w-full cursor-crosshair overflow-hidden rounded-[2rem] border border-brand-light/5 bg-brand-light/[0.01]"
	style="--chart-height: {height}px"
	onmousemove={handleMouseMove}
	onmouseleave={() => (hoveredResult = null)}
	role="presentation"
>
	{#if history && history.length >= 2}
		{@const points = getChartPoints(history, chartWidth)}
		<svg class="h-full w-full" preserveAspectRatio="none" viewBox="0 0 1000 300">
			<defs>
				<filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
					<feGaussianBlur stdDeviation="3" result="blur" />
					<feComposite in="SourceGraphic" in2="blur" operator="over" />
				</filter>
			</defs>

			<!-- Grid Lines -->
			{#each gridLines as line (line)}
				<line
					x1="0"
					y1={chartHeight * line}
					x2={chartWidth}
					y2={chartHeight * line}
					stroke="white"
					stroke-opacity="0.03"
					stroke-width="1"
				/>
			{/each}

			<line
				x1="0"
				y1={availabilityTop - 12}
				x2={chartWidth}
				y2={availabilityTop - 12}
				stroke="white"
				stroke-opacity="0.06"
				stroke-width="1"
			/>
			<path
				d={getLatencyPathD(points)}
				fill="none"
				stroke="#73E2A7"
				stroke-width="3"
				stroke-linecap="round"
				stroke-linejoin="round"
				filter="url(#glow)"
				class="opacity-90"
			/>
			{#each getAvailabilitySegments(points) as segment (segment.key)}
				<path
					d={segment.d}
					fill="none"
					stroke={segment.color}
					stroke-width="3"
					stroke-linecap="round"
					stroke-linejoin="round"
					class="opacity-80"
				/>
			{/each}

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
					y2={chartHeight}
					stroke="#DEF4C6"
					stroke-width="1"
					stroke-dasharray="4,4"
					opacity="0.2"
				/>
			{/if}
		</svg>

		{#if hoveredResult}
			{@const hoveredPoint = points.find((p) => p.result === hoveredResult) ?? points[0]}
			{@const hX_p = (hoveredPoint.x / chartWidth) * 100}
			{@const hY_p = (hoveredPoint.latencyY / chartHeight) * 100}

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
