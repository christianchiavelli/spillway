<script lang="ts">
  import ChevronLeft from '~icons/material-symbols/chevron-left-rounded';
  import ChevronRight from '~icons/material-symbols/chevron-right-rounded';

  import type { ResolvedPathname } from '$app/types';

  import PillNav from './PillNav.svelte';
  import StepLink from './StepLink.svelte';

  type Step = { href: ResolvedPathname | null; label: string };
  type Props = {
    /** The kinds of period, as links. */
    kinds: readonly { href: ResolvedPathname; label: string; current: boolean }[];
    /** The period shown, as a heading. */
    title: string;
    /** The period before and the one after; a null href where there is none. */
    previous: Step;
    next: Step;
    /** How far the data goes, under the title. */
    note: string;
  };

  let { kinds, title, previous, next, note }: Props = $props();
</script>

<!--
  The period bar. The arrows come before the title, so they stay put whatever
  its length and stepping through days is one click after another on the same
  spot; the kind of period, chosen far less often, sits at the end. The title's
  margin pushes it there, so on a row too narrow for it, it starts the next
  line from the left, as the list's filters do. On a phone the arrows go to the
  edges with the title centred between them, and the kinds fill the row below.
-->
<div
  class="grid grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-x-300 gap-y-500 mobile:flex mobile:flex-wrap mobile:gap-x-500 mobile:gap-y-400"
>
  <div class="contents mobile:flex mobile:gap-200">
    <StepLink
      href={previous.href}
      label={previous.label}
      icon={ChevronLeft}
      class="col-start-1 row-start-1"
    />
    <StepLink
      href={next.href}
      label={next.label}
      icon={ChevronRight}
      class="col-start-3 row-start-1"
    />
  </div>
  <div class="col-start-2 row-start-1 flex flex-col text-center mobile:mr-auto mobile:text-left">
    <h2 class="ts-primary-subheading-m text-balance text-text-primary-default">{title}</h2>
    <p class="ts-primary-body-s text-text-secondary-default">{note}</p>
  </div>
  <div class="col-span-3">
    <PillNav label="Period" items={kinds} />
  </div>
</div>
