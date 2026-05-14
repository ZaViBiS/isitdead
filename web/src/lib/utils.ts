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
	history: CheckResult[];
	history30d?: CheckResult[];
	incidents?: CheckResult[];
}

export interface NotificationPreference {
	id?: number;
	user_id?: number;
	server_id?: number;
	channel: string;
	event: string;
	enabled: boolean;
	destination?: string;
}

export function getStatusColor(status: string, latency: number): string {
	if (!status) return '#D62246';
	if (!(status.startsWith('2') || status === 'Connected')) return '#D62246';
	if (latency > 300) return '#E5B181';
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
	return getLatestCheck(server.history30d ?? server.history);
}

export function getRecentHistory(history: CheckResult[], hours: number): CheckResult[] {
	const since = Date.now() - hours * 60 * 60 * 1000;
	return history.filter((result) => new Date(result.created_at).getTime() >= since);
}

export function getHourlyBuckets(history: CheckResult[], nowMs = Date.now()): string[] {
	const buckets: string[] = Array(24).fill('#1f332f');

	for (let i = 0; i < 24; i++) {
		const hourStart = nowMs - (24 - i) * 60 * 60 * 1000;
		const hourEnd = nowMs - (23 - i) * 60 * 60 * 1000;

		const hourResults = history.filter((h) => {
			const d = new Date(h.created_at).getTime();
			return d >= hourStart && d < hourEnd;
		});

		if (hourResults.length > 0) {
			let worstColor = '#73E2A7';
			for (const res of hourResults) {
				const color = getStatusColor(res.status, res.latency);
				if (color === '#D62246') {
					worstColor = '#D62246';
					break;
				}
				if (color === '#E5B181') {
					worstColor = '#E5B181';
				}
			}
			buckets[i] = worstColor;
		}
	}

	return buckets;
}
