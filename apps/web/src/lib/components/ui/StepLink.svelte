<script lang="ts">
  import type { ResolvedPathname } from '$app/types';

  import type { Icon as IconComponent } from './icon.ts';

  type Props = {
    /** Where it goes, or null where there is nothing that way yet. */
    href: ResolvedPathname | null;
    /** What it is called, since it shows an arrow only. */
    label: string;
    icon: IconComponent;
    class?: string | undefined;
  };

  let { href, label, icon: Icon, class: className }: Props = $props();

  const frame =
    'inline-flex size-[2rem] shrink-0 items-center justify-center rounded-200 ring-1 ring-stroke-neutral-default ring-inset';
</script>

{#if href === null}
  <!-- Not a link: a link to a period with no data would lead nowhere. -->
  <span
    role="link"
    aria-disabled="true"
    aria-label={label}
    class={[frame, 'text-text-primary-disabled', className]}
  >
    <Icon width={16} height={16} />
  </span>
{:else}
  <a
    {href}
    aria-label={label}
    data-sveltekit-noscroll
    data-sveltekit-keepfocus
    class={[
      frame,
      'bg-surface-container-primary text-text-primary-default hover:ring-stroke-neutral-hover',
      className,
    ]}
  >
    <Icon width={16} height={16} />
  </a>
{/if}
