<script lang="ts">
	let { class: className = 'h-8 w-8', title = 'isitdead' } = $props<{
		class?: string;
		title?: string;
	}>();
</script>

<svg
	class={className}
	viewBox="0 0 64 64"
	fill="none"
	xmlns="http://www.w3.org/2000/svg"
	role="img"
	aria-label={title}
>
	<title>{title}</title>
	<rect width="64" height="64" rx="17" fill="#12211D" />
	<rect x="13" y="22" width="38" height="20" rx="10" fill="#DEF4C6" fill-opacity="0.14" />
	<rect
		x="13"
		y="22"
		width="38"
		height="20"
		rx="10"
		stroke="#DEF4C6"
		stroke-opacity="0.22"
		stroke-width="4"
	/>
	<circle cx="40" cy="32" r="9" fill="#73E2A7" />
	<path d="M20 32H27" stroke="#D62246" stroke-width="5" stroke-linecap="round" />
</svg>
