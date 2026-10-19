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
  import { int } from '#lib/features/curtailment/format.ts';
  import { enhanceForm } from '#lib/forms/enhance.ts';
  import { localized } from '#lib/i18n.ts';
  import { m } from '#lib/paraglide/messages.js';

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
  <title>{m.new_alert_head()} | Spillway</title>
</svelte:head>

<PageHeader title={m.new_alert_title()} description={m.new_alert_description()} />

<div class="mx-auto flex max-w-[83.25rem] flex-col gap-600 px-600 py-700">
  <!--
    Side by side, the two cards are as tall as each other: the preview's chart
    grows to where the form ends, whatever the form shows.
  -->
  <div class="grid gap-600 laptop:grid-cols-[minmax(0,5fr)_minmax(0,7fr)]">
    <Card title={m.new_alert_watch()} description={m.new_alert_watch_description()}>
      {#if full || form?.create.full}
        <ErrorSummary
          title={m.new_alert_full({ count: limits.perAccount })}
          detail={m.new_alert_full_detail()}
        />
      {:else if problems.length > 0}
        <ErrorSummary title={m.new_alert_check()} items={problems} />
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
          <a href={localized(resolve('/alerts'))} class={buttonStyles({ variant: 'secondary' })}
            >{m.new_alert_cancel()}</a
          >
          <Button
            type="submit"
            icon={NotificationAdd}
            disabled={full}
            aria-disabled={creating || undefined}
          >
            {creating ? m.new_alert_creating() : m.new_alert_create()}
          </Button>
        </div>
      </form>
    </Card>

    <Card
      title={m.new_alert_last_days({
        days: int(data.recent?.series[0]?.curtailedMwh.length ?? 30),
      })}
      description={subsystem === undefined
        ? m.new_alert_last_days_description()
        : m.new_alert_source_in({
            source: sources.find((s) => s.value === source)?.label ?? '',
            subsystem: subsystemLabel(subsystem),
          })}
    >
      {#if data.recent === null}
        <div class="flex flex-1 items-center justify-center">
          <EmptyState
            icon={Downloading}
            title={m.new_alert_nothing()}
            description={m.new_alert_nothing_body()}
          />
        </div>
      {:else}
        <Preview recent={data.recent} {source} {subsystem} {threshold} />
      {/if}
    </Card>
  </div>
</div>
