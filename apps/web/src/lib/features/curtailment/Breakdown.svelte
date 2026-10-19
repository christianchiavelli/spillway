<script lang="ts">
  import type { Curtailment } from '#lib/api/generated/types.gen.ts';
  import BarRows from '#lib/components/charts/BarRows.svelte';
  import Card from '#lib/components/ui/Card.svelte';

  import { energy } from './format.ts';
  import { reasons, sources, subsystemLabel } from './grid.ts';

  type Props = { overview: Curtailment };

  let { overview }: Props = $props();

  const rows = <T,>(
    items: readonly T[],
    label: (item: T) => string,
    values: (item: T) => [number, number],
  ) =>
    items
      .map((item) => {
        const v = values(item);
        return { label: label(item), values: v, total: v[0] + v[1] };
      })
      .sort((a, b) => b.total - a.total)
      .map((r) => ({ ...r, value: energy(r.total) }));

  const bySubsystem = $derived(
    rows(
      overview.bySubsystem,
      (s) => subsystemLabel(s.subsystem),
      (s) => [s.windMwh, s.solarMwh],
    ),
  );
  const byReason = $derived(
    rows(
      overview.byReason,
      (r) => reasons[r.reason].label,
      (r) => [r.windMwh, r.solarMwh],
    ),
  );
  const given = $derived(
    [...overview.byReason]
      .sort((a, b) => b.windMwh + b.solarMwh - (a.windMwh + a.solarMwh))
      .map((r) => reasons[r.reason]),
  );
</script>

<div class="grid gap-600 laptop:grid-cols-2">
  <Card
    title="By Subsystem"
    description="The national grid is run as four regions, its subsystems."
  >
    <BarRows title="Curtailed energy by subsystem" series={sources} rows={bySubsystem} />
  </Card>
  <Card title="By Reason" description="The reason the ONS gave when it limited a plant.">
    {#if byReason.length === 0}
      <p class="ts-primary-body-m text-text-secondary-default">
        The ONS limited no plant in this period.
      </p>
    {:else}
      <BarRows title="Curtailed energy by reason" series={sources} rows={byReason} />
      <dl class="grid gap-x-400 gap-y-200 ts-primary-body-s laptop:grid-cols-[auto_1fr]">
        {#each given as reason (reason.label)}
          <dt class="ts-primary-subheading-xs text-text-primary-default">{reason.label}</dt>
          <dd class="text-text-secondary-default">{reason.meaning}</dd>
        {/each}
      </dl>
    {/if}
  </Card>
</div>
