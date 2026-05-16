export interface CheckResult {
	id: number;
	status: string;
	latency: number;
	created_at: string;
}

export interface Server {
	id: number;
	name: string;
	url: string;
	check_type: string;
	public: boolean;
	public_slug: string;
	check_interval: number;
	timeout: number;
	slow_threshold: number;
	history: CheckResult[];
	history30d?: CheckResult[];
	incidents?: CheckResult[];
	check_count_30d?: number;
	uptime_30d?: number;
	avg_latency_30d?: number;
	current_status?: string;
	current_latency?: number;
	hourly_buckets?: DashboardBucket[];
}

export type DashboardBucket = 'ok' | 'slow' | 'error' | 'empty';

export interface NotificationPreference {
	id?: number;
	user_id?: number;
	server_id?: number;
	channel: string;
	event: string;
	enabled: boolean;
	destination?: string;
}

export function getStatusColor(status: string, latency: number, slowThreshold = 300): string {
	if (!status) return '#D62246';
	if (!(status.startsWith('2') || status === 'Connected')) return '#D62246';
	if (latency > slowThreshold) return '#E5B181';
	return '#73E2A7';
}

export function getFaviconUrl(url: string): string {
	try {
		const domain = new URL(url.startsWith('http') ? url : `http://${url}`).hostname;
		return `https://www.google.com/s2/favicons?domain=${domain}&sz=128`;
	} catch {
		return '';
	}
}

export function formatDate(dateStr: string): string {
	const date = new Date(dateStr);
	return new Intl.DateTimeFormat('en-US', {
		day: '2-digit',
		month: '2-digit',
		year: 'numeric',
		hour: '2-digit',
		minute: '2-digit',
		second: '2-digit',
		hour12: false
	}).format(date);
}

export function calculateUptime(history: CheckResult[]) {
	if (!history || history.length === 0) return 0;
	const online = history.filter((r) => r.status.startsWith('2') || r.status === 'Connected').length;
	return (online / history.length) * 100;
}

export function calculateAvgLatency(history: CheckResult[]) {
	if (!history || history.length === 0) return 0;
	const sum = history.reduce((acc, r) => acc + r.latency, 0);
	return Math.round(sum / history.length);
}

export function getLatestCheck(history?: CheckResult[]): CheckResult | null {
	return history && history.length > 0 ? history[history.length - 1] : null;
}

export function getCurrentCheck(server: Server): CheckResult | null {
	if (server.current_status) {
		return {
			id: 0,
			status: server.current_status,
			latency: server.current_latency ?? 0,
			created_at: ''
		};
	}
	return getLatestCheck(server.history30d ?? server.history);
}

export function getDashboardBucketColor(bucket: DashboardBucket): string {
	if (bucket === 'ok') return '#73E2A7';
	if (bucket === 'slow') return '#E5B181';
	if (bucket === 'error') return '#D62246';
	return '#1f332f';
}

export function getRecentHistory(history: CheckResult[], hours: number): CheckResult[] {
	const since = Date.now() - hours * 60 * 60 * 1000;
	return history.filter((result) => new Date(result.created_at).getTime() >= since);
}

function getChartState(result: CheckResult, slowThreshold: number) {
	if (!(result.status.startsWith('2') || result.status === 'Connected')) return 'error';
	if (result.latency > slowThreshold) return 'slow';
	return 'ok';
}

function largestTriangleThreeBuckets(history: CheckResult[], threshold: number): CheckResult[] {
	if (threshold >= history.length || threshold <= 2) return history;

	const sampled: CheckResult[] = [history[0]];
	const bucketSize = (history.length - 2) / (threshold - 2);
	let selectedIndex = 0;

	for (let bucket = 0; bucket < threshold - 2; bucket++) {
		const avgStart = Math.floor((bucket + 1) * bucketSize) + 1;
		const avgEnd = Math.min(Math.floor((bucket + 2) * bucketSize) + 1, history.length);
		const avgRange = history.slice(avgStart, avgEnd);
		const avgX =
			avgRange.reduce((sum, point) => sum + new Date(point.created_at).getTime(), 0) /
			Math.max(avgRange.length, 1);
		const avgY =
			avgRange.reduce((sum, point) => sum + point.latency, 0) / Math.max(avgRange.length, 1);

		const rangeStart = Math.floor(bucket * bucketSize) + 1;
		const rangeEnd = Math.min(Math.floor((bucket + 1) * bucketSize) + 1, history.length - 1);
		const anchor = history[selectedIndex];
		const anchorX = new Date(anchor.created_at).getTime();

		let maxArea = -1;
		let nextIndex = rangeStart;

		for (let index = rangeStart; index < rangeEnd; index++) {
			const point = history[index];
			const pointX = new Date(point.created_at).getTime();
			const area = Math.abs(
				(anchorX - avgX) * (point.latency - anchor.latency) -
					(anchorX - pointX) * (avgY - anchor.latency)
			);
			if (area > maxArea) {
				maxArea = area;
				nextIndex = index;
			}
		}

		sampled.push(history[nextIndex]);
		selectedIndex = nextIndex;
	}

	sampled.push(history[history.length - 1]);
	return sampled;
}

export function sampleChartHistory(
	history: CheckResult[],
	slowThreshold: number,
	maxPoints = 220
): CheckResult[] {
	if (history.length <= maxPoints) return history;

	const importantIDs = new Set<number>([history[0].id, history[history.length - 1].id]);
	let peak = history[0];

	for (let index = 1; index < history.length; index++) {
		const previous = history[index - 1];
		const current = history[index];
		if (current.latency > peak.latency) peak = current;

		if (getChartState(previous, slowThreshold) !== getChartState(current, slowThreshold)) {
			importantIDs.add(previous.id);
			importantIDs.add(current.id);
		}
	}

	importantIDs.add(peak.id);

	const importantCount = importantIDs.size;
	const baselineBudget = Math.max(maxPoints - importantCount, 3);
	const sampled = largestTriangleThreeBuckets(history, baselineBudget);
	const merged = new Map<number, CheckResult>();

	for (const result of sampled) merged.set(result.id, result);
	for (const result of history) {
		if (importantIDs.has(result.id)) merged.set(result.id, result);
	}

	return [...merged.values()].sort(
		(left, right) => new Date(left.created_at).getTime() - new Date(right.created_at).getTime()
	);
}

export function getHourlyBuckets(
	history: CheckResult[],
	nowMs = Date.now(),
	slowThreshold = 300
): string[] {
	const buckets: string[] = Array(24).fill('#1f332f');
	const windowStart = nowMs - 24 * 60 * 60 * 1000;
	const hourMs = 60 * 60 * 1000;

	for (const result of history) {
		const createdAt = new Date(result.created_at).getTime();
		if (createdAt < windowStart || createdAt >= nowMs) continue;

		const bucketIndex = Math.floor((createdAt - windowStart) / hourMs);
		const color = getStatusColor(result.status, result.latency, slowThreshold);
		const current = buckets[bucketIndex];

		if (current === '#D62246') continue;
		if (color === '#D62246' || current === '#1f332f') {
			buckets[bucketIndex] = color;
			continue;
		}
		if (color === '#E5B181') {
			buckets[bucketIndex] = color;
		}
	}

	return buckets;
}
