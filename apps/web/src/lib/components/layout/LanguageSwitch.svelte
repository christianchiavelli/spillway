<script lang="ts">
  import { page } from '$app/state';
  import { editions, inEdition } from '#lib/i18n.ts';
  import { m } from '#lib/paraglide/messages.js';
  import { getLocale } from '#lib/paraglide/runtime.js';

  const current = $derived(getLocale());
</script>

<!--
  Both editions at once, EN | PT, the current one underlined: each link opens
  this same page, query and all, in the other language. A screen reader hears
  each language's name in that language. The page loads whole, so every word
  on it is read again in the new language.
-->
<nav aria-label={m.language_nav()} class="flex items-center">
  {#each editions as edition, i (edition.locale)}
    {#if i > 0}
      <span aria-hidden="true" class="h-400 w-px bg-stroke-neutral-on-dark"></span>
    {/if}
    <a
      href={inEdition(page.url, edition.locale)}
      hreflang={edition.locale}
      lang={edition.locale}
      aria-current={edition.locale === current ? 'true' : undefined}
      data-sveltekit-reload
      class={[
        'rounded-200 px-300 py-200 ts-primary-button-s underline-offset-4',
        edition.locale === current
          ? 'text-text-on-dark-primary underline decoration-2'
          : 'text-text-on-dark-secondary hover:text-text-on-dark-primary',
      ]}
    >
      <span aria-hidden="true">{edition.short}</span>
      <span class="sr-only">{edition.name}</span>
    </a>
  {/each}
</nav>
