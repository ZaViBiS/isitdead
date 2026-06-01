<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import {
		Activity,
		Plus,
		Trash2,
		ExternalLink,
		RefreshCw,
		AlertCircle,
		Globe2,
		ShieldCheck,
		Settings,
		Mail,
		MessageCircle,
		X,
		LockKeyhole,
		ShieldAlert
	} from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		getDashboardBucketColor,
		getCurrentCheck,
		supportsSlowThreshold,
		getFaviconUrl,
		formatDateOnly,
		type Server,
		type NotificationPreference,
		type User,
		type BillingPlan
	} from '$lib/utils';

	type TelegramStatus = {
		linked: boolean;
		linked_at?: string;
		bot_name?: string;
		link_available: boolean;
	};

	let servers = $state<Server[]>([]);
	let user = $state<User | null>(null);
	let billingPlans = $state<BillingPlan[]>([]);
	let isLoading = $state(true);
	let isAdding = $state(false);
	let isEditing = $state(false);
	let error = $state('');

	let newName = $state('');
	let newUrl = $state('');
	let newType = $state('http');
	let newInterval = $state(300);
	let newTimeout = $state(10);
	let newSlowThreshold = $state(300);
	let newSSLEnabled = $state(false);
	let newNotifyEmailDown = $state(true);
	let newNotifyEmailRecovered = $state(true);
	let newNotifyTelegramDown = $state(false);
	let newNotifyTelegramRecovered = $state(false);
	let newNotifyDiscordDown = $state(false);
	let newNotifyDiscordRecovered = $state(false);
	let telegramLinkUrl = $state('');
	let isCreatingTelegramLink = $state(false);
	let telegramStatus = $state<TelegramStatus>({ linked: false, link_available: false });
	let telegramStatusMessage = $state('');
	let telegramPolling = $state(false);
	let telegramPollTimer: ReturnType<typeof setInterval> | null = null;

	let editingServer = $state<Server | null>(null);
	let editName = $state('');
	let editUrl = $state('');
	let editType = $state('http');
	let editInterval = $state(300);
	let editTimeout = $state(10);
	let editSlowThreshold = $state(300);
	let editSSLEnabled = $state(false);
	let editNotifyEmailDown = $state(true);
	let editNotifyEmailRecovered = $state(true);
	let editNotifyTelegramDown = $state(false);
	let editNotifyTelegramRecovered = $state(false);
	let editNotifyDiscordDown = $state(false);
	let editNotifyDiscordRecovered = $state(false);

	const checkTypeOptions = [
		{
			value: 'http',
			label: 'HTTP',
			description: 'Websites and API endpoints',
			icon: Globe2
		},
		{
			value: 'ping',
			label: 'TCP',
			description: 'Ports and service sockets',
			icon: Activity
		}
	];

	function frontendCheckType(checkType: string) {
		return checkType === 'ping' ? 'ping' : 'http';
	}

	function safeExternalHref(rawUrl: string) {
		try {
			const parsed = new URL(rawUrl);
			return parsed.protocol === 'http:' || parsed.protocol === 'https:' ? parsed.href : '';
		} catch {
			return '';
		}
	}

	async function fetchServers() {
		const token = localStorage.getItem('token');
		if (!token) {
			goto(resolve('/login'));
			return;
		}

		try {
			const res = await fetch('/api/dashboard/servers', {
				headers: { Authorization: `Bearer ${token}` }
			});

			if (res.ok) {
				const data = (await res.json()) as Server[];
				servers = data.map((s) => ({ ...s, history: [] }));
			} else if (res.status === 401) {
				localStorage.removeItem('token');
				goto(resolve('/login'));
			}
		} catch {
			error = 'Connection error';
		} finally {
			isLoading = false;
		}
	}

	async function fetchBilling() {
		const token = localStorage.getItem('token');
		try {
			const [meRes, plansRes] = await Promise.all([
				token
					? fetch('/api/me', { headers: { Authorization: `Bearer ${token}` } })
					: Promise.resolve(null),
				fetch('/api/billing/plans')
			]);
			if (meRes?.ok) user = (await meRes.json()) as User;
			if (plansRes.ok) billingPlans = (await plansRes.json()) as BillingPlan[];
		} catch {
			// Billing details are non-critical for loading monitor data.
		}
	}

	function normalizeMonitorUrl(url: string, checkType: string) {
		const trimmed = url.trim();
		if (checkType !== 'http' || trimmed === '') return trimmed;
		if (trimmed.startsWith('http://') || trimmed.startsWith('https://')) return trimmed;
		return `https://${trimmed}`;
	}

	function willNormalizeMonitorUrl(url: string, checkType: string) {
		const trimmed = url.trim();
		return (
			checkType === 'http' &&
			trimmed !== '' &&
			!trimmed.startsWith('http://') &&
			!trimmed.startsWith('https://')
		);
	}

	function getTargetPlaceholder(checkType: string) {
		if (checkType === 'ping') return 'example.com:80';
		return 'example.com or http://example.com';
	}

	function selectNewType(checkType: string) {
		newType = frontendCheckType(checkType);
		if (newType !== 'http') newSSLEnabled = false;
	}

	function selectEditType(checkType: string) {
		editType = frontendCheckType(checkType);
		if (editType !== 'http') editSSLEnabled = false;
	}

	function notificationPayload(
		emailDown: boolean,
		emailRecovered: boolean,
		telegramDown: boolean,
		telegramRecovered: boolean,
		discordDown: boolean,
		discordRecovered: boolean
	): NotificationPreference[] {
		return [
			{ channel: 'email', event: 'down', enabled: emailDown },
			{ channel: 'email', event: 'recovered', enabled: emailRecovered },
			{ channel: 'telegram', event: 'down', enabled: telegramDown },
			{ channel: 'telegram', event: 'recovered', enabled: telegramRecovered },
			{ channel: 'discord', event: 'down', enabled: discordDown },
			{ channel: 'discord', event: 'recovered', enabled: discordRecovered }
		];
	}

	function preferenceEnabled(
		prefs: NotificationPreference[],
		channel: string,
		event: string,
		fallback: boolean
	) {
		return (
			prefs.find((pref) => pref.channel === channel && pref.event === event)?.enabled ?? fallback
		);
	}

	async function fetchNotificationPreferences(serverID: number) {
		const token = localStorage.getItem('token');
		const res = await fetch(`/api/servers/${serverID}/notifications`, {
			headers: { Authorization: `Bearer ${token}` }
		});
		if (!res.ok) throw new Error('Failed to load notification preferences');
		return (await res.json()) as NotificationPreference[];
	}

	async function saveNotificationPreferences(
		serverID: number,
		emailDown: boolean,
		emailRecovered: boolean,
		telegramDown: boolean,
		telegramRecovered: boolean,
		discordDown: boolean,
		discordRecovered: boolean
	) {
		const token = localStorage.getItem('token');
		const res = await fetch(`/api/servers/${serverID}/notifications`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
			body: JSON.stringify(
				notificationPayload(
					emailDown,
					emailRecovered,
					telegramDown,
					telegramRecovered,
					discordDown,
					discordRecovered
				)
			)
		});
		if (!res.ok) throw new Error('Failed to save notification preferences');
	}

	function sleep(ms: number) {
		return new Promise((resolve) => setTimeout(resolve, ms));
	}

	async function waitForFirstCheck(serverID: number, token: string) {
		for (let attempt = 0; attempt < 6; attempt += 1) {
			try {
				const res = await fetch(`/api/servers/${serverID}/results?limit=1`, {
					headers: { Authorization: `Bearer ${token}` }
				});
				if (res.ok) {
					const results = (await res.json()) as unknown[];
					if (results.length > 0) return;
				}
			} catch {
				// The dashboard refresh below still shows the monitor if the first check is delayed.
			}
			await sleep(500);
		}
	}

	async function fetchTelegramStatus(silent = false) {
		const token = localStorage.getItem('token');
		if (!token) return;
		try {
			const res = await fetch('/api/telegram/status', {
				headers: { Authorization: `Bearer ${token}` }
			});
			if (!res.ok) throw new Error('Failed to load Telegram status');
			telegramStatus = (await res.json()) as TelegramStatus;
			if (telegramStatus.linked) {
				telegramStatusMessage = 'Telegram is connected.';
				stopTelegramStatusPolling();
			} else if (!silent) {
				telegramStatusMessage = '';
			}
		} catch {
			if (!silent) error = 'Failed to load Telegram status';
		}
	}

	function stopTelegramStatusPolling() {
		if (telegramPollTimer) {
			clearInterval(telegramPollTimer);
			telegramPollTimer = null;
		}
		telegramPolling = false;
	}

	function startTelegramStatusPolling() {
		stopTelegramStatusPolling();
		telegramPolling = true;
		let attempts = 0;
		telegramPollTimer = setInterval(async () => {
			attempts += 1;
			await fetchTelegramStatus(true);
			if (telegramStatus.linked || attempts >= 45) {
				if (!telegramStatus.linked) {
					telegramStatusMessage = 'Still waiting for Telegram confirmation.';
				}
				stopTelegramStatusPolling();
			}
		}, 2000);
	}

	function handleTelegramAction() {
		if (telegramStatus.linked) {
			fetchTelegramStatus();
			return;
		}
		createTelegramLink();
	}

	async function createTelegramLink() {
		const token = localStorage.getItem('token');
		isCreatingTelegramLink = true;
		telegramLinkUrl = '';
		telegramStatusMessage = '';
		try {
			const res = await fetch('/api/telegram/link-token', {
				method: 'POST',
				headers: { Authorization: `Bearer ${token}` }
			});
			if (!res.ok) throw new Error('Failed to create Telegram link');
			const data = (await res.json()) as {
				url?: string;
				token: string;
				bot_name?: string;
				link_available: boolean;
			};
			telegramLinkUrl = data.url ?? data.token;
			telegramStatus = {
				...telegramStatus,
				bot_name: data.bot_name ?? telegramStatus.bot_name,
				link_available: data.link_available
			};
			if (data.link_available) {
				telegramStatusMessage = 'Waiting for Telegram confirmation...';
				startTelegramStatusPolling();
			} else {
				telegramStatusMessage = 'TELEGRAM_BOT_NAME is missing on the server.';
			}
		} catch {
			error = 'Failed to create Telegram link';
		} finally {
			isCreatingTelegramLink = false;
		}
	}

	function isServerOnline(server: Server) {
		const current = getCurrentCheck(server);
		return current?.status.startsWith('2') === true || current?.status === 'Connected';
	}

	function getOverallUptime() {
		const serversWithData = servers.filter((server) => (server.check_count_30d ?? 0) > 0);
		if (serversWithData.length === 0) return 0;
		const total = serversWithData.reduce((sum, server) => sum + (server.uptime_30d ?? 0), 0);
		return total / serversWithData.length;
	}

	function getOverallLatency() {
		const serversWithData = servers.filter((server) => (server.check_count_30d ?? 0) > 0);
		if (serversWithData.length === 0) return 0;
		const total = serversWithData.reduce((sum, server) => sum + (server.avg_latency_30d ?? 0), 0);
		return Math.round(total / serversWithData.length);
	}

	function getHealthyCount() {
		return servers.filter(isServerOnline).length;
	}

	function getAttentionCount() {
		return Math.max(servers.length - getHealthyCount(), 0);
	}

	function openAdd() {
		telegramLinkUrl = '';
		isAdding = true;
	}

	function currentPlan() {
		const planID = user?.plan ?? 'free';
		return billingPlans.find((plan) => plan.id === planID) ?? billingPlans[0];
	}

	function monitorLimitLabel() {
		const plan = currentPlan();
		if (!plan) return `${servers.length} monitors`;
		return `${servers.length}/${plan.monitor_limit} monitors`;
	}

	function allowedIntervalPresets() {
		const minInterval = currentPlan()?.min_interval ?? 30;
		return intervalPresets.filter((preset) => preset >= minInterval);
	}

	function getSSLLabel(server: Server) {
		const status = server.ssl_status;
		if (!status) return 'SSL';
		if (status.self_signed) return 'SSL self-signed';
		if (!status.valid) return 'SSL invalid';
		return `SSL ${status.days_remaining}d`;
	}

	function getSSLBadgeClass(server: Server) {
		const status = server.ssl_status;
		if (!status) return 'border-brand-light/10 bg-brand-light/[0.04] text-brand-light/35';
		if (status.self_signed) return 'border-brand-soft/20 bg-brand-soft/10 text-brand-soft';
		if (!status.valid || status.days_remaining < 7) {
			return 'border-brand-accent/20 bg-brand-accent/10 text-brand-accent';
		}
		if (status.days_remaining <= 14) return 'border-brand-gold/20 bg-brand-gold/10 text-brand-gold';
		if (status.days_remaining <= 30)
			return 'border-brand-primary/20 bg-brand-primary/10 text-brand-primary';
		return 'border-brand-primary/20 bg-brand-primary/10 text-brand-primary';
	}

	async function addServer(e: SubmitEvent) {
		e.preventDefault();
		const token = localStorage.getItem('token');
		try {
			const res = await fetch('/api/servers', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({
					name: newName,
					url: normalizeMonitorUrl(newUrl, frontendCheckType(newType)),
					check_type: frontendCheckType(newType),
					check_interval: Number(newInterval),
					timeout: Number(newTimeout),
					slow_threshold: Number(newSlowThreshold),
					ssl_enabled: newSSLEnabled
				})
			});

			if (res.ok) {
				const server = (await res.json()) as Server;
				await saveNotificationPreferences(
					server.id,
					newNotifyEmailDown,
					newNotifyEmailRecovered,
					newNotifyTelegramDown,
					newNotifyTelegramRecovered,
					newNotifyDiscordDown,
					newNotifyDiscordRecovered
				);
				await waitForFirstCheck(server.id, token ?? '');
				isAdding = false;
				newName = '';
				newUrl = '';
				newType = 'http';
				newInterval = 300;
				newTimeout = 10;
				newSlowThreshold = 300;
				newSSLEnabled = false;
				newNotifyEmailDown = true;
				newNotifyEmailRecovered = true;
				newNotifyTelegramDown = false;
				newNotifyTelegramRecovered = false;
				newNotifyDiscordDown = false;
				newNotifyDiscordRecovered = false;
				await fetchServers();
				await fetchBilling();
			} else {
				const data = await res.json();
				error = data.error ?? 'Failed to add server';
			}
		} catch {
			error = 'Failed to add server';
		}
	}

	async function openEdit(server: Server) {
		editingServer = server;
		editName = server.name;
		editUrl = server.url;
		editType = frontendCheckType(server.check_type);
		editInterval = server.check_interval;
		editTimeout = server.timeout;
		editSlowThreshold = server.slow_threshold;
		editSSLEnabled = server.ssl_enabled && editType === 'http';
		editNotifyEmailDown = true;
		editNotifyEmailRecovered = true;
		editNotifyTelegramDown = false;
		editNotifyTelegramRecovered = false;
		telegramLinkUrl = '';
		isEditing = true;

		try {
			const prefs = await fetchNotificationPreferences(server.id);
			editNotifyEmailDown = preferenceEnabled(prefs, 'email', 'down', true);
			editNotifyEmailRecovered = preferenceEnabled(prefs, 'email', 'recovered', true);
			editNotifyTelegramDown = preferenceEnabled(prefs, 'telegram', 'down', false);
			editNotifyTelegramRecovered = preferenceEnabled(prefs, 'telegram', 'recovered', false);
			editNotifyDiscordDown = preferenceEnabled(prefs, 'discord', 'down', false);
			editNotifyDiscordRecovered = preferenceEnabled(prefs, 'discord', 'recovered', false);
		} catch {
			error = 'Failed to load notification preferences';
		}
	}

	async function updateServer(e: SubmitEvent) {
		e.preventDefault();
		if (!editingServer) return;
		const token = localStorage.getItem('token');
		try {
			const res = await fetch(`/api/servers/${editingServer.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({
					name: editName,
					url: normalizeMonitorUrl(editUrl, frontendCheckType(editType)),
					check_type: frontendCheckType(editType),
					check_interval: Number(editInterval),
					timeout: Number(editTimeout),
					slow_threshold: Number(editSlowThreshold),
					ssl_enabled: editSSLEnabled
				})
			});

			if (res.ok) {
				const updated = await res.json();
				const idx = servers.findIndex((s) => s.id === updated.id);
				if (idx !== -1) {
					// Зберігаємо історію, яку ми вже завантажили
					servers[idx] = {
						...servers[idx],
						...updated,
						history: servers[idx].history,
						history30d: servers[idx].history30d
					};
				}
				await saveNotificationPreferences(
					editingServer.id,
					editNotifyEmailDown,
					editNotifyEmailRecovered,
					editNotifyTelegramDown,
					editNotifyTelegramRecovered,
					editNotifyDiscordDown,
					editNotifyDiscordRecovered
				);
				isEditing = false;
				editingServer = null;
			} else {
				const data = await res.json();
				error = data.error ?? 'Failed to update server';
			}
		} catch {
			error = 'Failed to update server';
		}
	}

	async function deleteServer(id: number) {
		if (!confirm('Are you sure you want to delete this monitor?')) return;
		const token = localStorage.getItem('token');
		try {
			const res = await fetch(`/api/servers/${id}`, {
				method: 'DELETE',
				headers: { Authorization: `Bearer ${token}` }
			});
			if (res.ok) {
				const idx = servers.findIndex((s) => s.id === id);
				if (idx !== -1) servers.splice(idx, 1);
			}
		} catch {
			error = 'Failed to delete server';
		}
	}

	const intervalPresets = [
		30, 60, 180, 300, 600, 1800, 3600, 7200, 14400, 21600, 36000, 43200, 86400
	];
	const timeoutPresets = [3, 5, 10, 15, 30, 60];

	function formatInterval(seconds: number) {
		if (seconds < 60) return `${seconds}s`;
		const m = Math.floor(seconds / 60);
		const s = seconds % 60;
		if (m < 60) return s > 0 ? `${m}m ${s}s` : `${m}m`;
		const h = Math.floor(m / 60);
		const mm = m % 60;
		return mm > 0 ? `${h}h ${mm}m` : `${h}h`;
	}

	onMount(() => {
		fetchServers();
		fetchBilling();
		fetchTelegramStatus(true);
	});

	onDestroy(stopTelegramStatusPolling);
</script>

<div class="relative isolate min-h-[calc(100vh-4rem)] overflow-hidden">
	<div class="pointer-events-none absolute inset-0 -z-10">
		<div
			class="absolute inset-0 bg-[linear-gradient(to_right,rgba(222,244,198,0.035)_1px,transparent_1px),linear-gradient(to_bottom,rgba(222,244,198,0.035)_1px,transparent_1px)] [mask-image:linear-gradient(to_bottom,black,transparent_86%)] bg-[size:64px_64px]"
		></div>
		<div class="absolute inset-x-0 top-0 h-64 bg-brand-light/[0.025]"></div>
	</div>

	<div class="container mx-auto max-w-7xl px-4 py-6 sm:px-6 sm:py-10 lg:py-12">
		<section class="glass-panel mb-6 overflow-hidden rounded-[2.5rem] p-4 sm:p-5">
			<div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_22rem] lg:items-stretch">
				<div class="rounded-[2rem] border border-brand-light/10 bg-brand-dark/60 p-5 sm:p-6">
					<div class="signal-pill mb-5 py-1.5">
						<span class="relative flex h-2 w-2">
							<span
								class="absolute inline-flex h-full w-full animate-ping rounded-full bg-brand-primary opacity-40"
							></span>
							<span class="relative inline-flex h-2 w-2 rounded-full bg-brand-primary"></span>
						</span>
						Live operations
					</div>
					<div class="flex flex-col gap-5 xl:flex-row xl:items-end xl:justify-between">
						<div class="min-w-0">
							<h1
								class="display-gradient text-3xl leading-[0.95] font-black tracking-[-0.06em] sm:text-5xl"
							>
								Your monitors
							</h1>
							<p
								class="mt-4 flex max-w-2xl items-start gap-2 text-sm leading-6 text-brand-light/55 sm:text-base"
							>
								<ShieldCheck class="mt-1 h-4 w-4 shrink-0 text-brand-primary" />
								Track uptime, response time, and recent failures in one focused workspace.
							</p>
						</div>
						<div class="flex w-full flex-col gap-3 sm:w-auto sm:flex-row sm:items-center">
							<a
								href={resolve('/pricing')}
								class="inline-flex items-center justify-center rounded-2xl border border-brand-light/10 bg-brand-light/[0.035] px-4 py-3 text-sm font-black text-brand-light/65 transition hover:border-brand-primary/30 hover:text-brand-primary"
							>
								{currentPlan()?.name ?? 'Free'} · {monitorLimitLabel()}
							</a>
							<button
								onclick={() => (isAdding ? (isAdding = false) : openAdd())}
								class="inline-flex items-center justify-center gap-2 rounded-2xl bg-brand-primary px-5 py-3 font-black text-brand-dark shadow-2xl shadow-brand-primary/15 transition hover:-translate-y-0.5 hover:bg-brand-primary/90 active:translate-y-0"
							>
								<Plus class="h-5 w-5" /> Add monitor
							</button>
						</div>
					</div>
				</div>

				<div class="grid gap-3">
					<div
						class="flex items-center justify-between gap-4 rounded-3xl border px-4 py-3 {getAttentionCount() >
						0
							? 'border-brand-accent/20 bg-brand-accent/10'
							: 'border-brand-primary/20 bg-brand-primary/10'}"
					>
						<div
							class="flex items-center gap-2 {getAttentionCount() > 0
								? 'text-brand-accent'
								: 'text-brand-primary'}"
						>
							<span
								class="micro-label {getAttentionCount() > 0
									? '!text-brand-accent'
									: '!text-brand-primary'}"
							>
								Needs attention
							</span>
							<Activity class="h-4 w-4" />
						</div>
						<div class="flex items-center gap-2">
							<div
								class="text-xl font-black {getAttentionCount() > 0
									? 'text-brand-accent'
									: 'text-brand-primary'}"
							>
								{isLoading ? '--' : getAttentionCount()}
							</div>
							<div class="text-sm font-bold text-brand-light/45">
								{isLoading ? 'Loading' : getAttentionCount() > 0 ? 'offline now' : 'all online'}
							</div>
						</div>
					</div>

					<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-1 xl:grid-cols-2">
						<div class="soft-panel rounded-[2rem] p-5">
							<div class="micro-label mb-4">Avg uptime</div>
							<div class="text-3xl font-black text-brand-primary">
								{isLoading ? '--' : `${getOverallUptime().toFixed(1)}%`}
							</div>
						</div>
						<div class="soft-panel rounded-[2rem] p-5">
							<div class="micro-label mb-4">Avg response</div>
							<div class="text-3xl font-black text-brand-light/85">
								{isLoading ? '--' : getOverallLatency()}<span
									class="ml-1 text-sm text-brand-light/30">{isLoading ? '' : 'ms'}</span
								>
							</div>
						</div>
						<div class="soft-panel rounded-[2rem] p-5 sm:col-span-2 lg:col-span-1 xl:col-span-2">
							<div class="mb-3 flex items-center justify-between gap-3">
								<div class="micro-label">Telegram</div>
								<span
									class="rounded-full px-2.5 py-1 text-[10px] font-black tracking-widest uppercase {telegramStatus.linked
										? 'bg-brand-primary/10 text-brand-primary'
										: 'bg-brand-light/5 text-brand-light/35'}"
								>
									{telegramStatus.linked ? 'Connected' : 'Not connected'}
								</span>
							</div>
							<div class="text-sm font-bold text-brand-light/55">
								{telegramStatus.linked
									? telegramStatus.linked_at
										? `Linked ${formatDateOnly(telegramStatus.linked_at)}`
										: 'Alerts can be sent to Telegram'
									: telegramStatus.link_available
										? 'Ready to connect'
										: 'Bot name is not configured'}
							</div>
						</div>
					</div>
				</div>
			</div>
		</section>

		{#if error}
			<div
				class="animate-in fade-in slide-in-from-top-4 mb-6 flex items-center gap-3 rounded-2xl border border-brand-accent/20 bg-brand-accent/10 p-4 text-brand-accent"
			>
				<AlertCircle class="h-5 w-5 shrink-0" />
				{error}
			</div>
		{/if}

		{#if isAdding}
			<section
				class="mobile-form-panel glass-panel animate-in zoom-in-95 fixed inset-0 z-[60] flex h-dvh min-w-0 flex-col overflow-hidden rounded-none duration-200 sm:static sm:z-auto sm:mb-6 sm:h-auto sm:rounded-[1.75rem]"
			>
				<div
					class="flex shrink-0 items-center justify-between gap-3 border-b border-brand-light/10 bg-brand-panel/95 px-4 py-4 backdrop-blur-sm sm:bg-transparent sm:px-6 sm:py-5 sm:backdrop-blur-none"
				>
					<div class="flex items-center gap-3">
						<div class="rounded-2xl bg-brand-primary/10 p-2.5 text-brand-primary">
							<Activity class="h-5 w-5 sm:h-6 sm:w-6" />
						</div>
						<div>
							<div class="signal-kicker mb-1">Monitor composer</div>
							<h2 class="text-xl font-black sm:text-2xl">Add monitor</h2>
						</div>
					</div>
					<button
						type="button"
						onclick={() => (isAdding = false)}
						class="rounded-xl p-2 transition hover:bg-brand-light/5 sm:hidden"
						aria-label="Close add monitor form"
					>
						<X class="h-5 w-5 text-brand-light/35" />
					</button>
				</div>
				<form
					onsubmit={addServer}
					class="grid min-w-0 flex-1 gap-4 overflow-x-hidden overflow-y-auto p-4 pb-4 sm:flex-none sm:overflow-visible sm:p-6 sm:pb-6 lg:grid-cols-[minmax(0,1.2fr)_minmax(21rem,0.8fr)]"
				>
					<div class="soft-panel min-w-0 rounded-[2rem] p-4 sm:p-5">
						<div class="mb-5 flex items-start justify-between gap-4">
							<div>
								<div class="micro-label mb-2">Target</div>
								<h3 class="text-lg font-black">What should be monitored</h3>
							</div>
							<div
								class="rounded-2xl bg-brand-primary/10 px-3 py-2 text-xs font-black text-brand-primary"
							>
								{newType === 'ping' ? 'TCP' : newType.toUpperCase()}
							</div>
						</div>
						<div class="grid gap-4 md:grid-cols-[0.9fr_1.1fr]">
							<div class="min-w-0 space-y-2">
								<label for="name" class="ml-1 text-xs font-bold text-brand-light/45 uppercase"
									>Monitor name</label
								>
								<input
									id="name"
									type="text"
									bind:value={newName}
									required
									class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
									placeholder="Production API"
								/>
							</div>
							<div class="space-y-2">
								<label for="url" class="ml-1 text-xs font-bold text-brand-light/45 uppercase"
									>URL or host</label
								>
								<input
									id="url"
									type="text"
									inputmode="url"
									bind:value={newUrl}
									required
									class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
									placeholder={getTargetPlaceholder(newType)}
								/>
								{#if willNormalizeMonitorUrl(newUrl, newType)}
									<p
										class="rounded-2xl border border-brand-primary/20 bg-brand-primary/10 px-4 py-3 text-sm font-medium text-brand-primary"
									>
										We will save this as <span class="font-black"
											>{normalizeMonitorUrl(newUrl, newType)}</span
										>.
									</p>
								{/if}
							</div>
							<div class="space-y-2 md:col-span-2">
								<div class="ml-1 text-xs font-bold text-brand-light/45 uppercase">Check type</div>
								<div class="grid gap-2 sm:grid-cols-2">
									{#each checkTypeOptions as option (option.value)}
										{@const TypeIcon = option.icon}
										<button
											type="button"
											onclick={() => selectNewType(option.value)}
											class="group rounded-2xl border p-3 text-left transition {newType ===
											option.value
												? 'border-brand-primary/35 bg-brand-primary/10 shadow-lg shadow-brand-primary/10'
												: 'border-brand-light/10 bg-brand-dark/40 hover:border-brand-light/20 hover:bg-brand-light/[0.035]'}"
										>
											<div class="mb-3 flex items-center justify-between gap-3">
												<span
													class="rounded-xl p-2 {newType === option.value
														? 'bg-brand-primary text-brand-dark'
														: 'bg-brand-light/5 text-brand-light/45 group-hover:text-brand-light/70'}"
												>
													<TypeIcon class="h-4 w-4" />
												</span>
												{#if newType === option.value}
													<span class="h-2 w-2 rounded-full bg-brand-primary"></span>
												{/if}
											</div>
											<div class="text-sm font-black">{option.label}</div>
											<div class="mt-1 text-xs leading-5 text-brand-light/40">
												{option.description}
											</div>
										</button>
									{/each}
								</div>
							</div>
							{#if newType === 'http'}
								<label
									class="flex cursor-pointer items-center justify-between gap-4 rounded-2xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3 md:col-span-2"
								>
									<span>
										<span class="block text-sm font-black">Monitor SSL certificate</span>
										<span class="mt-1 block text-xs text-brand-light/40">
											Daily certificate check with reminders 30, 14, and 7 days before expiry.
										</span>
									</span>
									<input
										type="checkbox"
										bind:checked={newSSLEnabled}
										class="h-5 w-5 accent-brand-primary"
									/>
								</label>
							{/if}
						</div>
					</div>

					<div class="soft-panel min-w-0 rounded-[2rem] p-4 sm:p-5">
						<div class="mb-5">
							<div class="micro-label mb-2">Rules</div>
							<h3 class="text-lg font-black">When should it alert</h3>
						</div>
						<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-1 xl:grid-cols-2">
							{#if supportsSlowThreshold(newType)}
								<div class="space-y-2">
									<div class="ml-1 flex items-center justify-between">
										<label for="timeout" class="text-xs font-bold text-brand-light/45 uppercase"
											>Timeout</label
										>
										<span
											class="rounded-full bg-brand-primary/10 px-2.5 py-1 text-xs font-black whitespace-nowrap text-brand-primary"
											>{newTimeout}s</span
										>
									</div>
									<select
										id="timeout"
										bind:value={newTimeout}
										class="w-full cursor-pointer rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
									>
										{#each timeoutPresets as preset (preset)}
											<option value={preset}>{preset}s</option>
										{/each}
									</select>
								</div>
								<div class="space-y-2">
									<div class="ml-1 flex items-center justify-between">
										<label
											for="slow-threshold"
											class="text-xs font-bold text-brand-light/45 uppercase">Slow threshold</label
										>
										<span
											class="rounded-full bg-brand-gold/10 px-2.5 py-1 text-xs font-black whitespace-nowrap text-brand-gold"
											>{newSlowThreshold}ms</span
										>
									</div>
									<input
										id="slow-threshold"
										type="number"
										min="1"
										bind:value={newSlowThreshold}
										required
										class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
									/>
								</div>
							{/if}
							<div class="min-w-0 space-y-2 sm:col-span-2 lg:col-span-1 xl:col-span-2">
								<div class="ml-1 flex items-center justify-between">
									<label for="interval" class="text-xs font-bold text-brand-light/45 uppercase"
										>Interval</label
									>
									<span
										class="rounded-full bg-brand-primary/10 px-2.5 py-1 text-xs font-black whitespace-nowrap text-brand-primary"
										>{formatInterval(newInterval)}</span
									>
								</div>
								<div class="grid grid-cols-4 gap-2 sm:grid-cols-5 lg:grid-cols-4 xl:grid-cols-5">
									{#each allowedIntervalPresets() as preset (preset)}
										<button
											type="button"
											onclick={() => (newInterval = preset)}
											class="rounded-xl border px-3 py-2 text-xs font-bold whitespace-nowrap transition {newInterval ===
											preset
												? 'border-brand-primary bg-brand-primary text-brand-dark shadow-lg shadow-brand-primary/20'
												: 'border-brand-light/10 bg-brand-light/5 text-brand-light/45 hover:border-brand-light/20'}"
										>
											{formatInterval(preset)}
										</button>
									{/each}
								</div>
							</div>
						</div>
					</div>
					<div class="soft-panel grid gap-3 rounded-[2rem] p-4 sm:grid-cols-2 sm:p-5 lg:col-span-2">
						<div class="flex items-center gap-3 sm:col-span-2">
							<Mail class="h-4 w-4 text-brand-primary" />
							<div>
								<div class="micro-label">Notifications</div>
								<div class="mt-1 text-sm font-bold text-brand-light/75">
									Choose what deserves an alert
								</div>
							</div>
						</div>
						<label
							class="flex cursor-pointer items-center justify-between gap-4 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3"
						>
							<span class="text-sm font-bold text-brand-light/75">Email when down</span>
							<input
								type="checkbox"
								bind:checked={newNotifyEmailDown}
								class="h-5 w-5 accent-brand-primary"
							/>
						</label>
						<label
							class="flex cursor-pointer items-center justify-between gap-4 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3"
						>
							<span class="text-sm font-bold text-brand-light/75">Email when recovered</span>
							<input
								type="checkbox"
								bind:checked={newNotifyEmailRecovered}
								class="h-5 w-5 accent-brand-primary"
							/>
						</label>
						<label
							class="flex cursor-pointer items-center justify-between gap-4 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3 {telegramStatus.linked
								? ''
								: 'opacity-55'}"
						>
							<span class="text-sm font-bold text-brand-light/75">Telegram when down</span>
							<input
								type="checkbox"
								bind:checked={newNotifyTelegramDown}
								disabled={!telegramStatus.linked}
								class="h-5 w-5 accent-brand-primary"
							/>
						</label>
						<label
							class="flex cursor-pointer items-center justify-between gap-4 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3 {telegramStatus.linked
								? ''
								: 'opacity-55'}"
						>
							<span class="text-sm font-bold text-brand-light/75">Telegram when recovered</span>
							<input
								type="checkbox"
								bind:checked={newNotifyTelegramRecovered}
								disabled={!telegramStatus.linked}
								class="h-5 w-5 accent-brand-primary"
							/>
						</label>
						<div
							class="flex flex-col gap-3 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3 sm:col-span-2"
						>
							<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
								<div class="flex items-center gap-3">
									<MessageCircle
										class="h-4 w-4 {telegramStatus.linked
											? 'text-brand-primary'
											: 'text-brand-light/35'}"
									/>
									<div>
										<div class="text-sm font-black text-brand-light/80">Telegram</div>
										<div
											class="mt-1 text-xs font-bold {telegramStatus.linked
												? 'text-brand-primary'
												: telegramPolling
													? 'text-brand-gold'
													: 'text-brand-light/35'}"
										>
											{telegramStatus.linked
												? 'Connected'
												: telegramPolling
													? 'Waiting for confirmation'
													: 'Not connected'}
										</div>
									</div>
								</div>
								<div class="flex flex-col gap-2 sm:flex-row sm:items-center">
									{#if telegramLinkUrl && !telegramStatus.linked}
										{#if telegramLinkUrl.startsWith('http')}
											<a
												href={telegramLinkUrl}
												target="_blank"
												rel="external noreferrer"
												class="max-w-full truncate rounded-xl border border-brand-primary/20 bg-brand-primary/10 px-3 py-2 text-xs font-black text-brand-primary"
												>Open Telegram</a
											>
										{:else}
											<span
												class="max-w-full truncate rounded-xl border border-brand-gold/20 bg-brand-gold/10 px-3 py-2 text-xs font-black text-brand-gold"
												>{telegramLinkUrl}</span
											>
										{/if}
									{/if}
									<button
										type="button"
										onclick={handleTelegramAction}
										disabled={isCreatingTelegramLink || telegramPolling}
										class="rounded-xl border border-brand-light/10 px-4 py-2 text-xs font-black transition hover:bg-brand-light/5 disabled:cursor-wait disabled:opacity-50"
									>
										{telegramStatus.linked
											? 'Refresh'
											: isCreatingTelegramLink
												? 'Creating...'
												: telegramPolling
													? 'Waiting...'
													: 'Get link'}
									</button>
								</div>
							</div>
							{#if telegramStatusMessage}
								<div
									class="rounded-xl border px-3 py-2 text-xs font-bold {telegramStatus.linked
										? 'border-brand-primary/20 bg-brand-primary/10 text-brand-primary'
										: telegramPolling
											? 'border-brand-gold/20 bg-brand-gold/10 text-brand-gold'
											: 'border-brand-light/10 bg-brand-light/5 text-brand-light/45'}"
								>
									{telegramStatusMessage}
								</div>
							{/if}
						</div>
					</div>
					<div
						class="sticky bottom-0 z-10 mt-1 flex min-w-0 flex-col-reverse gap-3 border-t border-brand-light/10 bg-brand-panel/95 pt-4 pb-[calc(env(safe-area-inset-bottom)+0.25rem)] backdrop-blur-sm sm:static sm:mt-0 sm:flex-row sm:justify-end sm:border-0 sm:bg-transparent sm:p-0 sm:pt-2 sm:backdrop-blur-none lg:col-span-2"
					>
						<button
							type="button"
							onclick={() => (isAdding = false)}
							class="w-full rounded-2xl border border-brand-light/10 px-6 py-3 font-bold transition hover:bg-brand-light/5 sm:w-auto"
							>Cancel</button
						>
						<button
							type="submit"
							class="w-full rounded-2xl bg-brand-primary px-7 py-3 font-black text-brand-dark shadow-lg shadow-brand-primary/10 transition hover:bg-brand-primary/90 sm:w-auto"
							>Start monitoring</button
						>
					</div>
				</form>
			</section>
		{/if}

		{#if isEditing && editingServer}
			<div
				class="animate-in fade-in fixed inset-0 z-50 flex items-start justify-center bg-brand-dark/80 backdrop-blur-sm duration-200 sm:items-center sm:p-4"
			>
				<div
					class="mobile-form-panel glass-panel animate-in zoom-in-95 flex h-dvh w-full min-w-0 flex-col overflow-hidden rounded-none duration-200 sm:h-auto sm:max-h-[calc(100dvh-2rem)] sm:max-w-2xl sm:rounded-[1.75rem] lg:p-0"
				>
					<div
						class="flex shrink-0 items-start justify-between gap-4 border-b border-brand-light/10 bg-brand-panel/95 p-4 backdrop-blur-sm sm:border-0 sm:bg-transparent sm:p-6 sm:pb-0 sm:backdrop-blur-none lg:p-8 lg:pb-0"
					>
						<div class="flex items-center gap-3">
							<div class="rounded-2xl bg-brand-primary/10 p-2.5 text-brand-primary sm:p-3">
								<Settings class="h-5 w-5 sm:h-6 sm:w-6" />
							</div>
							<div>
								<h2 class="text-xl font-black sm:text-2xl">Edit monitor</h2>
								<p class="hidden text-sm text-brand-light/40 sm:block">
									Update the target and check schedule.
								</p>
							</div>
						</div>
						<button
							onclick={() => (isEditing = false)}
							class="rounded-xl p-2 transition hover:bg-brand-light/5"
						>
							<X class="h-6 w-6 text-brand-light/30 hover:text-brand-light" />
						</button>
					</div>

					<form
						onsubmit={updateServer}
						class="grid min-w-0 flex-1 gap-5 overflow-x-hidden overflow-y-auto p-4 pb-4 sm:flex-none sm:overflow-visible sm:p-6 sm:pt-5 sm:pb-6 lg:p-8 lg:pt-5"
					>
						<div class="grid gap-4">
							<div class="soft-panel rounded-[2rem] p-4 sm:p-5">
								<div class="mb-5 flex items-start justify-between gap-4">
									<div>
										<div class="micro-label mb-2">Target</div>
										<h3 class="text-lg font-black">What is monitored</h3>
									</div>
									<div
										class="rounded-2xl bg-brand-primary/10 px-3 py-2 text-xs font-black text-brand-primary"
									>
										{editType === 'ping' ? 'TCP' : editType.toUpperCase()}
									</div>
								</div>
								<div class="grid gap-4 md:grid-cols-2">
									{#if supportsSlowThreshold(editType)}
										<div class="space-y-2">
											<label
												for="edit-name"
												class="ml-1 text-xs font-bold text-brand-light/45 uppercase"
												>Monitor name</label
											>
											<input
												id="edit-name"
												type="text"
												bind:value={editName}
												required
												class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
											/>
										</div>
									{/if}
									<div class="space-y-2">
										<label
											for="edit-url"
											class="ml-1 text-xs font-bold text-brand-light/45 uppercase"
											>URL or host</label
										>
										<input
											id="edit-url"
											type="text"
											inputmode="url"
											bind:value={editUrl}
											required
											class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
										/>
										{#if willNormalizeMonitorUrl(editUrl, editType)}
											<p
												class="rounded-2xl border border-brand-primary/20 bg-brand-primary/10 px-4 py-3 text-sm font-medium text-brand-primary"
											>
												We will save this as <span class="font-black"
													>{normalizeMonitorUrl(editUrl, editType)}</span
												>.
											</p>
										{/if}
									</div>
									<div class="space-y-2 md:col-span-2">
										<div class="ml-1 text-xs font-bold text-brand-light/45 uppercase">
											Check type
										</div>
										<div class="grid gap-2 sm:grid-cols-2">
											{#each checkTypeOptions as option (option.value)}
												{@const TypeIcon = option.icon}
												<button
													type="button"
													onclick={() => selectEditType(option.value)}
													class="group rounded-2xl border p-3 text-left transition {editType ===
													option.value
														? 'border-brand-primary/35 bg-brand-primary/10 shadow-lg shadow-brand-primary/10'
														: 'border-brand-light/10 bg-brand-dark/40 hover:border-brand-light/20 hover:bg-brand-light/[0.035]'}"
												>
													<div class="mb-3 flex items-center justify-between gap-3">
														<span
															class="rounded-xl p-2 {editType === option.value
																? 'bg-brand-primary text-brand-dark'
																: 'bg-brand-light/5 text-brand-light/45 group-hover:text-brand-light/70'}"
														>
															<TypeIcon class="h-4 w-4" />
														</span>
														{#if editType === option.value}
															<span class="h-2 w-2 rounded-full bg-brand-primary"></span>
														{/if}
													</div>
													<div class="text-sm font-black">{option.label}</div>
													<div class="mt-1 text-xs leading-5 text-brand-light/40">
														{option.description}
													</div>
												</button>
											{/each}
										</div>
									</div>
									{#if editType === 'http'}
										<label
											class="flex cursor-pointer items-center justify-between gap-4 rounded-2xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3 md:col-span-2"
										>
											<span>
												<span class="block text-sm font-black">Monitor SSL certificate</span>
												<span class="mt-1 block text-xs text-brand-light/40">
													Daily certificate check with reminders 30, 14, and 7 days before expiry.
												</span>
											</span>
											<input
												type="checkbox"
												bind:checked={editSSLEnabled}
												class="h-5 w-5 accent-brand-primary"
											/>
										</label>
									{/if}
								</div>
							</div>

							<div class="soft-panel rounded-[2rem] p-4 sm:p-5">
								<div class="mb-5">
									<div class="micro-label mb-2">Rules</div>
									<h3 class="text-lg font-black">When it should alert</h3>
								</div>
								<div class="grid gap-4 md:grid-cols-2">
									<div class="min-w-0 space-y-2 md:col-span-2">
										<div class="ml-1 flex items-center justify-between">
											<label
												for="edit-interval"
												class="text-xs font-bold text-brand-light/45 uppercase">Interval</label
											>
											<span
												class="rounded-full bg-brand-primary/10 px-2.5 py-1 text-xs font-black whitespace-nowrap text-brand-primary"
												>{formatInterval(editInterval)}</span
											>
										</div>
										<div class="flex max-w-full gap-2 overflow-x-auto pb-1 sm:flex-wrap">
											{#each allowedIntervalPresets() as preset (preset)}
												<button
													type="button"
													onclick={() => (editInterval = preset)}
													class="min-w-14 rounded-xl border px-3 py-2 text-xs font-bold whitespace-nowrap transition {editInterval ===
													preset
														? 'border-brand-primary bg-brand-primary text-brand-dark shadow-lg shadow-brand-primary/20'
														: 'border-brand-light/10 bg-brand-light/5 text-brand-light/45 hover:border-brand-light/20'}"
												>
													{formatInterval(preset)}
												</button>
											{/each}
										</div>
									</div>
									<div class="space-y-2">
										<div class="ml-1 flex items-center justify-between">
											<label
												for="edit-timeout"
												class="text-xs font-bold text-brand-light/45 uppercase">Timeout</label
											>
											<span
												class="rounded-full bg-brand-primary/10 px-2.5 py-1 text-xs font-black whitespace-nowrap text-brand-primary"
												>{editTimeout}s</span
											>
										</div>
										<select
											id="edit-timeout"
											bind:value={editTimeout}
											class="w-full cursor-pointer rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
										>
											{#each timeoutPresets as preset (preset)}
												<option value={preset}>{preset}s</option>
											{/each}
										</select>
									</div>
									<div class="space-y-2">
										<div class="ml-1 flex items-center justify-between">
											<label
												for="edit-slow-threshold"
												class="text-xs font-bold text-brand-light/45 uppercase"
												>Slow threshold</label
											>
											<span
												class="rounded-full bg-brand-primary/10 px-2.5 py-1 text-xs font-black whitespace-nowrap text-brand-primary"
												>{editSlowThreshold}ms</span
											>
										</div>
										<input
											id="edit-slow-threshold"
											type="number"
											min="1"
											bind:value={editSlowThreshold}
											required
											class="w-full rounded-2xl border border-brand-light/10 bg-brand-dark/60 px-4 py-3 transition outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary"
										/>
									</div>
								</div>
							</div>
							<div class="soft-panel grid gap-3 rounded-[2rem] p-4 md:grid-cols-2 md:p-5">
								<div class="flex items-center gap-3 md:col-span-2">
									<Mail class="h-4 w-4 text-brand-primary" />
									<div>
										<div class="micro-label">Notifications</div>
										<div class="mt-1 text-sm font-bold text-brand-light/75">
											Choose what deserves an alert
										</div>
									</div>
								</div>
								<label
									class="flex cursor-pointer items-center justify-between gap-4 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3"
								>
									<span class="text-sm font-bold text-brand-light/75">Email when down</span>
									<input
										type="checkbox"
										bind:checked={editNotifyEmailDown}
										class="h-5 w-5 accent-brand-primary"
									/>
								</label>
								<label
									class="flex cursor-pointer items-center justify-between gap-4 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3"
								>
									<span class="text-sm font-bold text-brand-light/75">Email when recovered</span>
									<input
										type="checkbox"
										bind:checked={editNotifyEmailRecovered}
										class="h-5 w-5 accent-brand-primary"
									/>
								</label>
								<label
									class="flex cursor-pointer items-center justify-between gap-4 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3 {telegramStatus.linked
										? ''
										: 'opacity-55'}"
								>
									<span class="text-sm font-bold text-brand-light/75">Telegram when down</span>
									<input
										type="checkbox"
										bind:checked={editNotifyTelegramDown}
										disabled={!telegramStatus.linked}
										class="h-5 w-5 accent-brand-primary"
									/>
								</label>
								<label
									class="flex cursor-pointer items-center justify-between gap-4 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3 {telegramStatus.linked
										? ''
										: 'opacity-55'}"
								>
									<span class="text-sm font-bold text-brand-light/75">Telegram when recovered</span>
									<input
										type="checkbox"
										bind:checked={editNotifyTelegramRecovered}
										disabled={!telegramStatus.linked}
										class="h-5 w-5 accent-brand-primary"
									/>
								</label>
								<div
									class="flex flex-col gap-3 rounded-xl border border-brand-light/10 bg-brand-dark/40 px-4 py-3 md:col-span-2"
								>
									<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
										<div class="flex items-center gap-3">
											<MessageCircle
												class="h-4 w-4 {telegramStatus.linked
													? 'text-brand-primary'
													: 'text-brand-light/35'}"
											/>
											<div>
												<div class="text-sm font-black text-brand-light/80">Telegram</div>
												<div
													class="mt-1 text-xs font-bold {telegramStatus.linked
														? 'text-brand-primary'
														: telegramPolling
															? 'text-brand-gold'
															: 'text-brand-light/35'}"
												>
													{telegramStatus.linked
														? 'Connected'
														: telegramPolling
															? 'Waiting for confirmation'
															: 'Not connected'}
												</div>
											</div>
										</div>
										<div class="flex flex-col gap-2 sm:flex-row sm:items-center">
											{#if telegramLinkUrl && !telegramStatus.linked}
												{#if telegramLinkUrl.startsWith('http')}
													<a
														href={telegramLinkUrl}
														target="_blank"
														rel="external noreferrer"
														class="max-w-full truncate rounded-xl border border-brand-primary/20 bg-brand-primary/10 px-3 py-2 text-xs font-black text-brand-primary"
														>Open Telegram</a
													>
												{:else}
													<span
														class="max-w-full truncate rounded-xl border border-brand-gold/20 bg-brand-gold/10 px-3 py-2 text-xs font-black text-brand-gold"
														>{telegramLinkUrl}</span
													>
												{/if}
											{/if}
											<button
												type="button"
												onclick={handleTelegramAction}
												disabled={isCreatingTelegramLink || telegramPolling}
												class="rounded-xl border border-brand-light/10 px-4 py-2 text-xs font-black transition hover:bg-brand-light/5 disabled:cursor-wait disabled:opacity-50"
											>
												{telegramStatus.linked
													? 'Refresh'
													: isCreatingTelegramLink
														? 'Creating...'
														: telegramPolling
															? 'Waiting...'
															: 'Get link'}
											</button>
										</div>
									</div>
									{#if telegramStatusMessage}
										<div
											class="rounded-xl border px-3 py-2 text-xs font-bold {telegramStatus.linked
												? 'border-brand-primary/20 bg-brand-primary/10 text-brand-primary'
												: telegramPolling
													? 'border-brand-gold/20 bg-brand-gold/10 text-brand-gold'
													: 'border-brand-light/10 bg-brand-light/5 text-brand-light/45'}"
										>
											{telegramStatusMessage}
										</div>
									{/if}
								</div>
							</div>
						</div>
						<div
							class="sticky bottom-0 z-10 mt-1 flex min-w-0 flex-col-reverse gap-3 border-t border-brand-light/10 bg-brand-panel/95 pt-4 pb-[calc(env(safe-area-inset-bottom)+0.25rem)] backdrop-blur-sm sm:static sm:mt-0 sm:flex-row sm:justify-end sm:bg-transparent sm:p-0 sm:pt-4 sm:backdrop-blur-none"
						>
							<button
								type="button"
								onclick={() => (isEditing = false)}
								class="w-full rounded-2xl border border-brand-light/10 px-7 py-3 font-bold transition hover:bg-brand-light/5 sm:w-auto"
								>Cancel</button
							>
							<button
								type="submit"
								class="w-full rounded-2xl bg-brand-primary px-8 py-3 font-black text-brand-dark shadow-lg shadow-brand-primary/10 transition hover:bg-brand-primary/90 sm:w-auto"
								>Save changes</button
							>
						</div>
					</form>
				</div>
			</div>
		{/if}

		{#if isLoading}
			<div
				class="soft-panel flex min-h-80 flex-col items-center justify-center gap-4 rounded-[2rem]"
			>
				<RefreshCw class="h-9 w-9 animate-spin text-brand-primary" />
				<p class="animate-pulse font-medium text-brand-light/30">Loading monitors...</p>
			</div>
		{:else if servers.length === 0}
			<section
				class="glass-panel flex flex-col items-center justify-center rounded-[2.25rem] border-dashed px-5 py-20 text-center sm:py-28"
			>
				<div
					class="mb-6 rounded-3xl border border-brand-light/10 bg-brand-light/[0.04] p-5 text-brand-primary"
				>
					<Activity class="h-12 w-12" />
				</div>
				<h3 class="mb-2 text-2xl font-black">No monitors yet</h3>
				<p class="mx-auto max-w-sm text-sm leading-6 text-brand-light/40">
					Add your first website, API endpoint, or TCP service to start tracking uptime.
				</p>
				<button
					onclick={openAdd}
					class="mt-8 inline-flex w-full items-center justify-center gap-2 rounded-2xl bg-brand-primary px-6 py-3 font-black text-brand-dark shadow-lg shadow-brand-primary/20 transition hover:-translate-y-0.5 hover:bg-brand-primary/90 active:translate-y-0 sm:w-auto"
				>
					<Plus class="h-5 w-5" /> Add monitor
				</button>
			</section>
		{:else}
			<section class="grid gap-4">
				<div class="mb-1 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
					<div>
						<p class="signal-kicker">Monitor fleet</p>
						<h2 class="mt-2 text-2xl font-black tracking-tight sm:text-3xl">Active services</h2>
					</div>
					<p class="max-w-xl text-sm leading-6 text-brand-light/40">
						Current health, 24-hour pulse, and the fastest path into each monitor.
					</p>
				</div>
				{#each servers as s (s.id)}
					{@const uptime = s.uptime_30d ?? 0}
					{@const avgLatency = s.avg_latency_30d ?? 0}
					{@const current = getCurrentCheck(s)}
					{@const currentStatus = current?.status ?? 'unknown'}
					{@const currentLatency = current?.latency ?? 0}
					{@const isOnline = currentStatus.startsWith('2') || currentStatus === 'Connected'}
					{@const isUnknown = currentStatus === 'unknown'}

					<article
						class="rounded-3xl border border-brand-light/10 bg-brand-dark/70 p-4 transition hover:border-brand-primary/25"
					>
						<div
							class="grid gap-4 lg:grid-cols-[minmax(18rem,22rem)_minmax(0,1fr)_auto] lg:items-center"
						>
							<div class="order-2 min-w-0 lg:order-1">
								<div class="mb-3 flex min-w-0 flex-wrap items-center gap-2">
									<div class="flex min-w-0 flex-wrap items-center gap-2">
										<span
											class="rounded-lg border border-brand-light/10 bg-brand-light/[0.04] px-2 py-1 text-[10px] font-black tracking-widest whitespace-nowrap text-brand-light/35 uppercase"
										>
											every {formatInterval(s.check_interval)}
										</span>
										{#if supportsSlowThreshold(s.check_type)}
											<span
												class="rounded-lg border border-brand-gold/15 bg-brand-gold/10 px-2 py-1 text-[10px] font-black tracking-widest whitespace-nowrap text-brand-gold uppercase"
											>
												slow &gt; {s.slow_threshold}ms
											</span>
										{/if}
										{#if s.ssl_enabled}
											<span class="group/ssl relative">
												<button
													type="button"
													class="inline-flex items-center gap-1 rounded-lg border px-2 py-1 text-[10px] font-black tracking-widest whitespace-nowrap uppercase {getSSLBadgeClass(
														s
													)}"
												>
													{#if s.ssl_status?.self_signed}
														<ShieldAlert class="h-3 w-3" />
													{:else}
														<LockKeyhole class="h-3 w-3" />
													{/if}
													{getSSLLabel(s)}
												</button>
												<div
													class="pointer-events-none absolute bottom-full left-0 z-40 mb-2 hidden w-64 rounded-2xl border border-brand-light/10 bg-brand-dark/95 p-4 text-left text-xs leading-5 text-brand-light/60 shadow-2xl group-focus-within/ssl:block group-hover/ssl:block"
												>
													<div class="mb-2 font-black text-brand-light/85">
														SSL certificate status
													</div>
													<div>Green: more than 30 days left</div>
													<div>Mint: 30 days or less</div>
													<div>Gold: 14 days or less</div>
													<div>Red: under 7 days, expired, or invalid</div>
													<div class="text-brand-soft">Pink: self-signed certificate</div>
													{#if s.ssl_status?.expires_at}
														<div class="mt-2 border-t border-brand-light/10 pt-2">
															Expires: {formatDateOnly(s.ssl_status.expires_at)}
														</div>
													{/if}
												</div>
											</span>
										{/if}
									</div>
								</div>

								<div>
									<div class="flex h-8 w-full items-end gap-1">
										{#each s.hourly_buckets ?? [] as bucket, i (i)}
											{@const color = getDashboardBucketColor(bucket)}
											<div
												class="group/item relative flex-1 cursor-help rounded-sm opacity-70 transition hover:opacity-100"
												style="background-color: {color}; height: {color === '#39414D'
													? '38%'
													: '100%'}"
											>
												<div
													class="pointer-events-none absolute bottom-full left-1/2 z-50 mb-3 hidden -translate-x-1/2 group-hover/item:block"
												>
													<div
														class="rounded-xl border border-brand-light/10 bg-brand-dark/95 px-3 py-2 text-[11px] whitespace-nowrap shadow-2xl ring-1 ring-white/5 backdrop-blur-xl"
													>
														<div class="mb-1 font-black text-brand-light/40">{23 - i}h ago</div>
														{#if bucket !== 'empty'}
															<div class="flex items-center gap-2">
																<div
																	class="h-2 w-2 rounded-full"
																	style="background-color: {color}"
																></div>
																<span class="font-bold"
																	>{bucket === 'ok'
																		? 'Healthy'
																		: bucket === 'slow'
																			? 'Slow response'
																			: 'Down'}</span
																>
															</div>
														{:else}
															<div class="text-brand-light/20 italic">No check data</div>
														{/if}
													</div>
													<div
														class="mx-auto -mt-1 h-2 w-2 rotate-45 border-r border-b border-brand-light/10 bg-brand-dark"
													></div>
												</div>
											</div>
										{/each}
									</div>
									<div
										class="mt-2 flex justify-between text-[10px] font-bold text-brand-light/25 uppercase"
									>
										<span>24h ago</span>
										<span>Now</span>
									</div>
								</div>
							</div>

							<div class="order-1 min-w-0 lg:order-2">
								<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
									<div class="flex min-w-0 items-center gap-4">
										<div
											class="relative flex h-12 w-12 shrink-0 items-center justify-center overflow-hidden rounded-2xl border border-brand-light/10 bg-brand-light/[0.04] text-brand-primary"
										>
											{#if s.check_type === 'ssl'}
												<LockKeyhole class="h-6 w-6" />
											{:else if s.check_type === 'http' || s.check_type === 'links'}
												<Globe2 class="h-6 w-6" />
											{:else}
												<Activity class="h-6 w-6" />
											{/if}
											<img
												src={getFaviconUrl(s.url)}
												alt=""
												class="absolute h-8 w-8 rounded-md bg-brand-dark object-contain"
												onerror={(e) => {
													const target = e.currentTarget as HTMLImageElement;
													target.style.display = 'none';
												}}
											/>
										</div>
										<div class="min-w-0">
											<div class="flex flex-wrap items-center gap-2">
												<h3 class="min-w-0 truncate font-black">{s.name}</h3>
												<span
													class="rounded-lg border border-brand-light/10 bg-brand-light/[0.04] px-2 py-0.5 text-[10px] font-black tracking-widest text-brand-light/35 uppercase"
												>
													{s.check_type}
												</span>
												{#if s.public && s.public_slug}
													<a
														href={resolve('/status/[slug]', { slug: s.public_slug })}
														class="inline-flex items-center gap-1.5 rounded-lg border border-brand-primary/15 bg-brand-primary/10 px-2 py-1 text-[10px] font-black tracking-widest text-brand-primary uppercase"
													>
														Public page <ExternalLink class="h-3 w-3" />
													</a>
												{/if}
											</div>
											<div class="mt-1 flex min-w-0 items-center gap-2">
												<p
													class="max-w-[16rem] truncate text-sm font-medium text-brand-light/35 sm:max-w-xs"
												>
													{s.url}
												</p>
												{#if safeExternalHref(s.url)}
													<a
														href={safeExternalHref(s.url)}
														target="_blank"
														rel="external noreferrer"
														class="shrink-0 rounded-lg p-1 text-brand-light/25 transition hover:bg-brand-light/5 hover:text-brand-primary"
														title="Open target"
													>
														<ExternalLink class="h-3.5 w-3.5" />
													</a>
												{/if}
											</div>
										</div>
									</div>

									<div class="grid grid-cols-3 gap-3 text-right sm:min-w-64 lg:min-w-60">
										<div>
											<div
												class="truncate text-sm font-black {isOnline
													? 'text-brand-primary'
													: isUnknown
														? 'text-brand-light/45'
														: 'text-brand-accent'}"
												title={currentStatus}
											>
												{isOnline ? 'Online' : isUnknown ? 'Unknown' : 'Down'}
											</div>
											<div
												class="text-[10px] font-bold tracking-widest text-brand-light/25 uppercase"
											>
												status
											</div>
										</div>
										<div>
											<div
												class="text-sm font-black {uptime >= 99
													? 'text-brand-primary'
													: uptime >= 95
														? 'text-brand-soft'
														: 'text-brand-accent'}"
											>
												{uptime.toFixed(1)}%
											</div>
											<div
												class="text-[10px] font-bold tracking-widest text-brand-light/25 uppercase"
											>
												uptime
											</div>
										</div>
										<div>
											<div class="text-sm font-black text-[#F0AD4E]">
												{currentLatency || avgLatency}ms
											</div>
											<div
												class="text-[10px] font-bold tracking-widest text-brand-light/25 uppercase"
											>
												latency
											</div>
										</div>
									</div>
								</div>
							</div>

							<div class="order-3 grid grid-cols-[1fr_2.75rem_2.75rem] gap-2 lg:min-w-56">
								<a
									href={resolve('/dashboard/[id]', { id: String(s.id) })}
									class="inline-flex h-11 items-center justify-center rounded-xl bg-brand-light/[0.06] px-4 text-sm font-bold transition hover:bg-brand-light/10"
								>
									Details
								</a>
								<button
									onclick={() => openEdit(s)}
									class="inline-flex h-11 items-center justify-center rounded-xl bg-brand-light/[0.06] text-brand-light/45 transition hover:bg-brand-light/10 hover:text-brand-light"
									title="Edit monitor"
								>
									<Settings class="h-4 w-4" />
								</button>
								<button
									onclick={() => deleteServer(s.id)}
									class="inline-flex h-11 items-center justify-center rounded-xl bg-brand-accent/10 text-brand-accent/60 transition hover:bg-brand-accent/20 hover:text-brand-accent"
									title="Delete monitor"
								>
									<Trash2 class="h-4 w-4" />
								</button>
							</div>
						</div>
					</article>
				{/each}
			</section>
		{/if}
	</div>
</div>

<style>
	/* Custom animations for Svelte 5 */
	@keyframes fade-in {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}
	@keyframes slide-in-from-top-4 {
		from {
			transform: translateY(-1rem);
		}
		to {
			transform: translateY(0);
		}
	}
	@keyframes zoom-in-95 {
		from {
			transform: scale(0.95);
			opacity: 0;
		}
		to {
			transform: scale(1);
			opacity: 1;
		}
	}

	.animate-in {
		animation-duration: 300ms;
		animation-fill-mode: both;
	}
	.fade-in {
		animation-name: fade-in;
	}
	.slide-in-from-top-4 {
		animation-name: slide-in-from-top-4;
	}
	.zoom-in-95 {
		animation-name: zoom-in-95;
	}

	@media (max-width: 639px) {
		.mobile-form-panel {
			background-color: #182825;
			backdrop-filter: none;
		}
	}
</style>
