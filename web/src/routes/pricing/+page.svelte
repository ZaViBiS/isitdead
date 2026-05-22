<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { ArrowRight, CheckCircle2, Gauge, ShieldCheck, WalletCards } from 'lucide-svelte';
	import type { BillingPlan, User } from '$lib/utils';

	let plans = $state<BillingPlan[]>([]);
	let user = $state<User | null>(null);
	let isLoading = $state(true);
	let busyPlan = $state('');
	let error = $state('');

	onMount(() => {
		void loadPricing();
	});

	async function loadPricing() {
		isLoading = true;
		error = '';
		try {
			const planRes = await fetch('/api/billing/plans');
			if (planRes.ok) plans = (await planRes.json()) as BillingPlan[];

			const token = localStorage.getItem('token');
			if (token) {
				const meRes = await fetch('/api/me', {
					headers: { Authorization: `Bearer ${token}` }
				});
				if (meRes.ok) user = (await meRes.json()) as User;
			}
		} catch {
			error = 'Could not load pricing.';
		} finally {
			isLoading = false;
		}
	}

	function currentPlanID() {
		return user?.plan || 'free';
	}

	function planFeatures(plan: BillingPlan) {
		return [
			`${plan.monitor_limit} monitors`,
			'Fast checks included',
			'Full status history included',
			'Public status pages included',
			'Email, Telegram, and SSL alerts included'
		];
	}

	function redirectToStripe(url: string) {
		const target = new URL(url);
		if (target.protocol !== 'https:' || !target.hostname.endsWith('.stripe.com')) {
			throw new Error('Invalid billing redirect URL');
		}
		window.location.href = target.toString();
	}

	async function startCheckout(plan: BillingPlan) {
		error = '';
		const token = localStorage.getItem('token');
		if (!token) {
			window.location.href = resolve('/register');
			return;
		}
		if (plan.id === currentPlanID()) return;

		busyPlan = plan.id;
		try {
			const res = await fetch('/api/billing/checkout', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
				body: JSON.stringify({ plan: plan.id })
			});
			const data = await res.json();
			if (!res.ok) throw new Error(data.error ?? 'Could not start checkout');
			redirectToStripe(data.url);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Could not start checkout';
		} finally {
			busyPlan = '';
		}
	}

	async function openPortal() {
		error = '';
		const token = localStorage.getItem('token');
		if (!token) return;

		busyPlan = 'portal';
		try {
			const res = await fetch('/api/billing/portal', {
				method: 'POST',
				headers: { Authorization: `Bearer ${token}` }
			});
			const data = await res.json();
			if (!res.ok) throw new Error(data.error ?? 'Could not open billing portal');
			redirectToStripe(data.url);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Could not open billing portal';
		} finally {
			busyPlan = '';
		}
	}
</script>

<svelte:head>
	<title>Pricing - isitdead</title>
	<meta
		name="description"
		content="Choose an isitdead plan for uptime monitors, faster checks, public status pages, and alerting."
	/>
</svelte:head>

<section class="px-4 py-12 sm:px-6 sm:py-16">
	<div class="container mx-auto max-w-7xl">
		<div class="mb-10 flex flex-col justify-between gap-6 lg:flex-row lg:items-end">
			<div class="max-w-3xl">
				<div class="signal-pill mb-6">Pricing</div>
				<h1 class="text-4xl leading-none font-black tracking-tight sm:text-6xl">
					Pay only when you need more monitors.
				</h1>
				<p class="mt-5 max-w-2xl text-lg leading-8 text-brand-light/55">
					Every plan includes the same monitoring features. The only limit that changes is how many
					monitors you can run.
				</p>
			</div>
			{#if user && user.plan !== 'free'}
				<button
					onclick={openPortal}
					class="inline-flex items-center justify-center gap-2 rounded-2xl border border-brand-primary/25 bg-brand-primary/10 px-5 py-3 font-black text-brand-primary transition hover:bg-brand-primary/15"
				>
					<WalletCards class="h-5 w-5" />
					{busyPlan === 'portal' ? 'Opening...' : 'Manage billing'}
				</button>
			{/if}
		</div>

		{#if error}
			<div
				class="mb-6 rounded-2xl border border-brand-accent/20 bg-brand-accent/10 px-5 py-4 text-sm font-bold text-brand-accent"
			>
				{error}
			</div>
		{/if}

		{#if isLoading}
			<div class="glass-panel rounded-[2rem] p-8 text-brand-light/45">Loading pricing...</div>
		{:else}
			<div class="grid gap-4 lg:grid-cols-3">
				{#each plans as plan (plan.id)}
					{@const selected = currentPlanID() === plan.id}
					<article
						class="glass-panel flex min-h-[34rem] flex-col rounded-[2rem] p-6 transition {plan.id ===
						'pro'
							? 'border-brand-primary/35 bg-brand-primary/[0.075]'
							: ''}"
					>
						<div class="mb-6 flex items-start justify-between gap-4">
							<div>
								<div class="flex items-center gap-2">
									<h2 class="text-2xl font-black">{plan.name}</h2>
									{#if selected}
										<span
											class="rounded-full bg-brand-primary px-2.5 py-1 text-[10px] font-black text-brand-dark uppercase"
										>
											Current
										</span>
									{/if}
								</div>
								<p class="mt-2 text-sm leading-6 text-brand-light/45">{plan.description}</p>
							</div>
							<div class="rounded-2xl bg-brand-light/5 p-3 text-brand-primary">
								{#if plan.id === 'free'}
									<ShieldCheck class="h-6 w-6" />
								{:else}
									<Gauge class="h-6 w-6" />
								{/if}
							</div>
						</div>

						<div class="mb-6">
							<div class="text-5xl font-black tracking-tight">{plan.price}</div>
							<div class="mt-2 text-xs font-bold tracking-widest text-brand-light/30 uppercase">
								{plan.id === 'free' ? 'No card required' : 'More monitor slots'}
							</div>
						</div>

						<div class="space-y-3">
							{#each planFeatures(plan) as feature (feature)}
								<div class="flex items-center gap-3 text-sm font-semibold text-brand-light/65">
									<CheckCircle2 class="h-4 w-4 text-brand-primary" />
									<span>{feature}</span>
								</div>
							{/each}
						</div>

						<div class="mt-auto pt-8">
							{#if plan.id === 'free'}
								<a
									href={user ? resolve('/dashboard') : resolve('/register')}
									class="inline-flex w-full items-center justify-center gap-2 rounded-2xl border border-brand-light/10 bg-brand-light/[0.04] px-5 py-4 font-black transition hover:border-brand-primary/30"
								>
									{user ? 'Open dashboard' : 'Start free'}
									<ArrowRight class="h-5 w-5" />
								</a>
							{:else}
								<button
									onclick={() => startCheckout(plan)}
									disabled={selected || !plan.stripe_available || busyPlan !== ''}
									class="inline-flex w-full items-center justify-center gap-2 rounded-2xl bg-brand-primary px-5 py-4 font-black text-brand-dark transition hover:bg-brand-primary/90 disabled:cursor-not-allowed disabled:opacity-45"
								>
									{#if !plan.stripe_available}
										Unavailable
									{:else if selected}
										Current plan
									{:else if busyPlan === plan.id}
										Opening checkout...
									{:else}
										Upgrade to {plan.name}
									{/if}
									<ArrowRight class="h-5 w-5" />
								</button>
							{/if}
						</div>
					</article>
				{/each}
			</div>
		{/if}
	</div>
</section>
