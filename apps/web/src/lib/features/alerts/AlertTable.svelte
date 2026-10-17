<script lang="ts">
  import { Trash } from '@lucide/svelte';

  import type { SubmitFunction } from '$app/forms';
  import { enhance } from '$app/forms';
  import type { Alert } from '#lib/api/generated/types.gen.ts';
  import Button from '#lib/components/ui/Button.svelte';

  import { formatDay, formatMWh, sourceOf, subsystemLabel } from './alerts.ts';

  type Props = {
    alerts: readonly Alert[];
    /** What happens when a delete form is sent; the page decides. */
    ondelete: SubmitFunction;
    /** The alert being deleted right now, if any. */
    deleting: string | null;
  };

  let { alerts, ondelete, deleting }: Props = $props();
</script>

<!--
  The table: caption-style headers, 36px rows, zebra striping and
  numbers aligned right, here in tabular figures so they line up digit by
  digit. Each row's name is its header, so a screen reader announces it with
  every cell, the delete button included. The wrapper scrolls the table
  sideways on a phone, and is relative so the visually hidden text inside,
  which is absolutely positioned, scrolls with it instead of widening the page.
-->
<div class="relative overflow-x-auto rounded-300 border border-stroke-neutral-default">
  <table class="w-full min-w-[40rem] border-collapse">
    <caption class="sr-only">Your alerts, newest first</caption>
    <thead>
      <tr class="border-b border-stroke-neutral-default bg-surface-container-primary">
        <th
          scope="col"
          class="h-[2.25rem] px-300 text-left ts-primary-caption text-text-secondary-default"
          >Name</th
        >
        <th scope="col" class="px-300 text-left ts-primary-caption text-text-secondary-default"
          >Source</th
        >
        <th scope="col" class="px-300 text-left ts-primary-caption text-text-secondary-default"
          >Subsystem</th
        >
        <th scope="col" class="px-300 text-right ts-primary-caption text-text-secondary-default"
          >Daily Threshold</th
        >
        <th scope="col" class="px-300 text-left ts-primary-caption text-text-secondary-default"
          >Created</th
        >
        <th scope="col" class="px-300"><span class="sr-only">Actions</span></th>
      </tr>
    </thead>
    <tbody>
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
              <source.icon size={16} class="text-icon-secondary" />
              {source.label}
            </span>
          </td>
          <td class="px-300">{subsystemLabel(alert.subsystem)}</td>
          <td class="px-300 text-right tabular-nums">{formatMWh(alert.thresholdMwh)}</td>
          <td class="px-300 whitespace-nowrap">
            <time datetime={alert.createdAt}>{formatDay(alert.createdAt)}</time>
          </td>
          <td class="px-200 py-100 text-right">
            <form method="POST" action="?/delete" use:enhance={ondelete}>
              <input type="hidden" name="id" value={alert.id} />
              <input type="hidden" name="name" value={alert.name} />
              <Button
                type="submit"
                variant="ghost"
                size="sm"
                icon={Trash}
                aria-disabled={deleting === alert.id || undefined}
              >
                {deleting === alert.id ? 'Deleting…' : 'Delete'}
                <span class="sr-only">“{alert.name}”</span>
              </Button>
            </form>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
