<script lang="ts">
  import type { Snippet } from 'svelte';

  import type { Icon as IconComponent } from './icon.ts';

  type Props = {
    title: string;
    description: string;
    headingLevel?: 2 | 3;
    children?: Snippet;
  } & (
    | {
        /** The illustration, drawn for the place it is shown in. */
        art: Snippet;
        icon?: never;
      }
    | {
        /** In place of an illustration, where one would be too much, as inside a card. */
        icon: IconComponent;
        art?: never;
      }
  );

  let { title, description, headingLevel = 3, art, icon: Icon, children }: Props = $props();
</script>

<!-- An illustrated message, small size, with a heading that fits its place. -->
<div class="flex flex-col items-center gap-500 px-600 py-700 text-center">
  {#if Icon}
    <div
      aria-hidden="true"
      class="flex size-[2.5rem] items-center justify-center rounded-200 bg-surface-container-secondary text-icon-secondary"
    >
      <Icon width={20} height={20} />
    </div>
  {:else}
    <div aria-hidden="true" class="w-[9.875rem]">{@render art?.()}</div>
  {/if}
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
