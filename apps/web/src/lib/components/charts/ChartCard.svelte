<script lang="ts">
  import type { Snippet } from 'svelte';
  import BarChart from '~icons/material-symbols/bar-chart-rounded';
  import TableIcon from '~icons/material-symbols/table-outline-rounded';

  import Card from '#lib/components/ui/Card.svelte';
  import SegmentedControl from '#lib/components/ui/SegmentedControl.svelte';
  import { m } from '#lib/paraglide/messages.js';

  type Props = {
    title: string;
    description?: string | undefined;
    /** The chart's key, shown with the chart only. */
    legend?: Snippet;
    /** The chart. */
    children: Snippet;
    /** The chart's figures: what a screen reader reads, and anyone can copy. */
    table: Snippet;
  };

  let { title, description, legend, children, table }: Props = $props();
  const id = $props.id();
</script>

<!--
  A card that shows its chart or the chart's figures as a table, as the switch
  in its header says. The switch is two radio buttons and the swap is CSS, from
  the wrapper that holds both, so it works without JavaScript; the view not
  chosen is not displayed, so a screen reader skips it too.
-->
<div class="group/view contents">
  <Card {title} {description}>
    {#snippet aside()}
      <SegmentedControl
        legend={m.chart_show_as()}
        hideLegend
        size="sm"
        name={`${id}-view`}
        value="chart"
        options={[
          { value: 'chart', label: m.chart_chart(), icon: BarChart },
          { value: 'table', label: m.chart_table(), icon: TableIcon },
        ]}
      />
    {/snippet}
    <!-- The key goes with the chart, over it, and leaves the header to the switch. -->
    <div class="flex flex-col gap-500 group-has-[input[value=table]:checked]/view:hidden">
      {@render legend?.()}
      {@render children()}
    </div>
    <!-- As tall as a chart at most, so the card keeps its size: a long table scrolls. -->
    <div
      class="hidden group-has-[input[value=table]:checked]/view:block [&_[data-table-frame]]:max-h-[20rem]"
    >
      {@render table()}
    </div>
  </Card>
</div>
