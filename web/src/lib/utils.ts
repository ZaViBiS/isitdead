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

export function supportsSlowThreshold(checkType: string) {
	return checkType !== 'ssl';
}

export function getEffectiveSlowThreshold(checkType: string, slowThreshold: number) {
	return supportsSlowThreshold(checkType) ? slowThreshold : Number.POSITIVE_INFINITY;
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

export function sampleChartHistory(
	history: CheckResult[],
	slowThreshold: number,
	maxPoints = 120
): CheckResult[] {
	if (history.length <= maxPoints) return history;

	const firstTime = new Date(history[0].created_at).getTime();
	const lastTime = new Date(history[history.length - 1].created_at).getTime();
	const bucketWidth = Math.max((lastTime - firstTime) / maxPoints, 1);
	const buckets = new Map<number, CheckResult[]>();

	for (const result of history) {
		const bucketIndex = Math.min(
			Math.floor((new Date(result.created_at).getTime() - firstTime) / bucketWidth),
			maxPoints - 1
		);
		const bucket = buckets.get(bucketIndex);
		if (bucket) {
			bucket.push(result);
		} else {
			buckets.set(bucketIndex, [result]);
		}
	}

	return [...buckets.values()].map((bucket) => {
		const averageLatency = Math.round(
			bucket.reduce((sum, result) => sum + result.latency, 0) / bucket.length
		);
		const worst = bucket.reduce((selected, result) => {
			const selectedState = getChartState(selected, slowThreshold);
			const resultState = getChartState(result, slowThreshold);
			const severity = { ok: 0, slow: 1, error: 2 };

			if (severity[resultState] > severity[selectedState]) return result;
			if (severity[resultState] === severity[selectedState] && result.latency > selected.latency) {
				return result;
			}
			return selected;
		});

		return {
			...worst,
			latency:
				getChartState(worst, slowThreshold) === 'ok'
					? averageLatency
					: Math.max(averageLatency, worst.latency)
		};
	});
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
