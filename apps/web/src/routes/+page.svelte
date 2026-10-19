<script lang="ts">
  import ArrowForward from '~icons/material-symbols/arrow-forward-rounded';

  import { resolve } from '$app/paths';
  import type { PeriodKind, Subsystem } from '#lib/api/generated/types.gen.ts';
  import { buttonStyles } from '#lib/components/ui/button-styles.ts';
  import Card from '#lib/components/ui/Card.svelte';
  import PageHeader from '#lib/components/ui/PageHeader.svelte';
  import PeriodNav from '#lib/components/ui/PeriodNav.svelte';
  import Breakdown from '#lib/features/curtailment/Breakdown.svelte';
  import DataOnItsWay from '#lib/features/curtailment/DataOnItsWay.svelte';
  import { publishedThrough } from '#lib/features/curtailment/format.ts';
  import OverTime from '#lib/features/curtailment/OverTime.svelte';
  import {
    nextLabel,
    periodKinds,
    periodSearch,
    periodTitle,
    previousLabel,
  } from '#lib/features/curtailment/periods.ts';
  import PlantTable from '#lib/features/curtailment/PlantTable.svelte';
  import Totals from '#lib/features/curtailment/Totals.svelte';

  import { localized } from '#lib/i18n.ts';
  import { m } from '#lib/paraglide/messages.js';

  import type { PageProps } from './$types';

  let { data }: PageProps = $props();

  const here = (kind: PeriodKind, date?: string) => {
    const q = periodSearch(kind, date).toString();
    return localized(q ? resolve(`/?${q}`) : resolve('/'));
  };
  const plants = (kind: PeriodKind, date?: string, subsystem?: Subsystem) => {
    const search = periodSearch(kind, date);
    if (subsystem !== undefined) search.set('subsystem', subsystem);
    const q = search.toString();
    return localized(q ? resolve(`/plants?${q}`) : resolve('/plants'));
  };
</script>

<svelte:head>
  <title>{m.overview_head()} | Spillway</title>
</svelte:head>

<PageHeader title={m.overview_title()} description={m.overview_description()} />

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
        label: previousLabel(p.kind),
      }}
      next={{
        href: p.later === undefined ? null : here(p.kind, p.later),
        label: nextLabel(p.kind),
      }}
      note={publishedThrough(o.publishedThrough)}
    />

    <Totals overview={o} />
    <OverTime overview={o} />
    <Breakdown overview={o} plantsIn={(s: Subsystem) => plants(p.kind, data.query.date, s)} />

    <Card title={m.overview_plants_title()} description={m.overview_plants_description()}>
      {#snippet aside()}
        <a
          href={plants(p.kind, data.query.date)}
          class={buttonStyles({ variant: 'ghost', size: 'sm' })}
        >
          {m.overview_all_plants()}
          <ArrowForward width={16} height={16} />
        </a>
      {/snippet}
      {#if data.plants.items.length === 0}
        <p class="ts-primary-body-m text-text-secondary-default">
          {m.overview_no_plant()}
        </p>
      {:else}
        <PlantTable
          plants={data.plants.items}
          caption={m.overview_plants_caption({ period: periodTitle(p) })}
        />
      {/if}
    </Card>

    <p class="ts-primary-body-s text-text-secondary-default">
      {m.overview_source_note()}
    </p>
  {/if}
</div>
