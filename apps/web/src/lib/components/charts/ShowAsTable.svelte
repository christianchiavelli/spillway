<script lang="ts">
  import type { Snippet } from 'svelte';
  import TableIcon from '~icons/material-symbols/table-outline-rounded';

  type Props = { children: Snippet };

  let { children }: Props = $props();
</script>

<!--
  The chart's figures as a table, closed until asked for: what a screen reader
  reads in place of the drawing, and what anyone can copy. A native details
  element, so it opens without JavaScript.
-->
<details>
  <summary
    class="inline-flex cursor-pointer items-center gap-200 rounded-200 ts-primary-button-s text-text-primary-link hover:text-text-primary-link-hover"
  >
    <TableIcon width={16} height={16} />
    Show as table
  </summary>
  <div class="pt-400">
    {@render children()}
  </div>
</details>
