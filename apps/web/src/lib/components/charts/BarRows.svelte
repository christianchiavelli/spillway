<script lang="ts">
  import { scaleLinear } from 'd3-scale';

  import type { ResolvedPathname } from '$app/types';

  import { GAP, type Series, rightRounded, stack } from './geometry.ts';

  type Row = {
    key: string;
    label: string;
    /** A line under the label, such as what it means. */
    description?: string | undefined;
    /** A value per series, stacked left to right. */
    values: readonly number[];
    /** The figure on the right. */
    value: string;
    /** Where the row leads: the whole row is the link. */
    href?: ResolvedPathname | undefined;
  };

  type Props = {
    /** The list's name, for a screen reader. */
    title: string;
    series: readonly Series[];
    /** In the order they are ranked in. */
    rows: readonly Row[];
    /** The row pointed at here or in a chart beside the list. */
    active?: string | null;
    onactive?: (key: string | null) => void;
    /** The width to lay out for until the page is measured, as on the server. */
    fallbackWidth?: number;
  };

  let { title, series, rows, active = null, onactive, fallbackWidth = 360 }: Props = $props();
  const id = $props.id();

  const BAR = 12;
  /** The row's padding on each side, as its p-300 class sets it. */
  const PAD = 8;

  let width = $state(0);
  const plotW = $derived(Math.max(1, (width > 0 ? width : fallbackWidth) - 2 * PAD));

  const bars = $derived.by(() => {
    // One scale for every row, so the bars compare: the longest is the largest.
    const max = Math.max(1, ...rows.map((r) => r.values.reduce((a, b) => a + b, 0)));
    const x = scaleLinear().domain([0, max]).range([0, plotW]);
    return rows.map((row) =>
      // Each segment after the first gives up 2px to the gap before it.
      stack(row.values, x)
        .map((s, j) => ({
          series: series[s.index],
          x: s.from + (j ? GAP : 0),
          width: s.to - s.from - (j ? GAP : 0),
        }))
        .filter((s) => s.width >= 1),
    );
  });
</script>

<!--
  A ranked list, with a bar a row on one scale: words a screen reader reads as
  they are, the drawing under each row hidden from it. The list reaches past
  the card's text by a row's padding, so the rows' words line up with it and
  their highlight has room around them.
-->
<ol aria-label={title} bind:clientWidth={width} class="-mx-300 flex flex-col">
  {#each rows as row, i (row.key)}
    <li
      class={[
        'group relative flex flex-col gap-300 rounded-200 p-300 transition-colors',
        row.key === active && 'bg-surface-container-hover',
      ]}
      onpointerenter={() => onactive?.(row.key)}
      onpointerleave={() => onactive?.(null)}
    >
      <div class="flex items-start justify-between gap-500">
        <div class="flex min-w-0 flex-col gap-100">
          <!-- Broken only between words: a name like Centre-West stays whole. -->
          <span
            id={`${id}-${i}`}
            class={[
              'ts-primary-subheading-s text-text-primary-default',
              row.href && 'group-hover:text-text-primary-link group-hover:underline',
            ]}
          >
            {#each row.label.split(' ') as word, j (j)}{j > 0 ? ' ' : ''}<span
                class="whitespace-nowrap">{word}</span
              >{/each}
          </span>
          {#if row.description}
            <span class="ts-primary-body-s text-text-secondary-default">{row.description}</span>
          {/if}
        </div>
        <span class="shrink-0 ts-primary-subheading-s text-text-primary-default tabular-nums"
          >{row.value}</span
        >
      </div>
      <svg viewBox="0 0 {plotW} {BAR}" aria-hidden="true" class="block h-auto w-full">
        <!-- The track: the whole scale, so a short bar reads as short of the longest. -->
        <path d={rightRounded(0, 0, plotW, BAR)} class="fill-chart-grid" />
        {#each bars[i] ?? [] as segment, j (j)}
          {#if j === (bars[i]?.length ?? 0) - 1}
            <path d={rightRounded(segment.x, 0, segment.width, BAR)} class={segment.series?.fill} />
          {:else}
            <rect x={segment.x} width={segment.width} height={BAR} class={segment.series?.fill} />
          {/if}
        {/each}
      </svg>
      {#if row.href}
        <!--
          Over the whole row, so a click anywhere on it follows the link, and
          focus rings the row; named by the row's label.
        -->
        <a
          href={row.href}
          aria-labelledby={`${id}-${i}`}
          class="absolute inset-0 rounded-200"
          onfocus={() => onactive?.(row.key)}
          onblur={() => onactive?.(null)}
        ></a>
      {/if}
    </li>
  {/each}
</ol>
