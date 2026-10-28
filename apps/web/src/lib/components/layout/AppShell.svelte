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
  import { localized } from '#lib/i18n.ts';
  import { m } from '#lib/paraglide/messages.js';
  import type { Theme } from '#lib/theme.ts';

  import LanguageSwitch from './LanguageSwitch.svelte';
  import Logo from './Logo.svelte';
  import ThemeMenu from './ThemeMenu.svelte';

  type Props = {
    /** The signed-in address, or null for a visitor. */
    email: string | null;
    /** The theme the reader picked, and what to do when they pick another. */
    theme: Theme;
    onthemechange: (theme: Theme) => void;
    children: Snippet;
  };

  let { email, theme, onthemechange, children }: Props = $props();

  // The data is open to everyone; Alerts asks a visitor to sign in first.
  const nav = [
    { route: '/', label: m.nav_overview, icon: BarChart },
    { route: '/plants', label: m.nav_plants, icon: Factory },
    { route: '/history', label: m.nav_history, icon: History },
    { route: '/alerts', label: m.nav_alerts, icon: Notifications },
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
  {m.shell_skip()}
</a>

<!--
  The chrome: a dark header across the top and a dark rail on the left. On the
  narrowest phones, under 360px, its gaps close up so the logo keeps its size.
-->
<header
  data-chrome
  class="sticky top-0 z-(--z-header) flex h-[3.5rem] items-center gap-600 bg-surface-primary-cool-dark px-400 max-[22.5rem]:gap-300"
>
  <a href={localized(resolve('/'))} class="shrink-0 rounded-200">
    <Logo />
    <span class="sr-only">{m.shell_home()}</span>
  </a>
  <div class="ml-auto flex items-center gap-400 max-[22.5rem]:gap-200">
    <LanguageSwitch />
    <ThemeMenu {theme} onchange={onthemechange} />
    {#if email === null}
      <div class="flex items-center gap-300">
        <a
          href={localized(resolve('/sign-in'))}
          class={buttonStyles({ variant: 'on-dark', size: 'sm' })}
        >
          <Login width={16} height={16} />
          {m.shell_sign_in()}
        </a>
        <!-- On a phone the header has room for one button; Sign In leads to the other. -->
        <span class="hidden mobile:block">
          <a
            href={localized(resolve('/sign-up'))}
            class={buttonStyles({ variant: 'primary', size: 'sm' })}
          >
            {m.shell_create_account()}
          </a>
        </span>
      </div>
    {:else}
      <span class="hidden ts-primary-body-m text-text-on-dark-secondary mobile:inline">{email}</span
      >
      <form method="POST" action={localized(resolve('/sign-out'))}>
        <button type="submit" class={buttonStyles({ variant: 'on-dark', size: 'sm' })}>
          <Logout width={16} height={16} />
          {m.shell_sign_out()}
        </button>
      </form>
    {/if}
  </div>
</header>

<div class="flex min-h-[calc(100dvh-3.5rem)]">
  <!--
    The rail stays put under the header as the page scrolls: the pages are
    long, and an item sliding under the header would be half a target.
  -->
  <nav
    data-chrome
    aria-label={m.shell_nav()}
    class="sticky top-[3.5rem] h-[calc(100dvh-3.5rem)] w-[4rem] shrink-0 self-start bg-surface-primary-cool-dark px-[0.375rem] py-500"
  >
    <ul class="flex flex-col items-center gap-300">
      {#each nav as item (item.route)}
        <!--
          By route, not by URL: the same route is /alerts and /pt-br/alerts.
          A label too long for the rail, like Visão geral, wraps centred.
        -->
        {@const current = isCurrent(item.route)}
        <li>
          <a
            href={localized(resolve(item.route))}
            aria-current={current ? 'page' : undefined}
            class={[
              'flex w-[3.25rem] flex-col items-center gap-100 rounded-200 py-200 text-center ts-primary-caption text-text-on-dark-primary transition-colors duration-300',
              current
                ? 'bg-surface-primary-default hover:bg-surface-primary-hover'
                : 'hover:bg-surface-primary-dark',
            ]}
          >
            <item.icon width={20} height={20} />
            {item.label()}
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
        {m.shell_built_for()}
        <a
          href="https://dados.ons.org.br"
          rel="external noopener"
          class="text-text-primary-link hover:text-text-primary-link-hover">{m.shell_ons_data()}</a
        >
      </span>
      <span class="mobile:border-l mobile:border-stroke-neutral-default mobile:pl-400"
        >{m.shell_portfolio()}</span
      >
    </footer>
  </div>
</div>
