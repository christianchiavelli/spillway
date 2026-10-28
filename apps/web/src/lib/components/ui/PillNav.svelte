<script lang="ts">
  import type { ResolvedPathname } from '$app/types';

  import type { Icon } from './icon.ts';

  type Item = { href: ResolvedPathname; label: string; current: boolean; icon?: Icon | undefined };
  type Props = {
    /** What the links choose, for the landmark's name. */
    label: string;
    items: readonly Item[];
  };

  let { label, items }: Props = $props();
</script>

<!--
  The segmented control's look, made of links: each choice is an address, so
  it works without JavaScript and can be shared. On a phone the pill fills its
  row in equal parts; from the mobile breakpoint up it is as wide as its labels.
  The page keeps its scroll and focus when a choice is followed.
-->
<nav
  aria-label={label}
  class="flex w-full rounded-full border border-stroke-neutral-default bg-surface-track p-[0.125rem] mobile:w-fit"
>
  {#each items as item (item.href)}
    <a
      href={item.href}
      aria-current={item.current ? 'page' : undefined}
      data-sveltekit-noscroll
      data-sveltekit-keepfocus
      class={[
        'flex flex-1 items-center justify-center gap-200 rounded-full px-[0.875rem] py-[0.375rem] ts-primary-button-l transition-colors duration-300 mobile:flex-none',
        item.current
          ? 'bg-surface-primary-default text-text-on-dark-primary'
          : 'text-text-secondary-default hover:text-text-primary-default',
      ]}
    >
      {#if item.icon}
        <item.icon width={16} height={16} />
      {/if}
      {item.label}
    </a>
  {/each}
</nav>
