<script lang="ts">
  import type { ResolvedPathname } from '$app/types';
  import type { Curtailment, Subsystem } from '#lib/api/generated/types.gen.ts';
  import BarRows from '#lib/components/charts/BarRows.svelte';
  import ChartCard from '#lib/components/charts/ChartCard.svelte';
  import Choropleth from '#lib/components/charts/Choropleth.svelte';
  import Legend from '#lib/components/charts/Legend.svelte';
  import Card from '#lib/components/ui/Card.svelte';

  import { m } from '#lib/paraglide/messages.js';

  import { energy } from './format.ts';
  import { reasons, sources, subsystemLabel } from './grid.ts';
  import { periodTitle } from './periods.ts';
  import SplitTable from './SplitTable.svelte';
  import { height, shapes, width } from './subsystem-shapes.ts';

  type Props = {
    overview: Curtailment;
    /** The plants page for one subsystem, in the same period. */
    plantsIn: (subsystem: Subsystem) => ResolvedPathname;
  };

  let { overview, plantsIn }: Props = $props();

  /** A subsystem or a reason, as both cards rank them. */
  type Split = {
    key: string;
    label: string;
    description?: string;
    windMwh: number;
    solarMwh: number;
    href?: ResolvedPathname;
  };

  const total = (s: Split) => s.windMwh + s.solarMwh;
  const ranked = <T extends Split>(splits: T[]) => splits.sort((a, b) => total(b) - total(a));
  const bars = (splits: readonly Split[]) =>
    splits.map((s) => ({ ...s, values: [s.windMwh, s.solarMwh], value: energy(total(s)) }));

  const title = $derived(periodTitle(overview.period));
  const bySubsystem = $derived(
    ranked(
      overview.bySubsystem.map((s) => ({
        subsystem: s.subsystem,
        key: s.subsystem,
        label: subsystemLabel(s.subsystem),
        windMwh: s.windMwh,
        solarMwh: s.solarMwh,
        href: plantsIn(s.subsystem),
      })),
    ),
  );
  const byReason = $derived(
    ranked(
      overview.byReason.map((r) => ({
        key: r.reason,
        label: reasons[r.reason].label,
        description: reasons[r.reason].meaning,
        windMwh: r.windMwh,
        solarMwh: r.solarMwh,
      })),
    ),
  );

  // The subsystem pointed at, on the map or in the list: each shows it in the other.
  let active = $state<string | null>(null);
  const point = (key: string | null) => {
    active = key;
  };
</script>

<!--
  Where and why: the subsystems take the wider card, since the map needs the
  room, and the reasons the narrower one. Each ranks its rows on one scale, the
  wind and solar in each bar in the colours they have everywhere.
-->
<div class="grid gap-600 laptop:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
  <ChartCard
    title={m.breakdown_subsystem_title()}
    description={m.breakdown_subsystem_description()}
  >
    {#snippet legend()}
      <Legend series={sources} />
    {/snippet}
    <!-- Beside the figures where the card is wide enough, over them where not. -->
    <div class="@container">
      <div class="grid items-center gap-700 @xl:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
        <div class="mx-auto w-full max-w-[22rem]">
          <Choropleth
            {width}
            {height}
            regions={bySubsystem.map((s) => ({
              ...s,
              ...shapes[s.subsystem],
              value: total(s),
            }))}
            {active}
            onactive={point}
          />
        </div>
        <BarRows
          title={m.breakdown_subsystem_list()}
          series={sources}
          rows={bars(bySubsystem)}
          {active}
          onactive={point}
        />
      </div>
    </div>
    {#snippet table()}
      <SplitTable
        caption={m.breakdown_subsystem_caption({ period: title })}
        heading={m.breakdown_subsystem_heading()}
        rows={bySubsystem}
      />
    {/snippet}
  </ChartCard>

  {#if byReason.length === 0}
    <Card title={m.breakdown_reason_title()} description={m.breakdown_reason_description()}>
      <p class="ts-primary-body-m text-text-secondary-default">
        {m.breakdown_reason_none()}
      </p>
    </Card>
  {:else}
    <ChartCard title={m.breakdown_reason_title()} description={m.breakdown_reason_description()}>
      {#snippet legend()}
        <Legend series={sources} />
      {/snippet}
      <BarRows title={m.breakdown_reason_list()} series={sources} rows={bars(byReason)} />
      {#snippet table()}
        <SplitTable
          caption={m.breakdown_reason_caption({ period: title })}
          heading={m.breakdown_reason_heading()}
          rows={byReason}
        />
      {/snippet}
    </ChartCard>
  {/if}
</div>
