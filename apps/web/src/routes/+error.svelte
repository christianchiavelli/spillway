<script lang="ts">
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { buttonStyles } from '#lib/components/ui/button-styles.ts';
  import EmptyState from '#lib/components/ui/EmptyState.svelte';
  import TurbinesArt from '#lib/features/alerts/TurbinesArt.svelte';

  const notFound = $derived(page.status === 404);
  const title = $derived(notFound ? 'This page does not exist' : 'Something went wrong');
</script>

<svelte:head>
  <title>{notFound ? 'Not Found' : 'Error'} | Spillway</title>
</svelte:head>

<div class="mx-auto flex max-w-[40rem] flex-col items-center px-600 py-1200">
  <EmptyState
    {title}
    headingLevel={2}
    description={notFound
      ? 'Check the address, or go back to your alerts.'
      : (page.error?.message ?? 'Try again in a moment.')}
  >
    {#snippet art()}
      <TurbinesArt />
    {/snippet}
    {#if page.error?.reference}
      <p class="ts-primary-body-s text-text-secondary-default">
        Reference <code class="ts-primary-subheading-xs">{page.error.reference}</code>
      </p>
    {/if}
    <a href={resolve('/alerts')} class={buttonStyles({ variant: 'secondary' })}>Go to Your Alerts</a
    >
  </EmptyState>
  <h1 class="sr-only">{title}</h1>
</div>
