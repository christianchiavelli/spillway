<script lang="ts">
  import Columns from '#lib/components/charts/Columns.svelte';
  import ChartCard from '#lib/components/charts/ChartCard.svelte';
  import Legend from '#lib/components/charts/Legend.svelte';
  import Lines from '#lib/components/charts/Lines.svelte';
  import Card from '#lib/components/ui/Card.svelte';
  import PageHeader from '#lib/components/ui/PageHeader.svelte';
  import Table from '#lib/components/ui/Table.svelte';
  import DataOnItsWay from '#lib/features/curtailment/DataOnItsWay.svelte';
  import {
    dayAndMonth,
    dayOf,
    int,
    monthName,
    monthYear,
    one,
    pct,
    shareOf,
    shortMonthYear,
    twh,
  } from '#lib/features/curtailment/format.ts';
  import { sources } from '#lib/features/curtailment/grid.ts';
  import { type Year, byYear } from '#lib/features/curtailment/years.ts';
  import { m } from '#lib/paraglide/messages.js';

  import type { PageProps } from './$types';

  let { data }: PageProps = $props();

  const months = $derived(data.history?.months ?? []);
  const years = $derived(byYear(months));
  const lastDay = $derived(data.history ? dayOf(data.history.publishedThrough) : '');
  const firstMonth = $derived(months[0]?.month ?? '');
  const solarFrom = $derived(months.find((month) => month.solar !== undefined)?.month);

  // Two moments the columns cannot explain on their own, where there is room.
  const notes = $derived(
    [
      { month: solarFrom, text: m.history_note_solar(), lift: 64 },
      { month: '2023-08-01', text: m.history_note_blackout(), lift: 98 },
    ].flatMap((n) => {
      const index = months.findIndex((month) => month.month === n.month);
      return index < 0 ? [] : [{ index, text: n.text, lift: n.lift, minWidth: 720 }];
    }),
  );

  // On a phone a year is its total; the sources and their shares are from mobile up.
  const wide = 'hidden mobile:table-cell';

  const yearLabel = (y: Year) => {
    if (y.last.slice(0, 7) === lastDay.slice(0, 7) && !lastDay.endsWith('-12-31')) {
      return m.history_year_to({ year: y.year, day: dayAndMonth(lastDay) });
    }
    if (y.first === firstMonth && !firstMonth.endsWith('-01-01')) {
      return m.history_year_from({ year: y.year, month: monthName(y.first) });
    }
    return y.year;
  };

  const yearTick = (i: number) => {
    const day = months[i]?.month;
    return day?.endsWith('-01-01') ? day.slice(0, 4) : null;
  };

  // Each month's loss over all the plants could have made, in percent.
  const shares = $derived(
    sources.map((s) =>
      months.map((month) => {
        const e = month[s.value];
        return e === undefined ? null : shareOf(e) * 100;
      }),
    ),
  );
  const top = $derived(Math.max(10, ...shares.flat().filter((v): v is number => v !== null)));
</script>

<svelte:head>
  <title>{m.history_head()} | Spillway</title>
</svelte:head>

<PageHeader
  title={m.history_title()}
  description={firstMonth && solarFrom
    ? m.history_description({ wind: monthYear(firstMonth), solar: monthYear(solarFrom) })
    : m.history_description_short()}
/>

