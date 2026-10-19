<script lang="ts">
  import type { Snippet } from 'svelte';
  import BarChart from '~icons/material-symbols/bar-chart-rounded';
  import Factory from '~icons/material-symbols/factory-outline-rounded';
  import History from '~icons/material-symbols/history-rounded';
  import Login from '~icons/material-symbols/login-rounded';
  import Logout from '~icons/material-symbols/logout-rounded';
  import Notifications from '~icons/material-symbols/notifications-outline-rounded';

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

  // The data is open to everyone; Alerts asks a visitor to sign in first.
  const nav = [
    { route: '/', label: 'Overview', icon: BarChart },
    { route: '/plants', label: 'Plants', icon: Factory },
    { route: '/history', label: 'History', icon: History },
    { route: '/alerts', label: 'Alerts', icon: Notifications },
  ] as const;

  const isCurrent = (route: string) => {
    const id = page.route.id ?? '';
    return route === '/' ? id === '/' : id.startsWith(route);
  };
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
  <a href={resolve('/')} class="rounded-200">
    <Logo />
    <span class="sr-only">, home</span>
  </a>
  {#if email === null}
    <div class="ml-auto flex items-center gap-300">
      <a href={resolve('/sign-in')} class={buttonStyles({ variant: 'on-dark', size: 'sm' })}>
        <Login width={16} height={16} />
        Sign In
      </a>
      <!-- On a phone the header has room for one button; Sign In leads to the other. -->
      <span class="hidden mobile:block">
        <a href={resolve('/sign-up')} class={buttonStyles({ variant: 'primary', size: 'sm' })}>
          Create Account
        </a>
      </span>
    </div>
  {:else}
    <div class="ml-auto flex items-center gap-400">
      <span class="hidden ts-primary-body-m text-text-on-dark-secondary mobile:inline">{email}</span
      >
      <form method="POST" action={resolve('/sign-out')}>
        <button type="submit" class={buttonStyles({ variant: 'on-dark', size: 'sm' })}>
          <Logout width={16} height={16} />
          Sign Out
        </button>
      </form>
    </div>
  {/if}
</header>

<div class="flex min-h-[calc(100dvh-3.5rem)]">
  <!--
    The rail stays put under the header as the page scrolls: the pages are
    long, and an item sliding under the header would be half a target.
  -->
  <nav
    aria-label="Main"
    class="sticky top-[3.5rem] h-[calc(100dvh-3.5rem)] w-[4rem] shrink-0 self-start bg-surface-primary-cool-dark px-[0.375rem] py-500"
  >
    <ul class="flex flex-col items-center gap-300">
      {#each nav as item (item.route)}
        <!--
          By route, not by URL: on the server, resolve() writes a relative
          href (./alerts), which no pathname starts with, and the page would
          arrive with no item marked current.
        -->
        {@const current = isCurrent(item.route)}
        <li>
          <a
            href={resolve(item.route)}
            aria-current={current ? 'page' : undefined}
            class={[
              'flex w-[3.25rem] flex-col items-center gap-100 rounded-200 py-200 ts-primary-caption text-text-on-dark-primary transition-colors duration-300',
              current
                ? 'bg-surface-primary-default hover:bg-surface-primary-hover'
                : 'hover:bg-surface-primary-dark',
            ]}
          >
            <item.icon width={20} height={20} />
            {item.label}
          </a>
        </li>
      {/each}
    </ul>
  </nav>

  <div class="flex min-w-0 flex-1 flex-col rounded-tl-500 bg-surface-canvas">
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
      <span class="mobile:border-l mobile:border-stroke-neutral-default mobile:pl-400"
        >Spillway is a portfolio project</span
      >
    </footer>
  </div>
</div>
