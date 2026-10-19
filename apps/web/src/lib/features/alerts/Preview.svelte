<script lang="ts">
  import type { DailyCurtailment, Source, Subsystem } from '#lib/api/generated/types.gen.ts';
  import Columns from '#lib/components/charts/Columns.svelte';
  import { dayAndMonth, int, shortDay } from '#lib/features/curtailment/format.ts';
  import { sourceOf, subsystemLabel } from '#lib/features/curtailment/grid.ts';
  import { addDays } from '#lib/features/curtailment/series.ts';

  type Props = {
    recent: DailyCurtailment;
    source: Source;
    /** Not chosen yet, until the field is. */
    subsystem: Subsystem | undefined;
    /** Not a usable threshold yet, until one is typed. */
    threshold: number | undefined;
  };

  let { recent, source, subsystem, threshold }: Props = $props();

  const s = $derived(sourceOf(source));
  const days = $derived(
    recent.series.find((x) => x.source === source && x.subsystem === subsystem)?.curtailedMwh ?? [],
  );
  const dayAt = (i: number) => addDays(recent.from, i);
  // An alert fires on a day that loses more than its threshold, as the API counts.
  const fired = $derived(
    threshold === undefined ? [] : days.flatMap((v, i) => (v > threshold ? [i] : [])),
  );
  const max = $derived(Math.max(0, ...days, threshold ?? 0));
  const tick = (v: number) => (max >= 10_000 && v >= 1000 ? `${int(v / 1000)}K` : int(v));
</script>

{#if subsystem === undefined}
  <p class="ts-primary-body-l text-text-secondary-default">
    Choose a subsystem to see what its {s.label.toLowerCase()} plants lost in the last
    {days.length || 30} days the ONS has published.
  </p>
{:else}
  <Columns
    title={`${s.label} in the ${subsystemLabel(subsystem)}, the last ${days.length} days${
      threshold === undefined ? '' : ', against the threshold'
    }`}
    series={[s]}
    values={days.map((v) => [v])}
    {tick}
    xLabel={(i: number) => (i % 7 === 0 ? shortDay(dayAt(i)) : null)}
    xLabelShort={(i: number) => (i % 14 === 0 ? shortDay(dayAt(i)) : null)}
    rule={threshold === undefined
      ? undefined
      : { value: threshold, label: `${int(threshold)} MWh`, caption: 'threshold' }}
    fill={(i: number) =>
      threshold === undefined || (days[i] ?? 0) > threshold ? s.fill : 'fill-chart-muted'}
    tip={(i: number) => ({
      title: shortDay(dayAt(i)),
      rows: [{ label: s.label, value: `${int(days[i] ?? 0)} MWh`, fill: s.fill }],
    })}
    fallbackWidth={718}
    class="min-h-[16.25rem] flex-1"
  />
  <p class="ts-primary-body-l text-text-primary-default" aria-live="polite">
    {#if threshold === undefined}
      Enter a daily threshold to see how often it would have fired.
    {:else if fired.length === 0}
      It would not have fired in the last {days.length} days.
    {:else}
      It would have fired on
      <span class="ts-primary-subheading-m"
        >{fired.length} of the last {days.length} {days.length === 1 ? 'day' : 'days'}</span
      >, last on {dayAndMonth(dayAt(fired.at(-1) ?? 0))}.
    {/if}
  </p>
{/if}
