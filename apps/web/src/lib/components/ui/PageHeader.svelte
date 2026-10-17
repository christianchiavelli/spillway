<script lang="ts">
  import type { Snippet } from 'svelte';

  type Props = {
    title: string;
    description?: string | undefined;
    actions?: Snippet;
  };

  let { title, description, actions }: Props = $props();
</script>

<!-- The page header: a white band with the page title in Lora. -->
<header class="border-b border-stroke-neutral-default bg-surface-container-primary">
  <div
    class="mx-auto flex max-w-[83.25rem] flex-wrap items-center justify-between gap-500 px-600 py-500"
  >
    <div class="flex flex-col gap-100">
      <h1 class="ts-secondary-heading-s text-text-primary-default">{title}</h1>
      {#if description}
        <p class="max-w-[72ch] ts-primary-body-m text-text-secondary-default">{description}</p>
      {/if}
    </div>
    {@render actions?.()}
  </div>
</header>
