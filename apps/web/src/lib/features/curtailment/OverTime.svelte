<script lang="ts">
  import type { Curtailment } from '#lib/api/generated/types.gen.ts';
  import Columns from '#lib/components/charts/Columns.svelte';
  import ChartCard from '#lib/components/charts/ChartCard.svelte';
  import Legend from '#lib/components/charts/Legend.svelte';
  import Table from '#lib/components/ui/Table.svelte';

  import { dayOfMonth, energy, int, longDay, one, shortDay, weekdayShort } from './format.ts';
  import { sources } from './grid.ts';
  import { periodTitle } from './periods.ts';
  import { dailySlots, halfHours } from './series.ts';

  type Props = { overview: Curtailment };

  let { overview }: Props = $props();

  const p = $derived(overview.period);
  const title = $derived(periodTitle(p));
  const byHalfHour = $derived(p.kind === 'day');
  const halves = $derived(byHalfHour ? halfHours(overview) : []);
  const days = $derived(byHalfHour ? [] : dailySlots(overview));
  const slots = $derived(byHalfHour ? halves.map((h) => h.values) : days.map((d) => d.values));
  // A day in MWh, longer periods in GWh, so the axis keeps short numbers.
  const unit = $derived(byHalfHour ? 1 : 1000);
  const values = $derived(slots.map((v) => (v ? [v.wind / unit, v.solar / unit] : null)));

  const xLabel = (i: number) => {
    if (byHalfHour) return i % 6 === 0 ? (halves[i]?.time ?? null) : null;
    const day = days[i]?.day;
    if (day === undefined) return null;
    if (p.kind === 'week') return `${weekdayShort(day)} ${dayOfMonth(day)}`;
    const n = dayOfMonth(day);
    return n === 1 || n % 5 === 0 ? String(n) : null;
  };
  const xLabelShort = (i: number) => {
    if (byHalfHour) return i % 12 === 0 ? (halves[i]?.time ?? null) : null;
    const day = days[i]?.day;
    if (day === undefined) return null;
    if (p.kind === 'week') return weekdayShort(day);
    return [1, 10, 20, 30].includes(dayOfMonth(day)) ? String(dayOfMonth(day)) : null;
  };

  const tip = (i: number) => {
    const v = slots[i];
    if (!v) return null;
    const heading = byHalfHour
      ? `${halves[i]?.time} to ${halves[i]?.end}`
      : shortDay(days[i]?.day ?? '');
    // In the unit of the chart's axis: MWh for half hours, GWh for days.
    const figure = (mwh: number) => (byHalfHour ? `${int(mwh)} MWh` : energy(mwh));
    return {
      title: heading,
      rows: sources.map((s) => ({ label: s.label, value: figure(v[s.value]), fill: s.fill })),
      total: figure(v.wind + v.solar),
    };
  };
</script>

<ChartCard
  title={byHalfHour ? 'Hour by Hour' : 'Day by Day'}
  description={byHalfHour
    ? 'Energy curtailed in each half hour, in megawatt hours (MWh)'
    : 'Energy curtailed each day, in gigawatt hours (GWh)'}
>
  {#snippet legend()}
    <Legend series={sources} />
  {/snippet}

  <Columns
    title={byHalfHour
      ? `Energy curtailed in each half hour of ${title}`
      : `Energy curtailed each day, ${title}`}
    series={sources}
    {values}
    tick={(v: number) => int(v)}
    {xLabel}
    {xLabelShort}
    peak={(i: number, total: number) =>
      byHalfHour ? `${int(total)} MWh at ${halves[i]?.time}` : `${one(total)} GWh`}
    emptyNote={byHalfHour ? undefined : 'Not published yet'}
    {tip}
    class="h-[13.75rem] mobile:h-[18.75rem]"
  />

  {#snippet table()}
    <Table
      caption={byHalfHour
        ? `Energy curtailed in each half hour of ${title}, in MWh`
        : `Energy curtailed each day, ${title}, in GWh`}
      columns={[
        { label: byHalfHour ? 'Half hour' : 'Day' },
        { label: 'Wind', right: true },
        { label: 'Solar', right: true },
      ]}
      minWidth="mobile:min-w-[20rem]"
    >
      {#if byHalfHour}
        {#each halves.filter((h) => h.values !== null) as h (h.time)}
          <tr class="ts-primary-body-m text-text-primary-default even:bg-surface-zebra">
            <th scope="row" class="h-[2.25rem] px-300 text-left ts-primary-subheading-s"
              >{h.time} to {h.end}</th
            >
            <td class="px-300 text-right tabular-nums">{int(h.values?.wind ?? 0)}</td>
            <td class="px-300 text-right tabular-nums">{int(h.values?.solar ?? 0)}</td>
          </tr>
        {/each}
      {:else}
        {#each days.filter((d) => d.values !== null) as d (d.day)}
          <tr class="ts-primary-body-m text-text-primary-default even:bg-surface-zebra">
            <th scope="row" class="h-[2.25rem] px-300 text-left ts-primary-subheading-s"
              >{longDay(d.day)}</th
            >
            <td class="px-300 text-right tabular-nums">{one((d.values?.wind ?? 0) / 1000)}</td>
            <td class="px-300 text-right tabular-nums">{one((d.values?.solar ?? 0) / 1000)}</td>
          </tr>
        {/each}
      {/if}
    </Table>
  {/snippet}
</ChartCard>
