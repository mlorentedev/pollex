// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://mlorentedev.github.io',
	base: '/pollex',
	integrations: [
		starlight({
			title: 'Pollex',
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/mlorentedev/pollex' }],
			sidebar: [
				{ label: 'Getting Started', slug: 'getting-started' },
			],
		}),
	],
});
