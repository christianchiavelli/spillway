<script lang="ts" module>
  export type Toast = { id: string; title: string; body?: string | undefined };
</script>

<script lang="ts">
  import type { Attachment } from 'svelte/attachments';
  import { on } from 'svelte/events';
  import CheckCircle from '~icons/material-symbols/check-circle-outline-rounded';
  import Close from '~icons/material-symbols/close-rounded';

  import { m } from '#lib/paraglide/messages.js';

  type Props = { toasts: readonly Toast[]; ondismiss: (id: string) => void };

  let { toasts, ondismiss }: Props = $props();

  /** How long a toast stays once nobody is reading it. */
  const lifetime = 8000;

  /**
   * Closes the toast once it has been up for its lifetime, counted again from
   * the start each time the pointer leaves it, focus leaves it or the tab is
   * shown again: someone pointing at it, tabbing through it or away from the
   * page has not finished with it.
   */
  function leavesOnItsOwn(id: string): Attachment<HTMLElement> {
    return (toast) => {
      let hovered = false;
      let focused = false;
      let timer: ReturnType<typeof setTimeout> | undefined;

      const restart = () => {
        clearTimeout(timer);
        timer =
          hovered || focused || document.hidden
            ? undefined
            : setTimeout(() => {
                ondismiss(id);
              }, lifetime);
      };

      const stop = [
        on(toast, 'pointerenter', () => {
          hovered = true;
          restart();
        }),
        on(toast, 'pointerleave', () => {
          hovered = false;
          restart();
        }),
        on(toast, 'focusin', () => {
          focused = true;
          restart();
        }),
        on(toast, 'focusout', (event) => {
          focused = event.relatedTarget instanceof Node && toast.contains(event.relatedTarget);
          restart();
        }),
        on(document, 'visibilitychange', restart),
      ];
      restart();

      return () => {
        clearTimeout(timer);
        for (const off of stop) off();
      };
    };
  }
</script>

<!--
  The toast, in its success tone. The region is always on the page, empty or
  not, because a screen reader only announces what is added to a live region
  that already exists. It sits at the top right, where a dashboard's notices
  are looked for, under the page header's band so it covers none of the page's
  actions. A toast leaves on its own after a while, but not while the pointer
  or focus is on it, so it is never taken away from someone reading it; the
  next page takes it away too. It leaves without an animation: Svelte's
  transitions would not follow the reduced-motion rule in app.css.
-->
<div
  role="status"
  aria-live="polite"
  class="toast-region pointer-events-none fixed right-600 z-(--z-toast) flex w-[min(24rem,calc(100vw-7rem))] flex-col gap-300"
>
  {#each toasts as toast (toast.id)}
    <div
      class="pointer-events-auto flex min-h-[3.125rem] animate-toast-in items-start gap-500 rounded-300 border border-text-semantic-success bg-visual-light-green p-500 shadow-overlay"
      {@attach leavesOnItsOwn(toast.id)}
    >
      <span class="flex size-700 shrink-0 items-center justify-center rounded-full bg-white">
        <CheckCircle width={20} height={20} class="text-icon-success" />
      </span>
      <div class="flex min-w-0 flex-1 flex-col gap-100">
        <p class="ts-primary-subheading-s text-text-primary-default">{toast.title}</p>
        {#if toast.body}
          <p class="ts-primary-body-s break-words text-text-secondary-default">{toast.body}</p>
        {/if}
      </div>
      <button
        type="button"
        class="-m-200 flex size-700 shrink-0 cursor-pointer items-center justify-center rounded-200 text-icon-secondary hover:bg-white/60"
        onclick={() => {
          ondismiss(toast.id);
        }}
      >
        <Close width={16} height={16} />
        <span class="sr-only">{m.toast_dismiss({ title: toast.title })}</span>
      </button>
    </div>
  {/each}
</div>
