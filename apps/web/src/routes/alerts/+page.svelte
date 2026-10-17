<script lang="ts">
  import { BellPlus } from '@lucide/svelte';

  import { enhance } from '$app/forms';
  import Button from '#lib/components/ui/Button.svelte';
  import Card from '#lib/components/ui/Card.svelte';
  import EmptyState from '#lib/components/ui/EmptyState.svelte';
  import ErrorSummary from '#lib/components/ui/ErrorSummary.svelte';
  import PageHeader from '#lib/components/ui/PageHeader.svelte';
  import AlertFields from '#lib/features/alerts/AlertFields.svelte';
  import AlertTable from '#lib/features/alerts/AlertTable.svelte';
  import { alertError, alertFields, limits } from '#lib/features/alerts/alerts.ts';
  import TurbinesArt from '#lib/features/alerts/TurbinesArt.svelte';
  import { enhanceForm } from '#lib/forms/enhance.ts';

  import type { PageProps } from './$types';

  let { data, form }: PageProps = $props();

  // The page orchestrates: it owns the forms' actions and what a submit does.
  // The components below only draw what they are given.
  let creating = $state(false);
  let deleting = $state<string | null>(null);

  const create = enhanceForm({
    resetOnRedirect: true,
    onpending: (pending) => {
      creating = pending;
    },
  });
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

  const full = $derived(data.alerts.length >= limits.perAccount);
  const ids = {
    name: 'alert-name',
    source: 'alert-source',
    subsystem: 'alert-subsystem',
    thresholdMwh: 'alert-threshold',
  } as const;
  const problems = $derived(
    alertFields.flatMap((field) => {
      const text = alertError(field, form?.create?.errors[field]);
      return text === undefined ? [] : [{ href: `#${ids[field]}`, text }];
    }),
  );
</script>

<svelte:head>
  <title>Alerts | Spillway</title>
</svelte:head>

<PageHeader
  title="Alerts"
  description="An alert is a daily limit on the wind or solar energy the grid operator keeps from being generated in one subsystem."
/>

<div class="mx-auto flex max-w-[83.25rem] flex-col gap-600 px-600 py-700">
  <Card
    title="New Alert"
    description="Choose a source and a subsystem to watch, and how much curtailed energy in one day is too much."
  >
    {#if full}
      <ErrorSummary
        title="You have {limits.perAccount} alerts, the most an account can keep"
        detail="Delete one below to make room for another."
      />
    {:else if form?.create?.full}
      <ErrorSummary
        title="You have {limits.perAccount} alerts, the most an account can keep"
        detail="Delete one to make room for another."
      />
    {:else if problems.length > 0}
      <ErrorSummary title="Check the alert" items={problems} />
    {/if}

    <form
      method="POST"
      action="?/create"
      novalidate
      use:enhance={create}
      class="flex flex-col gap-700 pt-300"
    >
      <AlertFields values={form?.create?.values} errors={form?.create?.errors} />
      <div class="flex justify-end">
        <Button type="submit" icon={BellPlus} disabled={full} aria-disabled={creating || undefined}>
          {creating ? 'Creating Alert…' : 'Create Alert'}
        </Button>
      </div>
    </form>
  </Card>

  <Card title="Your Alerts" description="Newest first.">
    {#snippet aside()}
      <p class="ts-primary-body-m text-text-secondary-default">
        <span class="ts-primary-subheading-s text-text-primary-default">{data.alerts.length}</span>
        of {limits.perAccount}
      </p>
    {/snippet}

    {#if form?.delete?.failed}
      <ErrorSummary title="That alert could not be deleted" detail="Try again in a moment." />
    {/if}

    {#if data.alerts.length === 0}
      <EmptyState
        title="No alerts yet"
        description="Create one above. It will show here, with its threshold and the day you made it."
      >
        {#snippet art()}
          <TurbinesArt />
        {/snippet}
      </EmptyState>
    {:else}
      <AlertTable alerts={data.alerts} {ondelete} {deleting} />
    {/if}
  </Card>
</div>
