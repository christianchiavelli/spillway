<script lang="ts">
  import { m } from '#lib/paraglide/messages.js';

  import type { Tip } from './geometry.ts';

  type Props = {
    tip: Tip;
    /** The mark it points at: its left edge, width and top. */
    anchorX: number;
    anchorWidth: number;
    anchorY: number;
    /** The chart's size, which the tooltip stays inside. */
    width: number;
    height: number;
  };

  let { tip, anchorX, anchorWidth, anchorY, width, height }: Props = $props();

  const W = 200;
  const PAD = 16;
  const ROW = 20;

  const box = $derived.by(() => {
    const h = PAD + 12 + 14 + tip.rows.length * ROW + (tip.total ? ROW + 10 : 0) + 4;
    // Beside the mark, to its right where there is room and to its left where not.
    let x = anchorX + anchorWidth + 14;
    if (x + W > width) x = anchorX - 14 - W;
    x = Math.max(0, x);
    const y = Math.min(Math.max(0, anchorY - 12), Math.max(0, height - h));
    return { x, y, h };
  });
</script>

<!--
  Drawn in the chart's own SVG, so it needs no style attribute to sit where it
  points. Pointer only: the same figures are in the table under the chart.
-->
<g transform="translate({box.x} {box.y})" aria-hidden="true" class="pointer-events-none">
  <rect width={W} height={box.h} rx="4" class="fill-surface-primary-cool-dark" />
  <text x={PAD} y={PAD + 10} class="fill-text-on-dark-secondary ts-primary-subheading-xs"
    >{tip.title}</text
  >
  {#each tip.rows as row, i (row.label)}
    {@const y = PAD + 12 + 14 + i * ROW + ROW / 2}
    <rect x={PAD} y={y - 1} width="12" height="2" rx="1" class={row.fill} />
    <text x={PAD + 20} {y} dy="0.32em" class="fill-text-on-dark-secondary ts-primary-body-s"
      >{row.label}</text
    >
    <text
      x={W - PAD}
      {y}
      dy="0.32em"
      text-anchor="end"
      class="fill-text-on-dark-primary ts-primary-subheading-s tabular-nums">{row.value}</text
    >
  {/each}
  {#if tip.total}
    {@const y = PAD + 12 + 14 + tip.rows.length * ROW + 10}
    <line x1={PAD} x2={W - PAD} y1={y - 0.5} y2={y - 0.5} class="stroke-stroke-neutral-on-dark" />
    <text x={PAD} y={y + ROW / 2} dy="0.32em" class="fill-text-on-dark-secondary ts-primary-body-s"
      >{m.tooltip_total()}</text
    >
    <text
      x={W - PAD}
      y={y + ROW / 2}
      dy="0.32em"
      text-anchor="end"
      class="fill-text-on-dark-primary ts-primary-subheading-s tabular-nums">{tip.total}</text
    >
  {/if}
</g>
