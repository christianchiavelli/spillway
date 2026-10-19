<script lang="ts">
  import ArrowDown from '~icons/material-symbols/keyboard-arrow-down-rounded';

  import { resolve } from '$app/paths';
  import type { PeriodKind, Source, Subsystem } from '#lib/api/generated/types.gen.ts';
  import { buttonStyles } from '#lib/components/ui/button-styles.ts';
  import Card from '#lib/components/ui/Card.svelte';
  import PageHeader from '#lib/components/ui/PageHeader.svelte';
  import Pager from '#lib/components/ui/Pager.svelte';
  import PeriodNav from '#lib/components/ui/PeriodNav.svelte';
  import PillNav from '#lib/components/ui/PillNav.svelte';
  import DataOnItsWay from '#lib/features/curtailment/DataOnItsWay.svelte';
  import { dayAndMonth, int, publishedThrough } from '#lib/features/curtailment/format.ts';
  import { sources, subsystems } from '#lib/features/curtailment/grid.ts';
  import { periodKinds, periodSearch, periodTitle } from '#lib/features/curtailment/periods.ts';
  import PlantTable from '#lib/features/curtailment/PlantTable.svelte';

  import type { PageProps } from './$types';

  let { data }: PageProps = $props();

  type Query = {
    kind: PeriodKind;
    date: string | undefined;
    source: Source | undefined;
    subsystem: Subsystem | undefined;
    page: number;
  };
  const current = $derived<Query>({
    kind: data.ranking?.period.kind ?? data.query.period,
    date: data.query.date,
    source: data.source,
    subsystem: data.subsystem,
    page: data.page,
  });

  /** The address of this list with some of its choices changed. */
  function href(change: Partial<Query>) {
    const q = { ...current, page: 1, ...change };
    const search = periodSearch(q.kind, q.date);
    if (q.source) search.set('source', q.source);
    if (q.subsystem) search.set('subsystem', q.subsystem);
    if (q.page > 1) search.set('page', String(q.page));
    return search.size ? resolve(`/plants?${search.toString()}`) : resolve('/plants');
  }
</script>

<svelte:head>
  <title>Plants | Spillway</title>
</svelte:head>

<PageHeader
  title="Plants"
  description="Every wind and solar plant the ONS limited, by the energy it lost. Some are single plants, others groups that share one connection to the grid."
/>

<div class="mx-auto flex max-w-[83.25rem] flex-col gap-600 px-600 py-700">
  {#if data.ranking === null}
    <DataOnItsWay />
  {:else}
    {@const r = data.ranking}
    {@const p = r.period}
    <PeriodNav
      kinds={periodKinds.map((k) => ({
        label: k.label,
        href: href({ kind: k.value }),
        current: k.value === p.kind,
      }))}
      title={periodTitle(p)}
      previous={{
        href: p.earlier === undefined ? null : href({ date: p.earlier }),
        label: `Previous ${p.kind}`,
      }}
      next={{
        href: p.later === undefined ? null : href({ date: p.later }),
        label: `Next ${p.kind}`,
      }}
      note={publishedThrough(r.publishedThrough)}
    />

    <Card
      title={`${int(r.total)} ${r.total === 1 ? 'Plant' : 'Plants'} Limited`}
      description={`Of the ${int(r.reporting)} that reported ${
        p.start === p.through
          ? `on ${dayAndMonth(p.start)}`
          : `from ${dayAndMonth(p.start)} to ${dayAndMonth(p.through)}`
      }, the most energy lost first.`}
    >
      <!--
        The source and the subsystem narrow this list and nothing else, so they
        sit in its header; the period, which the whole page follows, stays above.
      -->
      {#snippet aside()}
        <div class="flex w-full flex-wrap items-center gap-400 mobile:w-auto">
          <PillNav
            label="Source"
            items={[
              {
                label: 'All',
                href: href({ source: undefined }),
                current: data.source === undefined,
              },
              ...sources.map((s) => ({
                label: s.label,
                icon: s.icon,
                href: href({ source: s.value }),
                current: data.source === s.value,
              })),
            ]}
          />
          <!--
            A form that asks for a new address, so the choice works without
            JavaScript, with the button shown only then; with it, choosing is
            enough.
          -->
          <form
            method="GET"
            action={resolve('/plants')}
            class="flex w-full items-center gap-300 mobile:w-auto"
          >
            {#if current.kind !== 'day'}
              <input type="hidden" name="period" value={current.kind} />
            {/if}
            {#if current.date}
              <input type="hidden" name="date" value={current.date} />
            {/if}
            {#if current.source}
              <input type="hidden" name="source" value={current.source} />
            {/if}
            <div
              class="relative flex h-[2.375rem] w-full items-center rounded-200 border border-stroke-control bg-surface-container-primary hover:border-stroke-control-hover has-[select:focus-visible]:outline-2 has-[select:focus-visible]:outline-offset-2 has-[select:focus-visible]:outline-stroke-primary-default mobile:w-[15rem]"
            >
              <select
                name="subsystem"
                aria-label="Subsystem"
                onchange={(e) => e.currentTarget.form?.requestSubmit()}
                class="select-menu h-full w-full cursor-pointer rounded-200 bg-transparent pr-800 pl-300 ts-primary-body-m text-text-primary-default outline-none"
              >
                <option value="" selected={data.subsystem === undefined}>All subsystems</option>
                {#each subsystems as s (s.value)}
                  <option value={s.value} selected={data.subsystem === s.value}>{s.label}</option>
                {/each}
              </select>
              <ArrowDown
                width={16}
                height={16}
                class="pointer-events-none absolute right-300 text-icon-secondary"
              />
            </div>
            <noscript>
              <button type="submit" class={buttonStyles({ variant: 'secondary', size: 'sm' })}
                >Show</button
              >
            </noscript>
          </form>
        </div>
      {/snippet}
      {#if r.items.length === 0}
        <p class="ts-primary-body-m text-text-secondary-default">
          {r.total === 0
            ? 'No plant lost energy in this period.'
            : 'This page is past the end of the list.'}
        </p>
      {:else}
        {@const first = (data.page - 1) * data.pageSize + 1}
        <PlantTable plants={r.items} caption={`Plants limited, ${periodTitle(p)}`} {first} full />
        <Pager
          from={first}
          to={first + r.items.length - 1}
          total={r.total}
          previous={data.page > 1 ? href({ page: data.page - 1 }) : null}
          next={first + r.items.length - 1 < r.total ? href({ page: data.page + 1 }) : null}
        />
      {/if}
    </Card>
  {/if}
</div>
