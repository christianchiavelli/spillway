<script lang="ts">
  import type { DailyCurtailment, Source, Subsystem } from '#lib/api/generated/types.gen.ts';
  import Columns from '#lib/components/charts/Columns.svelte';
  import EmptyState from '#lib/components/ui/EmptyState.svelte';
  import { dayAndMonth, int, shortDay } from '#lib/features/curtailment/format.ts';
  import { sourceOf, subsystemLabel } from '#lib/features/curtailment/grid.ts';
  import { addDays } from '#lib/features/curtailment/series.ts';
  import { m } from '#lib/paraglide/messages.js';

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
  const tick = (v: number) =>
    max >= 10_000 && v >= 1000 ? m.preview_thousands({ n: int(v / 1000) }) : int(v);
  // Its own sentence for each source: in Portuguese the adjective agrees with
  // "usinas", so it is not the source's label lowercased.
  const choose = $derived(source === 'solar' ? m.preview_choose_solar() : m.preview_choose_wind());
</script>

{#if subsystem === undefined}
  <!--
    Where the chart will go, and as tall, so it takes this place once a
    subsystem is chosen. The icon is the chosen source's, so it follows the form.
  -->
  <div class="flex min-h-[16.25rem] flex-1 items-center justify-center">
    <EmptyState icon={s.icon} title={m.preview_no_subsystem()} description={choose} />
  </div>
{:else}
  <Columns
    title={(threshold === undefined ? m.preview_chart : m.preview_chart_threshold)({
      source: s.label,
      subsystem: subsystemLabel(subsystem),
      days: int(days.length),
    })}
    series={[s]}
    values={days.map((v) => [v])}
    {tick}
    xLabel={(i: number) => (i % 7 === 0 ? shortDay(dayAt(i)) : null)}
    xLabelShort={(i: number) => (i % 14 === 0 ? shortDay(dayAt(i)) : null)}
    rule={threshold === undefined
      ? undefined
      : { value: threshold, label: `${int(threshold)} MWh`, caption: m.preview_threshold() }}
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
      {m.preview_enter_threshold()}
    {:else if fired.length === 0}
      {m.preview_never({ days: int(days.length) })}
    {:else}
      {m.preview_fired_on()}
      <span class="ts-primary-subheading-m"
        >{m.preview_fired_count({
          count: days.length,
          fired: int(fired.length),
          days: int(days.length),
        })}</span
      >{m.preview_fired_last({ day: dayAndMonth(dayAt(fired.at(-1) ?? 0)) })}
    {/if}
  </p>
{/if}
