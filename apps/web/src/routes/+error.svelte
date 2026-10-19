<script lang="ts">
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import TurbinesArt from '#lib/components/art/TurbinesArt.svelte';
  import { buttonStyles } from '#lib/components/ui/button-styles.ts';
  import EmptyState from '#lib/components/ui/EmptyState.svelte';
  import { localized } from '#lib/i18n.ts';
  import { m } from '#lib/paraglide/messages.js';

  const notFound = $derived(page.status === 404);
  const title = $derived(notFound ? m.error_title_not_found() : m.error_title_failed());
</script>

<svelte:head>
  <title>{notFound ? m.error_head_not_found() : m.error_head_failed()} | Spillway</title>
</svelte:head>

<div class="mx-auto flex max-w-[40rem] flex-col items-center px-600 py-1200">
  <EmptyState
    {title}
    headingLevel={2}
    description={notFound ? m.error_check_address() : (page.error?.message ?? m.error_try_again())}
  >
    {#snippet art()}
      <TurbinesArt />
    {/snippet}
    {#if page.error?.reference}
      <p class="ts-primary-body-s text-text-secondary-default">
        {m.error_reference()} <code class="ts-primary-subheading-xs">{page.error.reference}</code>
      </p>
    {/if}
    <a href={localized(resolve('/alerts'))} class={buttonStyles({ variant: 'secondary' })}
      >{m.error_to_alerts()}</a
    >
  </EmptyState>
  <h1 class="sr-only">{title}</h1>
</div>
