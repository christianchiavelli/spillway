<script lang="ts">
  import type { PlantCurtailment } from '#lib/api/generated/types.gen.ts';
  import DataBar from '#lib/components/charts/DataBar.svelte';
  import Table from '#lib/components/ui/Table.svelte';
  import { m } from '#lib/paraglide/messages.js';

  import { int, mwh, pct, shareOf } from './format.ts';
  import { sourceOf, subsystemLabel } from './grid.ts';

  type Props = {
    plants: readonly PlantCurtailment[];
    caption: string;
    /** The rank of the first row, so a later page goes on counting. */
    first?: number;
    /** The plants page shows the subsystem and the time limited too. */
    full?: boolean;
  };

  let { plants, caption, first = 1, full = false }: Props = $props();

  const max = $derived(Math.max(0, ...plants.map((p) => p.curtailedMwh)));

  /** "36 h", or minutes under an hour. */
  const limited = (minutes: number) =>
    minutes < 60 ? m.plants_minutes({ n: int(minutes) }) : m.plants_hours({ n: int(minutes / 60) });

  // On a phone a row is the plant and what it lost; the rest is from mobile up,
  // but the subsystem, the longest name, from laptop up: below it the table
  // would scroll sideways on a tablet, and the state already says the region.
  const wide = 'hidden mobile:table-cell';
  const widest = 'hidden laptop:table-cell';
</script>

<Table
  {caption}
  columns={[
    { label: m.plants_rank(), class: wide },
    { label: m.plants_plant() },
    { label: m.plants_state(), class: wide },
    ...(full ? [{ label: m.plants_subsystem(), class: widest }] : []),
    { label: m.plants_source(), class: wide },
    { label: m.plants_curtailed(), right: true },
    { label: m.plants_share(), right: true, class: wide },
    ...(full ? [{ label: m.plants_time(), right: true, class: wide }] : []),
  ]}
  minWidth={full ? 'mobile:min-w-[48rem] laptop:min-w-[60rem]' : 'mobile:min-w-[44rem]'}
>
  {#each plants as plant, i (plant.id)}
    {@const source = sourceOf(plant.source)}
    <tr class="ts-primary-body-m text-text-primary-default even:bg-surface-zebra">
      <td class={['w-[2.5rem] px-300 text-text-secondary-default tabular-nums', wide]}
        >{first + i}</td
      >
      <th
        scope="row"
        class="h-[2.25rem] max-w-[7.5rem] truncate px-300 text-left ts-primary-subheading-s mobile:max-w-[18rem]"
      >
        {plant.name}
      </th>
      <td class={['px-300', wide]}>{plant.state}</td>
      {#if full}
        <!-- Never broken at Centre-West's hyphen. -->
        <td class={['px-300 whitespace-nowrap', widest]}>{subsystemLabel(plant.subsystem)}</td>
      {/if}
      <td class={['px-300', wide]}>
        <span class="inline-flex items-center gap-200">
          <source.icon width={16} height={16} class="text-icon-secondary" />
          {source.label}
        </span>
      </td>
      <td class="px-300">
        <span class="flex items-center justify-end gap-300">
          <span class="hidden mobile:contents">
            <DataBar value={plant.curtailedMwh} {max} fill={source.fill} />
          </span>
          <span class="text-right whitespace-nowrap tabular-nums mobile:w-[5.5rem]"
            >{mwh(plant.curtailedMwh)}</span
          >
        </span>
      </td>
      <td class={['px-300 text-right tabular-nums', wide]}>{pct(shareOf(plant))}</td>
      {#if full}
        <td class={['px-300 text-right tabular-nums', wide]}>{limited(plant.limitedMinutes)}</td>
      {/if}
    </tr>
  {/each}
</Table>
