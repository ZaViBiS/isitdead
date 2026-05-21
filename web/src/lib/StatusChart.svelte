<script lang="ts">
	import { formatDate, type CheckResult } from '$lib/utils';

	type ChartSeries = {
		key: string;
		label: string;
		color: string;
		history: CheckResult[];
	};

	type HoveredPoint = {
		series: ChartSeries;
		result: CheckResult;
		x: number;
		y: number;
	};

	let {
		history,
		series = [],
		height = 100,
		slowThreshold = 300
	} = $props<{
		history?: CheckResult[];
		series?: ChartSeries[];
		height?: number;
		slowThreshold?: number;
	}>();

	let hoveredPoint = $state<HoveredPoint | null>(null);
	let items = $derived(visibleSeries());

	type ChartPoint = {
		x: number;
		y: number;
		result: CheckResult;
	};

	const chartWidth = 1000;
	const chartHeight = 300;
	const gridLines = [0.25, 0.5, 0.75];

	function visibleSeries() {
		const input: ChartSeries[] = series.length
			? series
			: history
				? [{ key: 'latency', label: 'Latency', color: '#73E2A7', history }]
				: [];
		return input.filter((item) => item.history.length >= 2);
	}

	function getTimeRange(items: ChartSeries[]) {
		const times = items.flatMap((item) =>
			item.history.map((result) => new Date(result.created_at).getTime())
		);
		return {
			start: Math.min(...times),
			end: Math.max(...times)
		};
	}

	function getMaxLatency(items: ChartSeries[]) {
		return Math.max(...items.flatMap((item) => item.history.map((r) => r.latency)), 100);
	}

	function getChartPoints(
		item: ChartSeries,
		range: { start: number; end: number },
		maxLatency: number
	) {
		const duration = Math.max(range.end - range.start, 1);
		const chartMax = maxLatency * 1.1;
		return item.history.map((r) => ({
			x: ((new Date(r.created_at).getTime() - range.start) / duration) * chartWidth,
			y: chartHeight - (r.latency / chartMax) * chartHeight,
			result: r
		}));
	}

	function getLatencyPathD(points: ChartPoint[]) {
		if (points.length < 2) return '';
		return points.map((p, i) => (i === 0 ? `M ${p.x},${p.y}` : `L ${p.x},${p.y}`)).join(' ');
	}

	function handleMouseMove(e: MouseEvent) {
		const items = visibleSeries();
		if (items.length === 0) return;
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const x = e.clientX - rect.left;
		const ratio = x / rect.width;
		const range = getTimeRange(items);
		const targetTime = range.start + ratio * (range.end - range.start);
		const maxLatency = getMaxLatency(items);

		const points: HoveredPoint[] = items.flatMap((item) =>
			getChartPoints(item, range, maxLatency).map((point) => ({
				series: item,
				result: point.result,
				x: point.x,
				y: point.y
			}))
		);
		hoveredPoint = points.reduce((closest, point) => {
			const closestDistance = Math.abs(new Date(closest.result.created_at).getTime() - targetTime);
			const nextDistance = Math.abs(new Date(point.result.created_at).getTime() - targetTime);
			return nextDistance < closestDistance ? point : closest;
		});
	}
</script>

<div
	class="status-chart relative w-full cursor-crosshair overflow-hidden rounded-[2rem] border border-brand-light/5 bg-brand-light/[0.01]"
	style="--chart-height: {height}px"
	onmousemove={handleMouseMove}
	onmouseleave={() => (hoveredPoint = null)}
	role="presentation"
>
	{#if items.length > 0}
		{@const range = getTimeRange(items)}
		{@const maxLatency = getMaxLatency(items)}
		<svg class="h-full w-full" preserveAspectRatio="none" viewBox="0 0 1000 300">
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

			{#each items as item (item.key)}
				{@const points = getChartPoints(item, range, maxLatency)}
				<path
					d={getLatencyPathD(points)}
					fill="none"
					stroke={item.color}
					stroke-width="3"
					stroke-linecap="round"
					stroke-linejoin="round"
					class="opacity-90"
				/>
			{/each}

			{#if hoveredPoint}
				<line
					x1={hoveredPoint.x}
					y1="0"
					x2={hoveredPoint.x}
					y2={chartHeight}
					stroke="#DEF4C6"
					stroke-width="1"
					stroke-dasharray="4,4"
					opacity="0.2"
				/>
			{/if}
		</svg>

		{#if hoveredPoint}
			{@const hX_p = (hoveredPoint.x / chartWidth) * 100}
			{@const hY_p = (hoveredPoint.y / chartHeight) * 100}

			<div
				class="pointer-events-none absolute z-10 h-4 w-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-brand-dark shadow-[0_0_15px_rgba(115,226,167,0.5)] transition-all duration-75"
				style="left: {hX_p}%; top: {hY_p}%; background-color: {hoveredPoint.series.color}"
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
					{hoveredPoint.series.label} · {formatDate(hoveredPoint.result.created_at)}
				</div>
				<div class="mt-1 flex items-end justify-between gap-4">
					<span class="text-2xl leading-none font-black" style="color: {hoveredPoint.series.color}">
						{hoveredPoint.result.latency}<span class="ml-0.5 text-xs opacity-50">ms</span>
					</span>
					<span
						class="rounded-md border border-brand-light/10 bg-brand-light/5 px-2 py-0.5 text-[10px] font-black text-brand-light/60 uppercase"
					>
						{hoveredPoint.result.status === 'Connected' ? 'Online' : hoveredPoint.result.status}
					</span>
				</div>
				{#if hoveredPoint.result.latency > slowThreshold}
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
