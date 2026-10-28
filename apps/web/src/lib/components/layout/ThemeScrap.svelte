<script lang="ts">
  import type { Theme } from '#lib/theme.ts';

  type Props = { theme: Theme };

  let { theme }: Props = $props();

  type Scheme = 'light' | 'dark';

  // Each theme in its own colours, whatever the page wears, so these name the
  // primitives the roles take in app.css, not the roles. The classes are
  // written out whole, so Tailwind finds them.
  const SCHEMES: Record<Scheme, Record<'rail' | 'canvas' | 'card' | 'wind' | 'solar', string>> = {
    light: {
      rail: 'fill-cool-grey-500',
      canvas: 'fill-cool-grey-25',
      card: 'fill-white stroke-grey-100',
      wind: 'fill-blue-200',
      solar: 'fill-amber-300',
    },
    dark: {
      rail: 'fill-cool-grey-600',
      canvas: 'fill-cool-grey-500',
      card: 'fill-cool-grey-490 stroke-cool-grey-400',
      wind: 'fill-blue-150',
      solar: 'fill-amber-200',
    },
  };

  /** Four half hours, wind under solar, in the scrap's own units. */
  const COLUMNS = [
    [3, 3],
    [5, 6],
    [6, 5],
    [4, 2],
  ] as const;

  const halves = $derived<[Scheme, number, number][]>(
    theme === 'system'
      ? [
          ['light', 0, 18],
          ['dark', 18, 18],
        ]
      : [[theme, 0, 36]],
  );
</script>

<!--
  A scrap of the app in a theme: the rail, the canvas, and a card with a chart
  of wind and solar. System is half of each, as the device may give either.
-->
<svg
  viewBox="0 0 36 24"
  aria-hidden="true"
  class="block h-600 w-[2.25rem] shrink-0 rounded-100 outline outline-stroke-neutral-default"
>
  {#each halves as [scheme, x, width] (scheme)}
    {@const s = SCHEMES[scheme]}
    <svg {x} {width} height="24" viewBox="{x} 0 {width} 24">
      <rect width="36" height="24" class={s.canvas} />
      <rect width="5" height="24" class={s.rail} />
      <rect x="9" y="4.5" width="24" height="16" rx="1.5" class={s.card} />
      {#each COLUMNS as [wind, solar], i (i)}
        <rect x={13 + i * 5} y={19 - wind} width="3" height={wind} class={s.wind} />
        <rect
          x={13 + i * 5}
          y={19 - wind - 0.75 - solar}
          width="3"
          height={solar}
          class={s.solar}
        />
      {/each}
    </svg>
  {/each}
</svg>
