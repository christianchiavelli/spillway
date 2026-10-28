<script lang="ts">
  import Check from '~icons/material-symbols/check-rounded';
  import Contrast from '~icons/material-symbols/contrast';

  import { enhance, type SubmitFunction } from '$app/forms';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { localized } from '#lib/i18n.ts';
  import { m } from '#lib/paraglide/messages.js';
  import { isTheme, type Theme } from '#lib/theme.ts';

  import ThemeScrap from './ThemeScrap.svelte';

  type Props = {
    /** The theme the reader picked, or System. */
    theme: Theme;
    /**
     * Called the moment another is picked, before the server has kept it, and
     * with the one before if the server does not keep it.
     */
    onchange: (theme: Theme) => void;
  };

  let { theme, onchange }: Props = $props();
  const id = $props.id();

  let menu = $state<HTMLElement>();

  const choices = [
    { value: 'light', label: m.theme_light },
    { value: 'dark', label: m.theme_dark },
    { value: 'system', label: m.theme_system },
  ] as const;

  const name = $derived(choices.find((choice) => choice.value === theme)?.label());

  /**
   * Without JavaScript the form posts, and the page comes back in the theme
   * picked. With it, the page takes the theme at once and the menu closes;
   * the post only stores it, so nothing is loaded again, and a post refused
   * or lost puts the page back in the theme it had, which the next one would
   * be in.
   */
  const choose: SubmitFunction = ({ formData, cancel }) => {
    const picked = formData.get('theme');
    if (!isTheme(picked)) {
      cancel();
      return;
    }
    const before = theme;
    onchange(picked);
    menu?.hidePopover();
    // The page already wears it, so the redirect back to it is not followed.
    return ({ result }) => {
      if (result.type === 'failure' || result.type === 'error') onchange(before);
    };
  };
</script>

<!--
  The theme, picked from a menu under a button in the header. The menu is a
  native popover, so it opens, closes on Escape or a click outside, and is
  reached by Tab after its button, all without JavaScript; each choice is a
  button of one form that posts it. Each shows a scrap of the app in its
  colours, and System says which of the two the device asks for now.
-->
<button
  type="button"
  popovertarget="{id}-menu"
  class="flex size-700 cursor-pointer items-center justify-center rounded-200 text-text-on-dark-secondary transition-colors duration-300 [anchor-name:--theme-menu] hover:bg-surface-primary-dark hover:text-text-on-dark-primary has-[+:popover-open]:bg-surface-primary-dark has-[+:popover-open]:text-text-on-dark-primary"
>
  <Contrast width={20} height={20} />
  <span class="sr-only">{m.theme_button({ theme: name ?? m.theme_system() })}</span>
</button>
<div
  bind:this={menu}
  id="{id}-menu"
  popover
  role="group"
  aria-labelledby="{id}-heading"
  class="theme-menu w-[15.5rem] max-w-[calc(100%-1.5rem)] rounded-200 border border-stroke-neutral-default bg-surface-container-primary py-200 text-text-primary-default shadow-overlay"
>
  <p
    id="{id}-heading"
    class="px-500 pt-300 pb-200 ts-primary-subheading-xs text-text-secondary-default"
  >
    {m.theme_heading()}
  </p>
  <form method="POST" action={localized(resolve('/theme'))} use:enhance={choose}>
    <input type="hidden" name="back" value={page.url.pathname + page.url.search} />
    <ul>
      {#each choices as choice (choice.value)}
        {@const current = choice.value === theme}
        <li>
          <button
            type="submit"
            name="theme"
            value={choice.value}
            aria-pressed={current}
            aria-describedby={choice.value === 'system' ? `${id}-follows` : undefined}
            class="grid w-full cursor-pointer grid-cols-[2.25rem_minmax(0,1fr)_1.25rem] items-center gap-400 px-500 py-300 text-start ts-primary-body-m transition-colors duration-300 hover:bg-surface-container-secondary focus-visible:-outline-offset-2"
          >
            <ThemeScrap theme={choice.value} />
            <span class="flex flex-col">
              {choice.label()}
              {#if choice.value === 'system'}
                <!-- For the eye: a screen reader hears that it follows the device. -->
                <span aria-hidden="true" class="ts-primary-body-s text-text-secondary-default">
                  <span class="device-dark:hidden">{m.theme_light_here()}</span>
                  <span class="hidden device-dark:inline">{m.theme_dark_here()}</span>
                </span>
              {/if}
            </span>
            <Check width={20} height={20} class={current ? 'text-icon-primary' : 'invisible'} />
          </button>
        </li>
      {/each}
    </ul>
    <span id="{id}-follows" hidden>{m.theme_follows()}</span>
  </form>
</div>
