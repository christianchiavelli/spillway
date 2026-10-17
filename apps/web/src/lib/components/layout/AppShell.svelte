<script lang="ts">
  import { Bell, LogOut } from '@lucide/svelte';
  import type { Snippet } from 'svelte';

  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { buttonStyles } from '#lib/components/ui/button-styles.ts';

  import Logo from './Logo.svelte';

  type Props = {
    /** The signed-in address, or null for a visitor. */
    email: string | null;
    children: Snippet;
  };

  let { email, children }: Props = $props();

  // Only pages that exist get a link: a visitor sees no navigation at all.
  const nav = [{ route: '/alerts', label: 'Alerts', icon: Bell }] as const;
</script>

<a
  href="#main"
  class="sr-only z-(--z-toast) rounded-200 bg-surface-container-primary px-400 py-300 ts-primary-button-l text-text-primary-default focus:not-sr-only focus:fixed focus:top-300 focus:left-300"
>
  Skip to content
</a>

<!-- The chrome: a dark header across the top and a dark rail on the left. -->
<header
  class="sticky top-0 z-(--z-header) flex h-[3.5rem] items-center gap-600 bg-surface-primary-cool-dark px-400"
>
  <a href={email === null ? resolve('/sign-in') : resolve('/alerts')} class="rounded-200">
    <Logo />
    <span class="sr-only">, home</span>
  </a>
  {#if email !== null}
    <div class="ml-auto flex items-center gap-400">
      <span class="hidden ts-primary-body-m text-text-on-dark-secondary mobile:inline">{email}</span
      >
      <form method="POST" action={resolve('/sign-out')}>
        <button type="submit" class={buttonStyles({ variant: 'on-dark', size: 'sm' })}>
          <LogOut size={16} />
          Sign Out
        </button>
      </form>
    </div>
  {/if}
</header>

<div class="flex min-h-[calc(100dvh-3.5rem)]">
  {#if email !== null}
    <nav
      aria-label="Main"
      class="w-[3.75rem] shrink-0 bg-surface-primary-cool-dark px-[0.375rem] py-500"
    >
      <ul class="flex flex-col items-center gap-300">
        {#each nav as item (item.route)}
          <!--
            By route, not by URL: on the server, resolve() writes a relative
            href (./alerts), which no pathname starts with, and the page would
            arrive with no item marked current.
          -->
          {@const current = page.route.id?.startsWith(item.route) ?? false}
          <li>
            <a
              href={resolve(item.route)}
              aria-current={current ? 'page' : undefined}
              class={[
                'flex w-[3rem] flex-col items-center gap-100 rounded-200 py-200 ts-primary-caption text-text-on-dark-primary transition-colors duration-300',
                current
                  ? 'bg-surface-primary-default hover:bg-surface-primary-hover'
                  : 'hover:bg-surface-primary-dark',
              ]}
            >
              <item.icon size={20} />
              {item.label}
            </a>
          </li>
        {/each}
      </ul>
    </nav>
  {/if}

  <div
    class={['flex min-w-0 flex-1 flex-col bg-surface-canvas', email !== null && 'rounded-tl-500']}
  >
    <main id="main" tabindex="-1" class="flex-1 outline-none">
      {@render children()}
    </main>
    <footer
      class="flex flex-wrap justify-end gap-400 border-t border-stroke-neutral-default bg-surface-container-primary px-600 py-500 ts-primary-button-s text-text-secondary-default"
    >
      <span>
        Built for the
        <a
          href="https://dados.ons.org.br"
          rel="external noopener"
          class="text-text-primary-link hover:text-text-primary-link-hover"
          >ONS's open curtailment data</a
        >
      </span>
      <span class="border-l border-stroke-neutral-default pl-400"
        >Spillway is a portfolio project</span
      >
    </footer>
  </div>
</div>
