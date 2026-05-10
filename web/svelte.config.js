import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({
			fallback: 'index.html',
			pages: 'dist',
			assets: 'dist'
		}),
		prerender: {
			handleHttpError: ({ path, referrer, message }) => {
				if (path.startsWith('/api/')) {
					return;
				}
				if (path === '/login' || path === '/features' || path === '/pricing') {
					return;
				}
				throw new Error(message);
			},
			handleMissingId: 'ignore'
		}
	}
};

export default config;