<div class="mx-auto flex max-w-[83.25rem] flex-col gap-600 px-600 py-700">
  {#if data.history === null}
    <DataOnItsWay />
  {:else}
    <ChartCard title={m.history_months_title()} description={m.history_months_description()}>
      {#snippet legend()}
        <Legend series={sources} />
      {/snippet}
      <Columns
        title={m.history_months_chart({ month: monthYear(firstMonth) })}
        series={sources}
        values={months.map((month) => [
          (month.wind?.curtailedMwh ?? 0) / 1e6,
          (month.solar?.curtailedMwh ?? 0) / 1e6,
        ])}
        tick={(v: number) => (v === 0 ? '0' : one(v))}
        xLabel={yearTick}
        peak={(i: number, total: number) =>
          `${shortMonthYear(months[i]?.month ?? '')}: ${one(total)} TWh`}
        tip={(i: number) => {
          const month = months[i];
          if (!month) return null;
          return {
            title: monthYear(month.month),
            rows: sources.map((s) => {
              const e = month[s.value];
              return {
                label: s.label,
                value: e === undefined ? m.history_not_published() : `${twh(e.curtailedMwh)} TWh`,
                fill: s.fill,
              };
            }),
          };
        }}
        {notes}
        class="h-[16rem] mobile:h-[20rem]"
      />
      {#snippet table()}
        <Table
          caption={m.history_months_caption()}
          columns={[
            { label: m.history_month() },
            { label: m.table_wind(), right: true },
            { label: m.table_solar(), right: true },
          ]}
          minWidth="mobile:min-w-[20rem]"
        >
          {#each [...months].reverse() as month (month.month)}
            <tr class="ts-primary-body-m text-text-primary-default even:bg-surface-zebra">
              <th scope="row" class="h-[2.25rem] px-300 text-left ts-primary-subheading-s">
                {monthYear(month.month)}
              </th>
              <td class="px-300 text-right tabular-nums">
                {int((month.wind?.curtailedMwh ?? 0) / 1000)}
              </td>
              <td class="px-300 text-right tabular-nums">
                {month.solar === undefined
                  ? m.history_not_published()
                  : int(month.solar.curtailedMwh / 1000)}
              </td>
            </tr>
          {/each}
        </Table>
      {/snippet}
    </ChartCard>

    <ChartCard title={m.history_share_title()} description={m.history_share_description()}>
      {#snippet legend()}
        <Legend series={sources} mark="line" />
      {/snippet}
      <Lines
        title={m.history_share_chart()}
        series={sources}
        values={shares}
        max={top}
        tick={(v: number) => `${v}%`}
        xLabel={yearTick}
        endLabel={(s: number, v: number) => `${sources[s]?.label} ${int(v)}%`}
        class="h-[14rem] mobile:h-[16.25rem]"
      />
      {#snippet table()}
        <Table
          caption={m.history_share_chart()}
          columns={[
            { label: m.history_month() },
            { label: m.table_wind(), right: true },
            { label: m.table_solar(), right: true },
          ]}
          minWidth="mobile:min-w-[20rem]"
        >
          {#each [...months].reverse() as month (month.month)}
            <tr class="ts-primary-body-m text-text-primary-default even:bg-surface-zebra">
              <th scope="row" class="h-[2.25rem] px-300 text-left ts-primary-subheading-s">
                {monthYear(month.month)}
              </th>
              <td class="px-300 text-right tabular-nums">
                {month.wind === undefined ? m.history_not_published() : pct(shareOf(month.wind))}
              </td>
              <td class="px-300 text-right tabular-nums">
                {month.solar === undefined ? m.history_not_published() : pct(shareOf(month.solar))}
              </td>
            </tr>
          {/each}
        </Table>
      {/snippet}
    </ChartCard>

    <Card title={m.history_years_title()}>
      <Table
        caption={m.history_years_caption()}
        columns={[
          { label: m.history_year() },
          { label: m.table_wind(), right: true, class: wide },
          { label: m.history_share(), right: true, class: wide },
          { label: m.table_solar(), right: true, class: wide },
          { label: m.history_share(), right: true, class: wide },
          { label: m.table_total(), right: true },
        ]}
        minWidth="mobile:min-w-[40rem]"
      >
        {#each years as y (y.year)}
          <tr class="ts-primary-body-m text-text-primary-default even:bg-surface-zebra">
            <th scope="row" class="h-[2.25rem] px-300 text-left ts-primary-subheading-s">
              {yearLabel(y)}
            </th>
            <td class={['px-300 text-right tabular-nums', wide]}>
              {y.wind ? `${twh(y.wind.curtailedMwh)} TWh` : ''}
            </td>
            <td class={['px-300 text-right tabular-nums', wide]}>
              {y.wind ? pct(shareOf(y.wind)) : ''}
            </td>
            <td class={['px-300 text-right tabular-nums', wide]}>
              {#if y.solar}
                {twh(y.solar.curtailedMwh)} TWh
              {:else}
                <span class="text-text-secondary-default">{m.history_not_published()}</span>
              {/if}
            </td>
            <td class={['px-300 text-right tabular-nums', wide]}>
              {y.solar ? pct(shareOf(y.solar)) : ''}
            </td>
            <td class="px-300 text-right ts-primary-subheading-s tabular-nums">
              {twh((y.wind?.curtailedMwh ?? 0) + (y.solar?.curtailedMwh ?? 0))} TWh
            </td>
          </tr>
        {/each}
      </Table>
    </Card>
  {/if}
</div>
