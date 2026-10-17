<script lang="ts">
  import type { Snippet } from 'svelte';

  type Props = {
    title: string;
    description: string;
    headingLevel?: 2 | 3;
    /** The illustration, drawn for the place it is shown in. */
    art: Snippet;
    children?: Snippet;
  };

  let { title, description, headingLevel = 3, art, children }: Props = $props();
</script>

<!-- An illustrated message, small size, with a heading that fits its place. -->
<div class="flex flex-col items-center gap-500 px-600 py-700 text-center">
  <div aria-hidden="true" class="w-[9.875rem]">{@render art()}</div>
  <div class="flex flex-col gap-300">
    <svelte:element
      this={`h${headingLevel}`}
      class="ts-primary-subheading-s text-text-primary-default"
    >
      {title}
    </svelte:element>
    <p class="max-w-[44ch] ts-primary-body-m text-text-secondary-default">{description}</p>
  </div>
  {@render children?.()}
</div>
