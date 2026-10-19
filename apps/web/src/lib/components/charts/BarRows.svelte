<script lang="ts">
  import { scaleLinear } from 'd3-scale';

  import { GAP, type Series, crisp, rightRounded, stack } from './geometry.ts';

  type Row = {
    label: string;
    /** A value per series, stacked left to right. */
    values: readonly number[];
    /** The figure written at the bar's end. */
    value: string;
  };

  type Props = {
    title: string;
    series: readonly Series[];
    rows: readonly Row[];
    fallbackWidth?: number;
  };

  let { title, series, rows, fallbackWidth = 580 }: Props = $props();

  let width = $state(0);
  const w = $derived(width > 0 ? width : fallbackWidth);

  const BAR = 20;
  const VALUE = 112;

  const layout = $derived.by(() => {
    // Narrow, each label goes above its bar, so the bars keep their length.
    const stacked = w < 420;
    const rowH = stacked ? 52 : 40;
    const labelW = stacked ? 0 : 188;
    const plotW = Math.max(1, w - labelW - VALUE);
    const max = Math.max(1, ...rows.map((r) => r.values.reduce((a, b) => a + b, 0)));
    const x = scaleLinear().domain([0, max]).range([0, plotW]);
    return {
      stacked,
      labelW,
      height: rows.length * rowH,
      rows: rows.map((row, i) => {
        const top = i * rowH;
        const barY = stacked ? top + 22 : top + (rowH - BAR) / 2;
        // Each segment after the first gives up 2px to the gap before it.
        const segments = stack(row.values, (v) => labelW + x(v))
          .map((s, j) => ({
            series: series[s.index],
            x: s.from + (j ? GAP : 0),
            width: s.to - s.from - (j ? GAP : 0),
          }))
          .filter((s) => s.width >= 1);
        const end = segments.at(-1);
        return {
          ...row,
          labelY: stacked ? top + 12 : top + rowH / 2,
          barY,
          segments,
          end: end ? end.x + end.width : labelW,
        };
      }),
    };
  });
</script>

<div bind:clientWidth={width}>
  <svg viewBox="0 0 {w} {layout.height}" role="img" aria-label={title} class="block h-auto w-full">
    {#each layout.rows as row (row.label)}
      <text x="0" y={row.labelY} dy="0.32em" class="fill-text-primary-default ts-primary-body-m"
        >{row.label}</text
      >
      <!-- Every row starts on the same baseline, drawn under the bars. -->
      <line
        x1={crisp(layout.labelW) - 1}
        x2={crisp(layout.labelW) - 1}
        y1={row.barY - 4}
        y2={row.barY + BAR + 4}
        class="stroke-chart-axis"
      />
      {#each row.segments as segment, j (j)}
        {#if j === row.segments.length - 1}
          <path
            d={rightRounded(segment.x, row.barY, segment.width, BAR)}
            class={segment.series?.fill}
          />
        {:else}
          <rect
            x={segment.x}
            y={row.barY}
            width={segment.width}
            height={BAR}
            class={segment.series?.fill}
          />
        {/if}
      {/each}
      <text
        x={row.end + 10}
        y={row.barY + BAR / 2}
        dy="0.32em"
        class="fill-text-primary-default ts-primary-subheading-s tabular-nums">{row.value}</text
      >
    {/each}
  </svg>
</div>
