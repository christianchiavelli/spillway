import type { Component } from 'svelte';
import type { SvelteHTMLElements } from 'svelte/elements';

/** An icon from `~icons/material-symbols/*`, which unplugin-icons compiles. */
export type Icon = Component<SvelteHTMLElements['svg']>;
