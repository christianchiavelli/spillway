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
  const id = $props.id();
</script>

<!--
  The table: caption-style headers, 36px rows, zebra striping, and numbers on
  the right in tabular figures. The frame scrolls sideways on a phone, and is
  relative so the visually hidden caption scrolls with it instead of widening
  the page. Where a card caps its height, it scrolls down too, under headers
  that stay: separate borders, since a collapsed one scrolls away from a
  sticky cell. A frame that scrolls has to take focus, or the keyboard could
  not scroll it (WCAG 2.1.1): it is a region named by the caption, in the tab
  order, which Svelte warns a region should not be.
-->
<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
  data-table-frame
  role="region"
  aria-labelledby={`${id}-caption`}
  tabindex="0"
  class="relative overflow-auto rounded-300 border border-stroke-neutral-default"
>
  <table class={['w-full border-separate border-spacing-0', minWidth]}>
    <caption id={`${id}-caption`} class="sr-only">{caption}</caption>
    <thead>
      <tr>
        {#each columns as column, i (i)}
          <th
            scope="col"
            class={[
              'sticky top-0 z-[1] h-[2.25rem] border-b border-stroke-neutral-default bg-surface-container-primary px-300 ts-primary-caption text-text-secondary-default',
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
