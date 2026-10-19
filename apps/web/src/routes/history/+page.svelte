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

  import type { PageProps } from './$types';

  let { data }: PageProps = $props();

  const months = $derived(data.history?.months ?? []);
  const years = $derived(byYear(months));
  const lastDay = $derived(data.history ? dayOf(data.history.publishedThrough) : '');
  const firstMonth = $derived(months[0]?.month ?? '');
  const solarFrom = $derived(months.find((m) => m.solar !== undefined)?.month);

  // Two moments the columns cannot explain on their own, where there is room.
  const notes = $derived(
    [
      { month: solarFrom, text: 'Solar data begins', lift: 64 },
      { month: '2023-08-01', text: 'Blackout of August 15, 2023', lift: 98 },
    ].flatMap((n) => {
      const index = months.findIndex((m) => m.month === n.month);
      return index < 0 ? [] : [{ index, text: n.text, lift: n.lift, minWidth: 720 }];
    }),
  );

  // On a phone a year is its total; the sources and their shares are from mobile up.
  const wide = 'hidden mobile:table-cell';

  const yearLabel = (y: Year) => {
    if (y.last.slice(0, 7) === lastDay.slice(0, 7) && !lastDay.endsWith('-12-31')) {
      return `${y.year}, to ${dayAndMonth(lastDay)}`;
    }
    if (y.first === firstMonth && !firstMonth.endsWith('-01-01')) {
      return `${y.year}, from ${monthName(y.first)}`;
    }
    return y.year;
  };

  const yearTick = (i: number) => {
    const m = months[i]?.month;
    return m?.endsWith('-01-01') ? m.slice(0, 4) : null;
  };

  // Each month's loss over all the plants could have made, in percent.
  const shares = $derived(
    sources.map((s) =>
      months.map((m) => {
        const e = m[s.value];
        return e === undefined ? null : shareOf(e) * 100;
      }),
    ),
  );
  const top = $derived(Math.max(10, ...shares.flat().filter((v): v is number => v !== null)));
</script>

<svelte:head>
  <title>History | Spillway</title>
</svelte:head>

<PageHeader
  title="History"
  description={firstMonth && solarFrom
    ? `Curtailment since the ONS began publishing it: ${monthYear(firstMonth)} for wind, ${monthYear(solarFrom)} for solar.`
    : 'Curtailment since the ONS began publishing it.'}
/>

<div class="mx-auto flex max-w-[83.25rem] flex-col gap-600 px-600 py-700">
  {#if data.history === null}
    <DataOnItsWay />
  {:else}
    <ChartCard
      title="Curtailed Each Month"
      description="In terawatt hours (TWh), a million megawatt hours."
    >
      {#snippet legend()}
        <Legend series={sources} />
      {/snippet}
      <Columns
        title={`Energy curtailed each month since ${monthYear(firstMonth)}`}
        series={sources}
        values={months.map((m) => [
          (m.wind?.curtailedMwh ?? 0) / 1e6,
          (m.solar?.curtailedMwh ?? 0) / 1e6,
        ])}
        tick={(v: number) => (v === 0 ? '0' : one(v))}
        xLabel={yearTick}
        peak={(i: number, total: number) =>
          `${shortMonthYear(months[i]?.month ?? '')}: ${one(total)} TWh`}
        tip={(i: number) => {
          const m = months[i];
          if (!m) return null;
          return {
            title: monthYear(m.month),
            rows: sources.map((s) => {
              const e = m[s.value];
              return {
                label: s.label,
                value: e === undefined ? 'Not published' : `${twh(e.curtailedMwh)} TWh`,
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
          caption="Energy curtailed each month, in GWh"
          columns={[
            { label: 'Month' },
            { label: 'Wind', right: true },
            { label: 'Solar', right: true },
          ]}
          minWidth="mobile:min-w-[20rem]"
        >
          {#each [...months].reverse() as m (m.month)}
            <tr class="ts-primary-body-m text-text-primary-default even:bg-surface-zebra">
              <th scope="row" class="h-[2.25rem] px-300 text-left ts-primary-subheading-s">
                {monthYear(m.month)}
              </th>
              <td class="px-300 text-right tabular-nums">
                {int((m.wind?.curtailedMwh ?? 0) / 1000)}
              </td>
              <td class="px-300 text-right tabular-nums">
                {m.solar === undefined ? 'Not published' : int(m.solar.curtailedMwh / 1000)}
              </td>
            </tr>
          {/each}
        </Table>
      {/snippet}
    </ChartCard>

    <ChartCard
      title="Share of What Could Have Been Generated"
      description="Each month, the curtailed energy over all the plants could have made."
    >
      {#snippet legend()}
        <Legend series={sources} mark="line" />
      {/snippet}
      <Lines
        title="Share of what wind and solar plants could have generated, each month"
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
          caption="Share of what wind and solar plants could have generated, each month"
          columns={[
            { label: 'Month' },
            { label: 'Wind', right: true },
            { label: 'Solar', right: true },
          ]}
          minWidth="mobile:min-w-[20rem]"
        >
          {#each [...months].reverse() as m (m.month)}
            <tr class="ts-primary-body-m text-text-primary-default even:bg-surface-zebra">
              <th scope="row" class="h-[2.25rem] px-300 text-left ts-primary-subheading-s">
                {monthYear(m.month)}
              </th>
              <td class="px-300 text-right tabular-nums">
                {m.wind === undefined ? 'Not published' : pct(shareOf(m.wind))}
              </td>
              <td class="px-300 text-right tabular-nums">
                {m.solar === undefined ? 'Not published' : pct(shareOf(m.solar))}
              </td>
            </tr>
          {/each}
        </Table>
      {/snippet}
    </ChartCard>

    <Card title="Year by Year">
      <Table
        caption="Curtailed energy by year"
        columns={[
          { label: 'Year' },
          { label: 'Wind', right: true, class: wide },
          { label: 'Share', right: true, class: wide },
          { label: 'Solar', right: true, class: wide },
          { label: 'Share', right: true, class: wide },
          { label: 'Total', right: true },
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
                <span class="text-text-secondary-default">Not published</span>
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
