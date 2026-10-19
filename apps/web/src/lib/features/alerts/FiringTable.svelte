<script lang="ts">
  import type { Firing } from '#lib/api/generated/types.gen.ts';
  import Table from '#lib/components/ui/Table.svelte';
  import { longDay, mwh, pct, shortDay } from '#lib/features/curtailment/format.ts';
  import { m } from '#lib/paraglide/messages.js';

  type Props = { firings: readonly Firing[] };

  let { firings }: Props = $props();

  // On a phone a firing is its day and its alert, so the table fits and never
  // has to scroll; what was lost, the threshold and the margin are from mobile up.
  const wide = 'hidden mobile:table-cell';
</script>

<Table
  caption={m.firings_caption()}
  columns={[
    { label: m.firings_day() },
    { label: m.firings_alert() },
    { label: m.firings_curtailed(), right: true, class: wide },
    { label: m.firings_threshold(), right: true, class: wide },
    { label: m.firings_over_by(), right: true, class: wide },
  ]}
  minWidth="mobile:min-w-[40rem]"
>
  {#each firings as f (`${f.day}-${f.alertId}`)}
    <tr class="ts-primary-body-m text-text-primary-default even:bg-surface-zebra">
      <th
        scope="row"
        class="h-[2.25rem] px-300 text-left ts-primary-subheading-s whitespace-nowrap"
      >
        <time datetime={f.day}>
          <span class="mobile:hidden">{shortDay(f.day)}</span>
          <span class="hidden mobile:inline">{longDay(f.day)}</span>
        </time>
      </th>
      <td class="max-w-[10rem] truncate px-300 mobile:max-w-[16rem]">{f.alertName}</td>
      <td class={['px-300 text-right whitespace-nowrap tabular-nums', wide]}>
        {mwh(f.curtailedMwh)}
      </td>
      <td class={['px-300 text-right text-text-secondary-default tabular-nums', wide]}>
        {mwh(f.thresholdMwh)}
      </td>
      <td class={['px-300 text-right tabular-nums', wide]}>
        {pct(f.curtailedMwh / f.thresholdMwh - 1)}
      </td>
    </tr>
  {/each}
</Table>
