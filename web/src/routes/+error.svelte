<script lang="ts">
	import { ArrowRight, Home, LayoutDashboard, SearchX } from 'lucide-svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
</script>

<svelte:head>
	<title>{page.status === 404 ? 'Page not found' : 'Something went wrong'} - isitdead</title>
	<meta
		name="description"
		content="The requested isitdead page could not be found. Return home or open your monitoring dashboard."
	/>
	{#if page.status === 404}
		<meta name="robots" content="noindex" />
	{/if}
</svelte:head>

<section class="relative isolate overflow-hidden px-4 py-14 sm:px-6 sm:py-20 lg:py-28">
	<div class="pointer-events-none absolute inset-0 -z-10">
		<div
			class="ambient-glow absolute top-[-10rem] left-1/2 h-[34rem] w-[34rem] -translate-x-1/2 rounded-full bg-brand-primary/16 blur-[140px]"
		></div>
		<div
			class="absolute inset-0 bg-[linear-gradient(to_right,rgba(222,244,198,0.04)_1px,transparent_1px),linear-gradient(to_bottom,rgba(222,244,198,0.04)_1px,transparent_1px)] [mask-image:linear-gradient(to_bottom,black,transparent_82%)] bg-[size:72px_72px]"
		></div>
	</div>

	<div class="container mx-auto grid max-w-6xl gap-12 lg:grid-cols-[0.9fr_1.1fr] lg:items-center">
		<div>
			<div
				class="mb-8 inline-flex items-center gap-3 rounded-full border border-brand-accent/20 bg-brand-accent/10 px-4 py-2 text-[11px] font-black tracking-[0.28em] text-brand-accent uppercase"
			>
				<SearchX class="h-4 w-4" />
				{page.status}
			</div>

			<h1 class="max-w-4xl text-4xl leading-[0.92] font-black text-brand-light sm:text-7xl">
				{page.status === 404 ? 'This page is not being monitored.' : 'This check failed.'}
			</h1>

			<p class="mt-7 max-w-2xl text-lg leading-8 font-medium text-brand-light/55">
				{page.status === 404
					? 'The route does not exist, moved, or was never published. You can go back to the product or open your monitors.'
					: page.error?.message || 'The app hit an unexpected error while loading this page.'}
			</p>

			<div class="mt-10 flex flex-col gap-4 sm:flex-row">
				<a
					href={resolve('/')}
					class="group inline-flex items-center justify-center rounded-2xl bg-brand-primary px-8 py-4 text-base font-black text-brand-dark shadow-2xl shadow-brand-primary/20 transition hover:-translate-y-0.5 hover:bg-brand-primary/90 active:translate-y-0"
				>
					<Home class="mr-2 h-5 w-5" />
					Go home
				</a>
				<a
					href={resolve('/dashboard')}
					class="inline-flex items-center justify-center rounded-2xl border border-brand-light/10 bg-brand-light/[0.03] px-8 py-4 text-base font-black text-brand-light/75 backdrop-blur transition hover:border-brand-primary/30 hover:text-brand-light"
				>
					<LayoutDashboard class="mr-2 h-5 w-5" />
					Open dashboard
				</a>
			</div>
		</div>

		<div class="relative">
			<div
				class="ambient-glow absolute -inset-4 rounded-[3.25rem] bg-gradient-to-br from-brand-accent/18 via-brand-primary/12 to-transparent blur-2xl"
			></div>
			<div class="glass-panel relative overflow-hidden rounded-[2.75rem] p-3">
				<div class="rounded-[2.25rem] border border-brand-light/10 bg-brand-dark/80 p-6 sm:p-8">
					<div class="mb-8 flex items-center justify-between">
						<div class="flex items-center gap-2">
							<span class="h-3 w-3 rounded-full bg-brand-accent"></span>
							<span class="h-3 w-3 rounded-full bg-[#E5B181]"></span>
							<span class="h-3 w-3 rounded-full bg-brand-primary"></span>
						</div>
						<span
							class="rounded-full bg-brand-accent/10 px-3 py-1 text-[10px] font-black tracking-widest text-brand-accent uppercase"
						>
							Not found
						</span>
					</div>

					<div class="grid gap-3">
						<div class="rounded-3xl border border-brand-light/10 bg-brand-light/[0.03] p-5">
							<div class="text-xs font-black tracking-[0.26em] text-brand-light/30 uppercase">
								Requested path
							</div>
							<div class="mt-3 text-2xl font-black break-all text-brand-light">
								{page.url.pathname}
							</div>
						</div>
						<div class="rounded-3xl border border-brand-primary/20 bg-brand-primary/10 p-5">
							<div class="flex items-center justify-between gap-4">
								<div>
									<div class="text-xs font-black tracking-[0.26em] text-brand-primary uppercase">
										Next step
									</div>
									<div class="mt-2 text-lg font-black">Create or inspect a monitor</div>
								</div>
								<ArrowRight class="h-6 w-6 text-brand-primary" />
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</section>
