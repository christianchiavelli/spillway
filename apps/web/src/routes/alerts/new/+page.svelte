<script lang="ts">
  import Downloading from '~icons/material-symbols/downloading-rounded';
  import NotificationAdd from '~icons/material-symbols/notification-add-outline-rounded';

  import { enhance } from '$app/forms';
  import { resolve } from '$app/paths';
  import Button from '#lib/components/ui/Button.svelte';
  import { buttonStyles } from '#lib/components/ui/button-styles.ts';
  import Card from '#lib/components/ui/Card.svelte';
  import EmptyState from '#lib/components/ui/EmptyState.svelte';
  import ErrorSummary from '#lib/components/ui/ErrorSummary.svelte';
  import PageHeader from '#lib/components/ui/PageHeader.svelte';
  import AlertFields from '#lib/features/alerts/AlertFields.svelte';
  import { alertError, alertFields, limits } from '#lib/features/alerts/alerts.ts';
  import Preview from '#lib/features/alerts/Preview.svelte';
  import { sources, subsystemLabel, subsystems } from '#lib/features/curtailment/grid.ts';
  import { enhanceForm } from '#lib/forms/enhance.ts';

  import type { PageProps } from './$types';

  let { data, form }: PageProps = $props();

  let creating = $state(false);
  const create = enhanceForm({
    onpending: (pending) => {
      creating = pending;
    },
  });

  // The preview follows the fields as they are filled in. Before anything is
  // typed, and without JavaScript, it shows what the form holds.
  type Draft = { source: string; subsystem: string; threshold: string };
  let typed = $state<Draft | null>(null);
  const draft = $derived<Draft>(
    typed ?? {
      source: form?.create.values.source ?? 'wind',
      subsystem: form?.create.values.subsystem ?? '',
      threshold: form?.create.values.thresholdMwh ?? '',
    },
  );
  function read(element: HTMLFormElement) {
    const fields = new FormData(element);
    const value = (name: string) => {
      const v = fields.get(name);
      return typeof v === 'string' ? v : '';
    };
    typed = {
      source: value('source'),
      subsystem: value('subsystem'),
      threshold: value('thresholdMwh'),
    };
  }
  const source = $derived(sources.find((s) => s.value === draft.source)?.value ?? 'wind');
  const subsystem = $derived(subsystems.find((s) => s.value === draft.subsystem)?.value);
  const threshold = $derived.by(() => {
    const n = Number(draft.threshold.replace(/[\s,]/g, ''));
    const usable =
      draft.threshold.trim() !== '' &&
      Number.isInteger(n) &&
      n >= limits.thresholdMin &&
      n <= limits.thresholdMax;
    return usable ? n : undefined;
  });

  const full = $derived(data.held >= limits.perAccount);
  const ids = {
    name: 'alert-name',
    source: 'alert-source',
    subsystem: 'alert-subsystem',
    thresholdMwh: 'alert-threshold',
  } as const;
  const problems = $derived(
    alertFields.flatMap((field) => {
      const text = alertError(field, form?.create.errors[field]);
      return text === undefined ? [] : [{ href: `#${ids[field]}`, text }];
    }),
  );
</script>

<svelte:head>
  <title>New Alert | Spillway</title>
</svelte:head>

<PageHeader
  title="New Alert"
  description="An alert fires on each day its curtailed energy goes over the threshold."
/>

<div class="mx-auto flex max-w-[83.25rem] flex-col gap-600 px-600 py-700">
  <!--
    Side by side, the two cards are as tall as each other: the preview's chart
    grows to where the form ends, whatever the form shows.
  -->
  <div class="grid gap-600 laptop:grid-cols-[minmax(0,5fr)_minmax(0,7fr)]">
    <Card
      title="What to Watch"
      description="Choose a source and a subsystem, and how much curtailed energy in one day is too much."
    >
      {#if full || form?.create.full}
        <ErrorSummary
          title="You have {limits.perAccount} alerts, the most an account can keep"
          detail="Delete one on the alerts page to make room for another."
        />
      {:else if problems.length > 0}
        <ErrorSummary title="Check the alert" items={problems} />
      {/if}

      <form
        method="POST"
        action="?/create"
        novalidate
        use:enhance={create}
        oninput={(e) => {
          read(e.currentTarget);
        }}
        onchange={(e) => {
          read(e.currentTarget);
        }}
        class="flex flex-col gap-700 pt-300"
      >
        <AlertFields values={form?.create.values} errors={form?.create.errors} />
        <div class="flex justify-end gap-300">
          <a href={resolve('/alerts')} class={buttonStyles({ variant: 'secondary' })}>Cancel</a>
          <Button
            type="submit"
            icon={NotificationAdd}
            disabled={full}
            aria-disabled={creating || undefined}
          >
            {creating ? 'Creating Alert…' : 'Create Alert'}
          </Button>
        </div>
      </form>
    </Card>

    <Card
      title={`The Last ${data.recent?.series[0]?.curtailedMwh.length ?? 30} Days`}
      description={subsystem === undefined
        ? 'What the alert would have caught, day by day.'
        : `${sources.find((s) => s.value === source)?.label} in the ${subsystemLabel(subsystem)}, day by day, against your threshold.`}
    >
      {#if data.recent === null}
        <div class="flex flex-1 items-center justify-center">
          <EmptyState
            icon={Downloading}
            title="Nothing to preview yet"
            description="The ONS's data is still loading. The alert works all the same."
          />
        </div>
      {:else}
        <Preview recent={data.recent} {source} {subsystem} {threshold} />
      {/if}
    </Card>
  </div>
</div>
