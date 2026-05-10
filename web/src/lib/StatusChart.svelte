<script lang="ts">
	import { getStatusColor, formatDate, type CheckResult } from '$lib/utils';

	let { history, height = 100 } = $props<{ history: CheckResult[], height?: number }>();

	let hoveredResult = $state<CheckResult | null>(null);

	function getChartPoints(history: CheckResult[], width: number, h: number) {
		if (!history || history.length < 2) return [];
		const maxLatency = Math.max(...history.map(r => r.latency), 100);
		// Add some padding at the top
		const chartMax = maxLatency * 1.1;
		const step = width / (history.length - 1);
		return history.map((r, i) => ({
			x: i * step,
			y: h - (r.latency / chartMax) * h,
			result: r
		}));
	}

	function getPathD(points: {x: number, y: number}[]) {
		if (points.length < 2) return '';
		return points.map((p, i) => (i === 0 ? `M ${p.x},${p.y}` : `L ${p.x},${p.y}`)).join(' ');
	}

	function handleMouseMove(e: MouseEvent) {
		if (!history || history.length === 0) return;
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const x = e.clientX - rect.left;
		const ratio = x / rect.width;
		const index = Math.min(Math.max(Math.round(ratio * (history.length - 1)), 0), history.length - 1);
		hoveredResult = history[index];
	}

	const gridLines = [0.25, 0.5, 0.75];
</script>

<div 
	class="relative w-full bg-brand-light/[0.01] rounded-[2rem] overflow-hidden cursor-crosshair border border-brand-light/5"
	style="height: {height}px"
	onmousemove={handleMouseMove}
	onmouseleave={() => hoveredResult = null}
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
					{#each points as p, i}
						<stop offset="{(i / (points.length - 1)) * 100}%" stop-color={getStatusColor(p.result.status, p.result.latency)} />
					{/each}
				</linearGradient>
				<filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
					<feGaussianBlur stdDeviation="3" result="blur" />
					<feComposite in="SourceGraphic" in2="blur" operator="over" />
				</filter>
			</defs>

			<!-- Grid Lines -->
			{#each gridLines as line}
				<line x1="0" y1={300 * line} x2="1000" y2={300 * line} stroke="white" stroke-opacity="0.03" stroke-width="1" />
			{/each}
			
			<path d={`M ${points[0].x},300 ` + points.map(p => `L ${p.x},${p.y}`).join(' ') + ` L ${points[points.length-1].x},300 Z`} fill="url(#grad-large)" />
			<path d={getPathD(points)} fill="none" stroke="url(#line-grad-large)" stroke-width="3" stroke-linejoin="round" filter="url(#glow)" class="opacity-90" />
			
			{#if hoveredResult}
				{@const hIndex = history.indexOf(hoveredResult)}
				{@const hX = (hIndex / (history.length - 1)) * 1000}
				<line x1={hX} y1="0" x2={hX} y2="300" stroke="#DEF4C6" stroke-width="1" stroke-dasharray="4,4" opacity="0.2" />
			{/if}
		</svg>

		{#if hoveredResult}
			{@const maxLatency = Math.max(...history.map((r: CheckResult) => r.latency), 100) * 1.1}
			{@const hIndex = history.indexOf(hoveredResult)}
			{@const hX_p = (hIndex / (history.length - 1)) * 100}
			{@const hY_p = (1 - hoveredResult.latency / maxLatency) * 100}
			
			<div 
				class="absolute w-4 h-4 rounded-full border-2 border-brand-dark shadow-[0_0_15px_rgba(115,226,167,0.5)] -translate-x-1/2 -translate-y-1/2 pointer-events-none z-10 transition-all duration-75"
				style="left: {hX_p}%; top: {hY_p}%; background-color: {getStatusColor(hoveredResult.status, hoveredResult.latency)}"
			></div>

			<div 
				class="absolute bg-brand-dark/95 border border-brand-light/10 p-4 rounded-2xl shadow-2xl backdrop-blur-xl pointer-events-none z-10 min-w-[180px] flex flex-col gap-1 transition-all duration-75" 
				style="left: {hX_p > 80 ? 'auto' : `${hX_p}%`}; right: {hX_p > 80 ? '20px' : 'auto'}; top: {hY_p > 50 ? 'auto' : `${hY_p + 5}%`}; bottom: {hY_p > 50 ? `${100 - hY_p + 5}%` : 'auto'}; margin-left: {hX_p > 80 ? '0' : '20px'}"
			>
				<div class="text-[10px] font-black text-brand-light/40 uppercase tracking-widest">{formatDate(hoveredResult.created_at)}</div>
				<div class="flex items-end justify-between gap-4 mt-1">
					<span class="text-2xl font-black leading-none" style="color: {getStatusColor(hoveredResult.status, hoveredResult.latency)}">
						{hoveredResult.latency}<span class="text-xs ml-0.5 opacity-50">ms</span>
					</span>
					<span class="text-[10px] uppercase font-black px-2 py-0.5 rounded-md bg-brand-light/5 text-brand-light/60 border border-brand-light/10">
						{hoveredResult.status === 'Connected' ? 'Online' : hoveredResult.status}
					</span>
				</div>
				{#if hoveredResult.latency > 300}
					<div class="mt-2 text-[9px] font-bold text-brand-soft flex items-center gap-1">
						<div class="h-1 w-1 rounded-full bg-brand-soft animate-pulse"></div>
						Latency threshold exceeded
					</div>
				{/if}
			</div>
		{/if}
	{:else}
		<div class="flex h-full items-center justify-center text-brand-light/20 italic font-bold uppercase tracking-widest text-xs">
			Collecting telemetry data...
		</div>
	{/if}
</div>
