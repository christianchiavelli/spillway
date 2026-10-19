<script lang="ts">
  import type { Snippet } from 'svelte';

  type Column = {
    label: string;
    /** Numbers line up on the right. */
    right?: boolean;
    /** More classes, such as the ones that hide a column on a phone. */
    class?: string;
    /** A heading only a screen reader needs, such as Actions. */
    hidden?: boolean;
  };

  type Props = {
    caption: string;
    columns: readonly Column[];
    /** The table's least width, as a class: below it the frame scrolls sideways. */
    minWidth: string;
    /** The rows, in the table's row and cell classes. */
    children: Snippet;
  };

  let { caption, columns, minWidth, children }: Props = $props();
</script>

<!--
  The table: caption-style headers, 36px rows, zebra striping, and numbers on
  the right in tabular figures. The frame scrolls sideways on a phone, and is
  relative so the visually hidden caption scrolls with it instead of widening
  the page.
-->
<div class="relative overflow-x-auto rounded-300 border border-stroke-neutral-default">
  <table class={['w-full border-collapse', minWidth]}>
    <caption class="sr-only">{caption}</caption>
    <thead>
      <tr class="border-b border-stroke-neutral-default bg-surface-container-primary">
        {#each columns as column, i (i)}
          <th
            scope="col"
            class={[
              'h-[2.25rem] px-300 ts-primary-caption text-text-secondary-default',
              column.right ? 'text-right' : 'text-left',
              column.class,
            ]}
          >
            {#if column.hidden}<span class="sr-only">{column.label}</span>{:else}{column.label}{/if}
          </th>
        {/each}
      </tr>
    </thead>
    <tbody>
      {@render children()}
    </tbody>
  </table>
</div>
