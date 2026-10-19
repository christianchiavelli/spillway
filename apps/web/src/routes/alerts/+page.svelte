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

  const pageOf = (n: number) => (n > 1 ? resolve(`/alerts?page=${n}`) : resolve('/alerts'));
  const firstFiring = $derived((data.page - 1) * data.perPage + 1);
  const lastFiring = $derived(firstFiring + data.firings.items.length - 1);
</script>

<svelte:head>
  <title>Alerts | Spillway</title>
</svelte:head>

<PageHeader
  title="Alerts"
  description="An alert watches one source in one subsystem, and fires on each day its curtailed energy goes over the threshold. Spillway checks the ONS's files every hour."
>
  {#snippet actions()}
    <a href={resolve('/alerts/new')} class={buttonStyles()}>
      <NotificationAdd width={20} height={20} />
      New Alert
    </a>
  {/snippet}
</PageHeader>

<div class="mx-auto flex max-w-[83.25rem] flex-col gap-600 px-600 py-700">
  <Card title="Your Alerts" description="Newest first.">
    {#snippet aside()}
      <p class="ts-primary-body-m text-text-secondary-default">
        <span class="ts-primary-subheading-s text-text-primary-default">{data.alerts.length}</span>
        of {limits.perAccount}
      </p>
    {/snippet}

    {#if form?.delete.failed}
      <ErrorSummary title="That alert could not be deleted" detail="Try again in a moment." />
    {/if}

    {#if data.alerts.length === 0}
      <EmptyState
        title="No alerts yet"
        description="Choose a source, a subsystem and a daily threshold, and see how often it would have fired before you save it."
      >
        {#snippet art()}
          <TurbinesArt />
        {/snippet}
        <a href={resolve('/alerts/new')} class={buttonStyles({ variant: 'secondary' })}>
          <NotificationAdd width={20} height={20} />
          New Alert
        </a>
      </EmptyState>
    {:else}
      <AlertTable alerts={data.alerts} {ondelete} {deleting} />
    {/if}
  </Card>

  {#if data.alerts.length > 0}
    <Card
      title="When They Fired"
      description={data.firings.from && data.firings.through
        ? `From ${dayAndMonth(data.firings.from)} to ${dayAndMonth(data.firings.through)}, the last 30 days the ONS has published, newest first.`
        : "The ONS's data is still loading."}
    >
      {#if data.firings.total === 0}
        <p class="ts-primary-body-m text-text-secondary-default">
          None of your alerts went over its threshold in these days.
        </p>
      {:else if data.firings.items.length === 0}
        <p class="ts-primary-body-m text-text-secondary-default">
          This page is past the end of the list.
          <a href={pageOf(1)} class="text-text-primary-link hover:text-text-primary-link-hover"
            >Go to the newest</a
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
          labels={{ previous: 'Newer', next: 'Older' }}
        />
      {/if}
    </Card>
  {/if}
</div>
