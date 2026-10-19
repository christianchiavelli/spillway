<script lang="ts">
  import ChevronLeft from '~icons/material-symbols/chevron-left-rounded';
  import ChevronRight from '~icons/material-symbols/chevron-right-rounded';

  import type { ResolvedPathname } from '$app/types';

  import { buttonStyles } from './button-styles.ts';

  type Props = {
    /** The first and last item on this page, counted from 1, and all of them. */
    from: number;
    to: number;
    total: number;
    /** The pages before and after; null where there is none. */
    previous: ResolvedPathname | null;
    next: ResolvedPathname | null;
    /** The words on the two links, such as Newer and Older. */
    labels?: { previous: string; next: string };
  };

  let {
    from,
    to,
    total,
    previous,
    next,
    labels = { previous: 'Previous', next: 'Next' },
  }: Props = $props();

  const whole = new Intl.NumberFormat('en-US');
  const button = buttonStyles({ variant: 'secondary', size: 'sm' });
  // The secondary button's frame, greyed: the lite build merges no classes, so
  // the disabled colour cannot be laid over the variant's own.
  const off =
    'inline-flex min-h-[2rem] cursor-not-allowed items-center gap-300 rounded-200 bg-surface-container-primary px-400 py-300 ts-primary-button-s whitespace-nowrap text-text-primary-disabled ring-1 ring-stroke-neutral-default ring-inset';
</script>

<div class="flex flex-wrap items-center justify-between gap-400">
  <p class="ts-primary-body-m text-text-secondary-default">
    {whole.format(from)} to {whole.format(to)} of {whole.format(total)}
  </p>
  <nav aria-label="Pages" class="flex items-center gap-300">
    {#if previous === null}
      <span aria-disabled="true" class={off}
        ><ChevronLeft width={16} height={16} />{labels.previous}</span
      >
    {:else}
      <a href={previous} data-sveltekit-noscroll class={button}
        ><ChevronLeft width={16} height={16} />{labels.previous}</a
      >
    {/if}
    {#if next === null}
      <span aria-disabled="true" class={off}
        >{labels.next}<ChevronRight width={16} height={16} /></span
      >
    {:else}
      <a href={next} data-sveltekit-noscroll class={button}
        >{labels.next}<ChevronRight width={16} height={16} /></a
      >
    {/if}
  </nav>
</div>
