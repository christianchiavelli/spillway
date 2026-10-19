<script lang="ts">
  import '@fontsource-variable/open-sans';
  import '@fontsource-variable/lora';
  import '../app.css';

  import { page } from '$app/state';
  import AppShell from '#lib/components/layout/AppShell.svelte';
  import Toaster, { type Toast } from '#lib/components/ui/Toaster.svelte';
  import { editions, inEdition } from '#lib/i18n.ts';
  import { m } from '#lib/paraglide/messages.js';

  import type { LayoutProps } from './$types';

  let { data, children }: LayoutProps = $props();

  let dismissed = $state<readonly string[]>([]);

  const toasts = $derived.by((): Toast[] => {
    const flash = data.flash;
    if (flash === null || dismissed.includes(flash.id)) return [];
    switch (flash.kind) {
      case 'alert-created':
        return [
          {
            id: flash.id,
            title: m.toast_alert_created(),
            body: m.toast_alert_created_body({ name: flash.name }),
          },
        ];
      case 'alert-deleted':
        return [
          {
            id: flash.id,
            title: m.toast_alert_deleted(),
            body: m.toast_alert_deleted_body({ name: flash.name }),
          },
        ];
      case 'account-created':
        return [{ id: flash.id, title: m.toast_welcome(), body: m.toast_welcome_body() }];
    }
  });
</script>

<!--
  Each page names its other edition, so a search engine shows the one in the
  reader's language. English is also the default, for every other language.
-->
<svelte:head>
  {#each editions as edition (edition.locale)}
    <link
      rel="alternate"
      hreflang={edition.locale}
      href={page.url.origin + inEdition(page.url, edition.locale)}
    />
  {/each}
  <link rel="alternate" hreflang="x-default" href={page.url.origin + inEdition(page.url, 'en')} />
</svelte:head>

<AppShell email={data.email}>
  {@render children()}
</AppShell>

<Toaster
  {toasts}
  ondismiss={(id) => {
    dismissed = [...dismissed, id];
  }}
/>
