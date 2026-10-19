<script lang="ts">
  import type { Curtailment } from '#lib/api/generated/types.gen.ts';
  import { m } from '#lib/paraglide/messages.js';

  import { compared, gwh, int, pct, shareOf } from './format.ts';
  import { sources } from './grid.ts';
  import { inThePeriod, periodHeadline, previousName } from './periods.ts';

  type Props = { overview: Curtailment };

  let { overview }: Props = $props();

  const p = $derived(overview.period);
  const total = $derived({
    curtailedMwh: overview.wind.curtailedMwh + overview.solar.curtailedMwh,
    generatedMwh: overview.wind.generatedMwh + overview.solar.generatedMwh,
  });
  const before = $derived(previousName(p));

  /** Big figures in GWh, or in MWh while they are under one. */
  const figure = (mwh: number) =>
    mwh >= 1000 ? { value: gwh(mwh), unit: 'GWh' } : { value: int(mwh), unit: 'MWh' };
  const hero = $derived(figure(total.curtailedMwh));
</script>

<!--
  The totals band: the period's whole loss first and largest, then each source
  and the plants. Two columns on a phone, one row from the laptop up.
-->
<dl
  class="grid grid-cols-2 rounded-300 border border-stroke-neutral-default bg-surface-container-primary laptop:grid-cols-[minmax(0,1.5fr)_repeat(3,minmax(0,1fr))]"
>
  <div
    class="col-span-2 flex flex-col gap-200 border-b border-stroke-neutral-default p-600 laptop:col-span-1 laptop:border-r laptop:border-b-0"
  >
    <dt class="ts-primary-subheading-s text-text-secondary-default">{periodHeadline(p)}</dt>
    <dd class="text-text-primary-default">
      <span class="text-[3rem] leading-[3.5rem] font-[600] tracking-[-0.04em]">{hero.value}</span>
      <span class="ts-primary-subheading-l">{hero.unit}</span>
    </dd>
    <dd class="ts-primary-body-m text-text-secondary-default">
      {m.totals_share_generated({ share: pct(shareOf(total)) })}
    </dd>
    {#if before !== undefined && overview.previousCurtailedMwh !== undefined}
      <dd class="ts-primary-body-m text-text-primary-default">
        {compared(total.curtailedMwh, overview.previousCurtailedMwh, before)}
      </dd>
    {/if}
  </div>

  {#each sources as source, i (source.value)}
    {@const e = overview[source.value]}
    {@const f = figure(e.curtailedMwh)}
    <div
      class={[
        'flex flex-col gap-200 border-b border-stroke-neutral-default p-600 laptop:border-b-0',
        i === 0 ? 'border-r' : 'laptop:border-r',
      ]}
    >
      <dt class="flex items-center gap-200 ts-primary-subheading-s text-text-secondary-default">
        <span class={['size-[0.625rem] rounded-100', source.swatch]}></span>{source.label}
      </dt>
      <dd class="text-text-primary-default">
        <span class="ts-primary-subheading-l">{f.value}</span>
        <span class="ts-primary-subheading-s">{f.unit}</span>
      </dd>
      <dd class="ts-primary-body-m text-text-secondary-default">
        {m.totals_share_potential({ share: pct(shareOf(e)) })}
      </dd>
    </div>
  {/each}

  <div class="col-span-2 flex flex-col gap-200 p-600 laptop:col-span-1">
    <dt class="ts-primary-subheading-s text-text-secondary-default">{m.totals_plants_limited()}</dt>
    <dd class="ts-primary-subheading-l text-text-primary-default">{int(overview.plantsLimited)}</dd>
    <dd class="ts-primary-body-m text-text-secondary-default">
      {m.totals_plants_reporting({
        reporting: int(overview.plantsReporting),
        inPeriod: inThePeriod(p),
      })}
    </dd>
  </div>
</dl>
