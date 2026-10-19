<script lang="ts">
  import Delete from '~icons/material-symbols/delete-outline-rounded';

  import type { SubmitFunction } from '$app/forms';
  import { enhance } from '$app/forms';
  import type { WatchedAlert } from '#lib/api/generated/types.gen.ts';
  import Button from '#lib/components/ui/Button.svelte';
  import Table from '#lib/components/ui/Table.svelte';
  import { int, shortDayYear } from '#lib/features/curtailment/format.ts';
  import { sourceOf, subsystemLabel } from '#lib/features/curtailment/grid.ts';
  import { m } from '#lib/paraglide/messages.js';

  import { formatMWh } from './alerts.ts';

  type Props = {
    alerts: readonly WatchedAlert[];
    /** What happens when a delete form is sent; the page decides. */
    ondelete: SubmitFunction;
    /** The alert being deleted right now, if any. */
    deleting: string | null;
  };

  let { alerts, ondelete, deleting }: Props = $props();
</script>

<!--
  Each row's name is its header, so a screen reader announces it with every
  cell, the delete button included.
-->
<Table
  caption={m.alerts_table_caption()}
  columns={[
    { label: m.alerts_table_name() },
    { label: m.alerts_table_source() },
    { label: m.alerts_table_subsystem() },
    { label: m.alerts_table_threshold(), right: true },
    { label: m.alerts_table_last_fired() },
    { label: m.alerts_table_days_fired(), right: true },
    { label: m.alerts_table_actions(), hidden: true },
  ]}
  minWidth="min-w-[52rem]"
>
  {#each alerts as alert (alert.id)}
    {@const source = sourceOf(alert.source)}
    <tr class="ts-primary-body-m text-text-primary-default even:bg-surface-zebra">
      <th
        scope="row"
        class="h-[2.25rem] max-w-[18rem] truncate px-300 text-left ts-primary-subheading-s"
      >
        {alert.name}
      </th>
      <td class="px-300">
        <span class="inline-flex items-center gap-200">
          <source.icon width={16} height={16} class="text-icon-secondary" />
          {source.label}
        </span>
      </td>
      <td class="px-300">{subsystemLabel(alert.subsystem)}</td>
      <td class="px-300 text-right tabular-nums">{formatMWh(alert.thresholdMwh)}</td>
      <td class="px-300 whitespace-nowrap">
        {#if alert.lastFiredOn === undefined}
          <span class="text-text-secondary-default">{m.alerts_never()}</span>
        {:else}
          <time datetime={alert.lastFiredOn}>{shortDayYear(alert.lastFiredOn)}</time>
        {/if}
      </td>
      <td class="px-300 text-right tabular-nums">
        {alert.daysFired === 0
          ? m.alerts_none()
          : m.alerts_days({ count: alert.daysFired, days: int(alert.daysFired) })}
      </td>
      <td class="px-200 py-100 text-right">
        <form method="POST" action="?/delete" use:enhance={ondelete}>
          <input type="hidden" name="id" value={alert.id} />
          <input type="hidden" name="name" value={alert.name} />
          <Button
            type="submit"
            variant="ghost"
            size="sm"
            icon={Delete}
            aria-disabled={deleting === alert.id || undefined}
          >
            {deleting === alert.id ? m.alerts_deleting() : m.alerts_delete()}
            <span class="sr-only">“{alert.name}”</span>
          </Button>
        </form>
      </td>
    </tr>
  {/each}
</Table>
