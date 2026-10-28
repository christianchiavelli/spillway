<script lang="ts">
  import { scaleQuantize } from 'd3-scale';

  import type { ResolvedPathname } from '$app/types';
  import { m } from '#lib/paraglide/messages.js';

  import { CHAR } from './geometry.ts';

  type Region = {
    key: string;
    label: string;
    value: number;
    /** The outline, in the map's own units. */
    d: string;
    /** Where the label goes, in the map's own units. */
    labelAt: readonly [number, number];
    /** Where a click on the region leads. */
    href: ResolvedPathname;
  };

  type Props = {
    /** The map's size in its own units. */
    width: number;
    height: number;
    regions: readonly Region[];
    /** The region pointed at here or in the figures beside the map. */
    active?: string | null;
    onactive?: (key: string | null) => void;
    /** The width to lay out for until the page is measured, as on the server. */
    fallbackWidth?: number;
  };

  let { width, height, regions, active = null, onactive, fallbackWidth = 260 }: Props = $props();

  const LINE = 14;
  const WRAP = 14 * CHAR;
  // The classes are written out whole, so Tailwind finds them.
  const STEPS = [
    { fill: 'fill-chart-scale-1', swatch: 'bg-chart-scale-1' },
    { fill: 'fill-chart-scale-2', swatch: 'bg-chart-scale-2' },
    { fill: 'fill-chart-scale-3', swatch: 'bg-chart-scale-3' },
    { fill: 'fill-chart-scale-4', swatch: 'bg-chart-scale-4' },
  ];

  let measured = $state(0);

  // Drawn in pixels, so its labels keep their size at any width: the outlines
  // are scaled into them, and their strokes are not.
  const w = $derived(measured > 0 ? measured : fallbackWidth);
  const k = $derived(w / width);

  /** A label's lines: words joined while they fit the width of 14 digits. */
  function lines(label: string): string[] {
    const out: string[] = [];
    for (const word of label.split(' ')) {
      const last = out.at(-1);
      if (last !== undefined && (last.length + 1 + word.length) * CHAR <= WRAP) {
        out[out.length - 1] = `${last} ${word}`;
      } else out.push(word);
    }
    return out;
  }

  const drawn = $derived.by(() => {
    // Each step a quarter of the largest value: the shades rank the regions,
    // and the figures beside the map say by how much.
    const step = scaleQuantize<(typeof STEPS)[number]>()
      .domain([0, Math.max(1, ...regions.map((r) => r.value))])
      .range(STEPS);
    return regions.map((r) => {
      const text = lines(r.label);
      return {
        ...r,
        fill: step(r.value).fill,
        x: r.labelAt[0] * k,
        y: r.labelAt[1] * k - ((text.length - 1) * LINE) / 2,
        lines: text,
      };
    });
  });
</script>

<!--
  Hidden from screen readers, and out of the tab order: it shows where, and
  the figures beside it, which link to the same places, say it all in words.
-->
<div bind:clientWidth={measured} aria-hidden="true">
  <svg viewBox="0 0 {w} {height * k}" class="block h-auto w-full">
    <g transform="scale({k})">
      {#each drawn as region (region.key)}
        <a
          href={region.href}
          aria-label={region.label}
          tabindex="-1"
          onpointerenter={() => onactive?.(region.key)}
          onpointerleave={() => onactive?.(null)}
        >
          <path
            d={region.d}
            vector-effect="non-scaling-stroke"
            stroke-width="1.5"
            stroke-linejoin="round"
            class={[
              region.fill,
              'stroke-surface-container-primary transition-opacity',
              active !== null && active !== region.key && 'opacity-40',
            ]}
          />
        </a>
      {/each}
      <!-- Outlined as well as left unfaded, since the lightest step barely stands out alone. -->
      {#each drawn.filter((r) => r.key === active) as region (region.key)}
        <path
          d={region.d}
          vector-effect="non-scaling-stroke"
          stroke-width="2"
          stroke-linejoin="round"
          class="pointer-events-none fill-none stroke-text-primary-default"
        />
      {/each}
    </g>
    <!--
      A label can run over a border or off the coast, so each one is the text's
      colour in a halo of the card's, which reads on every step and on the card
      around the map, in either theme.
    -->
    {#each drawn as region (region.key)}
      <text
        x={region.x}
        y={region.y}
        text-anchor="middle"
        stroke-width="3"
        stroke-linejoin="round"
        paint-order="stroke"
        class="pointer-events-none fill-text-primary-default stroke-surface-container-primary ts-primary-subheading-xs"
      >
        {#each region.lines as line, i (i)}
          <tspan x={region.x} dy={i === 0 ? '0.32em' : LINE}>{line}</tspan>
        {/each}
      </text>
    {/each}
  </svg>
  <!-- The shades only rank the regions, so the key says which way, not by how much. -->
  <p
    class="mt-400 flex items-center justify-center gap-300 ts-primary-body-s text-text-secondary-default"
  >
    {m.map_less()}
    <span class="flex">
      {#each STEPS as s (s.fill)}
        <span class={['h-[0.625rem] w-600', s.swatch]}></span>
      {/each}
    </span>
    {m.map_more()}
  </p>
</div>
