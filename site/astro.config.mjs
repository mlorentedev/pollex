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
			favicon: '/favicon.svg',
			logo: {
				src: './public/favicon.svg',
			},
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/mlorentedev/pollex' }],
			customCss: ['./src/styles/custom.css'],
			sidebar: [
				{ label: 'Getting Started', slug: 'getting-started' },
			],
		}),
	],
});
