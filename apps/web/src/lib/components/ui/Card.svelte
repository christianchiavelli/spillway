<script lang="ts">
  import type { Snippet } from 'svelte';

  type Props = {
    title: string;
    description?: string | undefined;
    headingLevel?: 2 | 3;
    /** Controls beside the title, such as a count or a button. */
    aside?: Snippet;
    children: Snippet;
  };

  let { title, description, headingLevel = 2, aside, children }: Props = $props();
  const id = $props.id();
</script>

<!-- The content wrapper and its header: a panel on the canvas, lighter in both themes. -->
<section
  aria-labelledby={`${id}-title`}
  class="flex flex-col gap-600 rounded-300 border border-stroke-neutral-default bg-surface-container-primary p-600"
>
  <!--
    The title takes the room the aside leaves, so its description wraps first
    and the aside drops under it only when the title would get narrower than
    15rem.
  -->
  <div class="flex flex-wrap items-start justify-between gap-500">
    <div class="flex min-w-[min(100%,15rem)] flex-1 flex-col gap-100">
      <svelte:element
        this={`h${headingLevel}`}
        id={`${id}-title`}
        class="ts-primary-subheading-m text-text-primary-default"
      >
        {title}
      </svelte:element>
      {#if description}
        <p class="max-w-[60ch] ts-primary-body-m text-pretty text-text-secondary-default">
          {description}
        </p>
      {/if}
    </div>
    {@render aside?.()}
  </div>
  {@render children()}
</section>
