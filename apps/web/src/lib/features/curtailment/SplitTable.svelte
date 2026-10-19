<script lang="ts">
  import Table from '#lib/components/ui/Table.svelte';

  import { energy } from './format.ts';

  type Row = { key: string; label: string; windMwh: number; solarMwh: number };

  type Props = {
    caption: string;
    /** The first column's heading: what each row is. */
    heading: string;
    rows: readonly Row[];
  };

  let { caption, heading, rows }: Props = $props();
</script>

<!-- Each row's wind, solar and total, the figures of a chart that stacks them. -->
<Table
  {caption}
  columns={[
    { label: heading },
    { label: 'Wind', right: true },
    { label: 'Solar', right: true },
    { label: 'Total', right: true },
  ]}
  minWidth="mobile:min-w-[24rem]"
>
  {#each rows as row (row.key)}
    <tr class="ts-primary-body-m text-text-primary-default even:bg-surface-zebra">
      <th scope="row" class="h-[2.25rem] px-300 text-left ts-primary-subheading-s">{row.label}</th>
      <td class="px-300 text-right tabular-nums">{energy(row.windMwh)}</td>
      <td class="px-300 text-right tabular-nums">{energy(row.solarMwh)}</td>
      <td class="px-300 text-right tabular-nums">{energy(row.windMwh + row.solarMwh)}</td>
    </tr>
  {/each}
</Table>
