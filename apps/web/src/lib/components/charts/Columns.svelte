<script lang="ts">
  import { scaleLinear } from 'd3-scale';

  import {
    CHAR,
    GAP,
    MAX_BAR,
    type Series,
    type Tip,
    crisp,
    fitLabels,
    stack,
    textWidth,
    topRounded,
  } from './geometry.ts';
  import Tooltip from './Tooltip.svelte';

  type Props = {
    /** The chart's name, all a screen reader is given: a table goes with it. */
    title: string;
    series: readonly Series[];
    /** A column each, a value per series; null for one not published yet. */
    values: readonly (readonly number[] | null)[];
    tick: (value: number) => string;
    /** The label under a column, or null for none. */
    xLabel: (i: number) => string | null;
    /** Shorter labels, for when the long ones do not fit. */
    xLabelShort?: ((i: number) => string | null) | undefined;
    /** The label over the tallest column. */
    peak?: ((i: number, total: number) => string) | undefined;
    /** A note over the columns not published yet. */
    emptyNote?: string | undefined;
    /** A dashed line across, such as a threshold. */
    rule?: { value: number; label: string; caption: string } | undefined;
    /** One fill for a whole column, in place of its series' fills. */
    fill?: ((i: number) => string) | undefined;
    tip?: ((i: number) => Tip | null) | undefined;
    /** Notes over columns, each on a leader as long as lift, left out under minWidth. */
    notes?: readonly { index: number; text: string; lift: number; minWidth: number }[] | undefined;
    /** The width to lay out for until the page is measured, as on the server. */
    fallbackWidth?: number;
    /** The frame's height, as classes: the chart fills it. */
    class: string;
  };

  let {
    title,
    series,
    values,
    tick,
    xLabel,
    xLabelShort,
    peak,
    emptyNote,
    rule,
    fill,
    tip,
    notes = [],
    fallbackWidth = 1234,
    class: className,
  }: Props = $props();

  const TOP = 26;
  const BOTTOM = 28;

  let width = $state(0);
  let height = $state(0);
  let hover = $state<number | null>(null);

  // The server cannot measure, so it lays out for the usual width and the
  // browser scales that to fit, until the frame is measured and laid out again.
  const w = $derived(width > 0 ? width : fallbackWidth);
  const h = $derived(height > 0 ? height : 300);

  const layout = $derived.by(() => {
    const totals = values.map((v) => (v ? v.reduce((a, b) => a + b, 0) : 0));
    const max = Math.max(1, ...totals, rule?.value ?? 0);
    const y = scaleLinear()
      .domain([0, max * 1.04])
      .nice(5);
    const ticks = y.ticks(5);
    const labels = ticks.map(tick);
    const left = Math.max(...labels.map((l) => l.length)) * CHAR + 14;
    const right = rule ? 96 : 4;
    const plotW = Math.max(1, w - left - right);
    const plotH = Math.max(1, h - TOP - BOTTOM);
    y.range([TOP + plotH, TOP]);
    const slotW = plotW / Math.max(1, values.length);
    const barW = Math.min(MAX_BAR, Math.max(2, Math.floor(slotW * 0.64)));
    const centre = (i: number) => left + slotW * i + slotW / 2;

    const columns = values.map((v, i) => {
      if (v === null) return null;
      const x = Math.round(centre(i) - barW / 2);
      // Each segment above the first gives up 2px to the gap below it.
      const segments = stack(v, y)
        .map((s, j) => ({
          series: series[s.index],
          top: s.to,
          height: s.from - (j ? GAP : 0) - s.to,
        }))
        .filter((s) => s.height >= 1);
      return { x, segments, top: segments.at(-1)?.top ?? y(0) };
    });

    const candidates = (label: (i: number) => string | null) =>
      values.flatMap((_, i) => {
        const text = label(i);
        return text === null ? [] : [{ i, x: centre(i), text }];
      });
    let ticksX = candidates(xLabel);
    let fit = fitLabels(ticksX, w);
    if (!fit.all && xLabelShort) {
      ticksX = candidates(xLabelShort);
      fit = fitLabels(ticksX, w);
    }

    let peakLabel: { x: number; y: number; text: string } | null = null;
    const highest = Math.max(...totals);
    const top = totals.indexOf(highest);
    const topColumn = columns[top];
    if (peak && topColumn && highest > 0) {
      const text = peak(top, highest);
      const half = textWidth(text) / 2;
      peakLabel = {
        text,
        x: Math.min(Math.max(centre(top), left + half), w - right - half),
        y: topColumn.top - 8,
      };
    }

    const firstEmpty = values.indexOf(null);
    const empties = values.filter((v) => v === null).length;
    const note =
      emptyNote && firstEmpty >= 0 && empties * slotW >= textWidth(emptyNote)
        ? { x: (left + slotW * firstEmpty + w - right) / 2, y: y(0) - 14 }
        : null;

    const annotations = notes.flatMap((n) => {
      const column = columns[n.index];
      if (!column || w < n.minWidth) return [];
      const half = textWidth(n.text) / 2;
      return [
        {
          x: centre(n.index),
          textX: Math.min(Math.max(centre(n.index), half), w - half),
          from: column.top - 6,
          to: column.top - n.lift + 8,
          y: column.top - n.lift,
          text: n.text,
        },
      ];
    });

    return {
      y,
      annotations,
      ticks: ticks.map((t, i) => ({ value: t, label: labels[i], at: y(t) })),
      left,
      right,
      plotW,
      plotH,
      slotW,
      barW,
      base: y(0),
      columns,
      xTicks: ticksX.filter((_, j) => fit.shown[j]),
      peakLabel,
      note,
      ruleY: rule ? crisp(y(rule.value)) : 0,
    };
  });

  const tipped = $derived.by(() => {
    if (hover === null || !tip) return null;
    const column = layout.columns[hover];
    const content = tip(hover);
    if (!column || !content) return null;
    return { column, content };
  });

  function point(event: PointerEvent) {
    const box = (event.currentTarget as SVGRectElement).getBoundingClientRect();
    const i = Math.floor(((event.clientX - box.left) / box.width) * values.length);
    hover = Math.min(values.length - 1, Math.max(0, i));
  }
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
          x2={w - layout.right}
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

    {#if tipped}
      <rect
        x={layout.left + layout.slotW * (hover ?? 0)}
        y={TOP}
        width={layout.slotW}
        height={layout.plotH}
        class="fill-surface-container-hover"
      />
    {/if}

    {#each layout.columns as column, i (i)}
      {#if column}
        {#each column.segments as segment, j (j)}
          {@const cls = fill?.(i) ?? segment.series?.fill}
          {#if j === column.segments.length - 1}
            <path d={topRounded(column.x, segment.top, layout.barW, segment.height)} class={cls} />
          {:else}
            <rect
              x={column.x}
              y={segment.top}
              width={layout.barW}
              height={segment.height}
              class={cls}
            />
          {/if}
        {/each}
      {/if}
    {/each}

    <line
      x1={layout.left}
      x2={w - layout.right}
      y1={crisp(layout.base)}
      y2={crisp(layout.base)}
      class="stroke-chart-axis"
    />

    {#each layout.xTicks as t (t.i)}
      <text
        x={t.x}
        y={h - 8}
        text-anchor="middle"
        class={[
          'ts-primary-body-s',
          values[t.i] === null ? 'fill-text-primary-disabled' : 'fill-text-secondary-default',
        ]}>{t.text}</text
      >
    {/each}

    {#if layout.peakLabel}
      <text
        x={layout.peakLabel.x}
        y={layout.peakLabel.y}
        text-anchor="middle"
        class="fill-text-primary-default ts-primary-subheading-xs">{layout.peakLabel.text}</text
      >
    {/if}

    {#each layout.annotations as a (a.text)}
      <line x1={a.x} x2={a.x} y1={a.from} y2={a.to} class="stroke-text-secondary-default" />
      <text
        x={a.textX}
        y={a.y}
        text-anchor="middle"
        class="fill-text-secondary-default ts-primary-body-s">{a.text}</text
      >
    {/each}

    {#if rule}
      <line
        x1={layout.left}
        x2={w - layout.right + 8}
        y1={layout.ruleY}
        y2={layout.ruleY}
        stroke-dasharray="4 3"
        class="stroke-text-primary-default"
      />
      <text
        x={w - layout.right + 14}
        y={layout.ruleY}
        dy="-0.3em"
        class="fill-text-primary-default ts-primary-subheading-xs">{rule.label}</text
      >
      <text
        x={w - layout.right + 14}
        y={layout.ruleY}
        dy="1.1em"
        class="fill-text-secondary-default ts-primary-body-s">{rule.caption}</text
      >
    {/if}

    {#if layout.note && emptyNote}
      <text
        x={layout.note.x}
        y={layout.note.y}
        text-anchor="middle"
        class="fill-text-secondary-default ts-primary-body-s">{emptyNote}</text
      >
    {/if}

    {#if tip}
      <!-- Where the pointer is read: the whole plot, a column at a time. -->
      <rect
        x={layout.left}
        y={TOP}
        width={layout.plotW}
        height={layout.plotH}
        fill="transparent"
        aria-hidden="true"
        onpointermove={point}
        onpointerdown={point}
        onpointerleave={() => (hover = null)}
      />
    {/if}

    {#if tipped}
      <Tooltip
        tip={tipped.content}
        anchorX={tipped.column.x}
        anchorWidth={layout.barW}
        anchorY={tipped.column.top}
        width={w}
        height={h}
      />
    {/if}
  </svg>
</div>
