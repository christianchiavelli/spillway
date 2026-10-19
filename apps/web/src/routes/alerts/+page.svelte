<script lang="ts">
  import NotificationAdd from '~icons/material-symbols/notification-add-outline-rounded';

  import { resolve } from '$app/paths';
  import TurbinesArt from '#lib/components/art/TurbinesArt.svelte';
  import { buttonStyles } from '#lib/components/ui/button-styles.ts';
  import Card from '#lib/components/ui/Card.svelte';
  import EmptyState from '#lib/components/ui/EmptyState.svelte';
  import ErrorSummary from '#lib/components/ui/ErrorSummary.svelte';
  import PageHeader from '#lib/components/ui/PageHeader.svelte';
  import Pager from '#lib/components/ui/Pager.svelte';
  import { limits } from '#lib/features/alerts/alerts.ts';
  import AlertTable from '#lib/features/alerts/AlertTable.svelte';
  import FiringTable from '#lib/features/alerts/FiringTable.svelte';
  import { dayAndMonth } from '#lib/features/curtailment/format.ts';
  import { enhanceForm } from '#lib/forms/enhance.ts';
  import { localized } from '#lib/i18n.ts';
  import { m } from '#lib/paraglide/messages.js';

  import type { PageProps } from './$types';

  let { data, form }: PageProps = $props();

  // The page owns the delete forms' action and what a submit does; the table
  // only draws what it is given.
  let deleting = $state<string | null>(null);
  const remove = enhanceForm({
    onpending: (pending) => {
      if (!pending) deleting = null;
    },
  });
  const ondelete: typeof remove = (input) => {
    const id = input.formData.get('id');
    deleting = typeof id === 'string' ? id : null;
    return remove(input);
  };

  const pageOf = (n: number) =>
    localized(n > 1 ? resolve(`/alerts?page=${n}`) : resolve('/alerts'));
  const firstFiring = $derived((data.page - 1) * data.perPage + 1);
  const lastFiring = $derived(firstFiring + data.firings.items.length - 1);
</script>

<svelte:head>
  <title>{m.alerts_head()} | Spillway</title>
</svelte:head>

<PageHeader title={m.alerts_title()} description={m.alerts_description()}>
  {#snippet actions()}
    <a href={localized(resolve('/alerts/new'))} class={buttonStyles()}>
      <NotificationAdd width={20} height={20} />
      {m.alerts_new()}
    </a>
  {/snippet}
</PageHeader>

<div class="mx-auto flex max-w-[83.25rem] flex-col gap-600 px-600 py-700">
  <Card title={m.alerts_yours()} description={m.alerts_newest_first()}>
    {#snippet aside()}
      <p class="ts-primary-body-m text-text-secondary-default">
        <span class="ts-primary-subheading-s text-text-primary-default">{data.alerts.length}</span>
        {m.alerts_of_max({ max: limits.perAccount })}
      </p>
    {/snippet}

    {#if form?.delete.failed}
      <ErrorSummary title={m.alerts_delete_failed()} detail={m.alerts_try_moment()} />
    {/if}

    {#if data.alerts.length === 0}
      <EmptyState title={m.alerts_none_title()} description={m.alerts_none_body()}>
        {#snippet art()}
          <TurbinesArt />
        {/snippet}
        <a href={localized(resolve('/alerts/new'))} class={buttonStyles({ variant: 'secondary' })}>
          <NotificationAdd width={20} height={20} />
          {m.alerts_new()}
        </a>
      </EmptyState>
    {:else}
      <AlertTable alerts={data.alerts} {ondelete} {deleting} />
    {/if}
  </Card>

  {#if data.alerts.length > 0}
    <Card
      title={m.alerts_fired_title()}
      description={data.firings.from && data.firings.through
        ? m.alerts_fired_description({
            from: dayAndMonth(data.firings.from),
            to: dayAndMonth(data.firings.through),
          })
        : m.alerts_data_loading()}
    >
      {#if data.firings.total === 0}
        <p class="ts-primary-body-m text-text-secondary-default">
          {m.alerts_none_fired()}
        </p>
      {:else if data.firings.items.length === 0}
        <p class="ts-primary-body-m text-text-secondary-default">
          {m.plants_past_end()}
          <a href={pageOf(1)} class="text-text-primary-link hover:text-text-primary-link-hover"
            >{m.alerts_go_newest()}</a
          >.
        </p>
      {:else}
        <FiringTable firings={data.firings.items} />
        <Pager
          from={firstFiring}
          to={lastFiring}
          total={data.firings.total}
          previous={data.page > 1 ? pageOf(data.page - 1) : null}
          next={lastFiring < data.firings.total ? pageOf(data.page + 1) : null}
          labels={{ previous: m.alerts_newer(), next: m.alerts_older() }}
        />
      {/if}
    </Card>
  {/if}
</div>
