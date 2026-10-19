<script lang="ts">
  import ArrowForward from '~icons/material-symbols/arrow-forward-rounded';

  import { resolve } from '$app/paths';
  import type { PeriodKind } from '#lib/api/generated/types.gen.ts';
  import { buttonStyles } from '#lib/components/ui/button-styles.ts';
  import Card from '#lib/components/ui/Card.svelte';
  import PageHeader from '#lib/components/ui/PageHeader.svelte';
  import PeriodNav from '#lib/components/ui/PeriodNav.svelte';
  import Breakdown from '#lib/features/curtailment/Breakdown.svelte';
  import DataOnItsWay from '#lib/features/curtailment/DataOnItsWay.svelte';
  import { publishedThrough } from '#lib/features/curtailment/format.ts';
  import OverTime from '#lib/features/curtailment/OverTime.svelte';
  import { periodKinds, periodSearch, periodTitle } from '#lib/features/curtailment/periods.ts';
  import PlantTable from '#lib/features/curtailment/PlantTable.svelte';
  import Totals from '#lib/features/curtailment/Totals.svelte';

  import type { PageProps } from './$types';

  let { data }: PageProps = $props();

  const here = (kind: PeriodKind, date?: string) => {
    const q = periodSearch(kind, date).toString();
    return q ? resolve(`/?${q}`) : resolve('/');
  };
  const plants = (kind: PeriodKind, date?: string) => {
    const q = periodSearch(kind, date).toString();
    return q ? resolve(`/plants?${q}`) : resolve('/plants');
  };
</script>

<svelte:head>
  <title>Overview | Spillway</title>
</svelte:head>

<PageHeader
  title="Overview"
  description="How much wind and solar energy the grid could not take, where, and why, from the data the grid operator publishes every day."
/>

<div class="mx-auto flex max-w-[83.25rem] flex-col gap-600 px-600 py-700">
  {#if data.overview === null || data.plants === null}
    <DataOnItsWay />
  {:else}
    {@const o = data.overview}
    {@const p = o.period}
    <PeriodNav
      kinds={periodKinds.map((k) => ({
        label: k.label,
        href: here(k.value, data.query.date),
        current: k.value === p.kind,
      }))}
      title={periodTitle(p)}
      previous={{
        href: p.earlier === undefined ? null : here(p.kind, p.earlier),
        label: `Previous ${p.kind}`,
      }}
      next={{ href: p.later === undefined ? null : here(p.kind, p.later), label: `Next ${p.kind}` }}
      note={publishedThrough(o.publishedThrough)}
    />

    <Totals overview={o} />
    <OverTime overview={o} />
    <Breakdown overview={o} />

    <Card
      title="Plants That Lost the Most"
      description="Some are single plants, others groups that share one connection to the grid."
    >
      {#snippet aside()}
        <a
          href={plants(p.kind, data.query.date)}
          class={buttonStyles({ variant: 'ghost', size: 'sm' })}
        >
          All Plants
          <ArrowForward width={16} height={16} />
        </a>
      {/snippet}
      {#if data.plants.items.length === 0}
        <p class="ts-primary-body-m text-text-secondary-default">
          No plant lost energy in this period.
        </p>
      {:else}
        <PlantTable
          plants={data.plants.items}
          caption={`The plants that lost the most energy, ${periodTitle(p)}`}
        />
      {/if}
    </Card>

    <p class="ts-primary-body-s text-text-secondary-default">
      From the ONS's open data, under CC-BY. The ONS checks each day after it publishes it, so the
      latest figures can still change.
    </p>
  {/if}
</div>
