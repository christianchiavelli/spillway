<script lang="ts">
  import '@fontsource-variable/open-sans';
  import '@fontsource-variable/lora';
  import '../app.css';

  import { setLucideProps } from '@lucide/svelte';

  import AppShell from '#lib/components/layout/AppShell.svelte';
  import Toaster, { type Toast } from '#lib/components/ui/Toaster.svelte';

  import type { LayoutProps } from './$types';

  let { data, children }: LayoutProps = $props();

  // Thin strokes everywhere: a light icon weight suits a dense dashboard.
  setLucideProps({ strokeWidth: 1.5 });

  let dismissed = $state<readonly string[]>([]);

  const toasts = $derived.by((): Toast[] => {
    const flash = data.flash;
    if (flash === null || dismissed.includes(flash.id)) return [];
    switch (flash.kind) {
      case 'alert-created':
        return [{ id: flash.id, title: 'Alert created', body: `“${flash.name}” is on your list.` }];
      case 'alert-deleted':
        return [
          { id: flash.id, title: 'Alert deleted', body: `“${flash.name}” is off your list.` },
        ];
      case 'account-created':
        return [{ id: flash.id, title: 'Welcome to Spillway', body: 'Your account is ready.' }];
    }
  });
</script>

<AppShell email={data.email}>
  {@render children()}
</AppShell>

<Toaster
  {toasts}
  ondismiss={(id) => {
    dismissed = [...dismissed, id];
  }}
/>
