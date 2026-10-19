<script lang="ts">
  import { scaleLinear } from 'd3-scale';
  import { line } from 'd3-shape';

  import { CHAR, type Series, crisp, fitLabels } from './geometry.ts';

  type Props = {
    title: string;
    series: readonly Series[];
    /** A list of values per series, one a point; null where there is none. */
    values: readonly (readonly (number | null)[])[];
    /** The top of the scale, rounded up to a tick. */
    max: number;
    tick: (value: number) => string;
    xLabel: (i: number) => string | null;
    /** The label at a line's last point. */
    endLabel: (series: number, value: number) => string;
    fallbackWidth?: number;
    class: string;
  };

  let {
    title,
    series,
    values,
    max,
    tick,
    xLabel,
    endLabel,
    fallbackWidth = 1234,
    class: className,
  }: Props = $props();

  const TOP = 16;
  const BOTTOM = 28;
  const RIGHT = 92;

  let width = $state(0);
  let height = $state(0);
  const w = $derived(width > 0 ? width : fallbackWidth);
  const h = $derived(height > 0 ? height : 260);

  const layout = $derived.by(() => {
    const y = scaleLinear().domain([0, max]).nice(4);
    const ticks = y.ticks(4);
    const labels = ticks.map(tick);
    const left = Math.max(...labels.map((l) => l.length)) * CHAR + 14;
    const plotW = Math.max(1, w - left - RIGHT);
    const plotH = Math.max(1, h - TOP - BOTTOM);
    y.range([TOP + plotH, TOP]);
    const count = Math.max(...values.map((v) => v.length));
    const step = plotW / Math.max(1, count - 1);
    const xAt = (i: number) => left + i * step;
    const path = line<{ v: number | null; i: number }>()
      .defined((d) => d.v !== null)
      .x((d) => xAt(d.i))
      .y((d) => y(d.v ?? 0));

    const candidates = Array.from({ length: count }, (_, i) => ({
      i,
      x: xAt(i),
      text: xLabel(i),
    })).flatMap((c) => (c.text === null ? [] : [{ ...c, text: c.text }]));
    const fit = fitLabels(candidates, w);

    return {
      left,
      ticks: ticks.map((t, i) => ({ value: t, label: labels[i], at: y(t) })),
      base: y(0),
      xTicks: candidates.filter((_, j) => fit.shown[j]),
      lines: values.map((v, s) => {
        const points = v.map((value, i) => ({ v: value, i }));
        const last = points.findLast((p) => p.v !== null);
        return {
          d: path(points) ?? '',
          end:
            last && last.v !== null
              ? { x: xAt(last.i), y: y(last.v), text: endLabel(s, last.v) }
              : null,
        };
      }),
    };
  });
</script>

<div class={['relative', className]} bind:clientWidth={width} bind:clientHeight={height}>
  <svg
    viewBox="0 0 {w} {h}"
    role="img"
    aria-label={title}
    class="block h-full w-full overflow-visible"
  >
    {#each layout.ticks as t (t.value)}
      {#if t.value > 0}
        <line
          x1={layout.left}
          x2={w - RIGHT}
          y1={crisp(t.at)}
          y2={crisp(t.at)}
          class="stroke-chart-grid"
        />
      {/if}
      <text
        x={layout.left - 10}
        y={t.at}
        dy="0.32em"
        text-anchor="end"
        class="fill-text-secondary-default ts-primary-body-s tabular-nums">{t.label}</text
      >
    {/each}
    <line
      x1={layout.left}
      x2={w - RIGHT}
      y1={crisp(layout.base)}
      y2={crisp(layout.base)}
      class="stroke-chart-axis"
    />
    {#each layout.xTicks as t (t.i)}
      <text
        x={t.x}
        y={h - 8}
        text-anchor="middle"
        class="fill-text-secondary-default ts-primary-body-s">{t.text}</text
      >
    {/each}
    {#each layout.lines as l, s (s)}
      <path
        d={l.d}
        fill="none"
        stroke-width="2"
        stroke-linejoin="round"
        stroke-linecap="round"
        class={series[s]?.stroke}
      />
      {#if l.end}
        <circle
          cx={l.end.x}
          cy={l.end.y}
          r="4"
          stroke-width="2"
          class={[series[s]?.fill, 'stroke-surface-container-primary']}
        />
        <text
          x={l.end.x + 12}
          y={l.end.y}
          dy="0.32em"
          class="fill-text-primary-default ts-primary-subheading-xs">{l.end.text}</text
        >
      {/if}
    {/each}
  </svg>
</div>
