<script lang="ts">
  import '@fontsource-variable/open-sans';
  import '@fontsource-variable/lora';
  import '../app.css';

  import type { Attachment } from 'svelte/attachments';

  import { page } from '$app/state';
  import AppShell from '#lib/components/layout/AppShell.svelte';
  import Toaster, { type Toast } from '#lib/components/ui/Toaster.svelte';
  import { editions, inEdition } from '#lib/i18n.ts';
  import { m } from '#lib/paraglide/messages.js';
  import { applyTheme, colorScheme, type Theme, themeColors } from '#lib/theme.ts';

  import type { LayoutProps } from './$types';

  let { data, children }: LayoutProps = $props();

  // The theme the server rendered, until the menu picks another: then that
  // one, at once, while the server stores it. Data loaded again brings the
  // cookie's, which another tab may have changed.
  let theme = $derived(data.theme);
  const pickTheme = (picked: Theme) => {
    theme = picked;
  };

  // <html> wears it: written by the server for the first paint, and from here
  // whenever it changes.
  const wear: Attachment<Document> = (document) => {
    applyTheme(theme, document.documentElement);
  };

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

<svelte:document {@attach wear} />

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
  <!--
    The scheme the page is drawn in, for what the browser draws itself, and
    the header's navy for the bar a phone's browser draws above it.
  -->
  <meta name="color-scheme" content={colorScheme(theme)} />
  {#each themeColors(theme) as color (color.media ?? 'any')}
    <meta name="theme-color" content={color.content} media={color.media} />
  {/each}
</svelte:head>

<AppShell email={data.email} {theme} onthemechange={pickTheme}>
  {@render children()}
</AppShell>

<Toaster
  {toasts}
  ondismiss={(id) => {
    dismissed = [...dismissed, id];
  }}
/>
