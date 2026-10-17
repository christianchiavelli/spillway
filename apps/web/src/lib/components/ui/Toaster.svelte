<script lang="ts" module>
  export type Toast = { id: string; title: string; body?: string | undefined };
</script>

<script lang="ts">
  import { CircleCheck, X } from '@lucide/svelte';

  type Props = { toasts: readonly Toast[]; ondismiss: (id: string) => void };

  let { toasts, ondismiss }: Props = $props();
</script>

<!--
  The toast, in its success tone. The region is always on the page, empty or
  not, because a screen reader only announces what is added to a live region
  that already exists. A toast stays until it is closed or the page changes:
  one that leaves on a timer can go before it is read.
-->
<div
  role="status"
  aria-live="polite"
  class="pointer-events-none fixed top-[5rem] right-600 z-(--z-toast) flex w-[min(24rem,calc(100vw-3rem))] flex-col gap-300"
>
  {#each toasts as toast (toast.id)}
    <div
      class="pointer-events-auto flex min-h-[3.125rem] animate-toast-in items-start gap-500 rounded-300 border border-text-semantic-success bg-visual-light-green p-500 shadow-overlay"
    >
      <span class="flex size-700 shrink-0 items-center justify-center rounded-full bg-white">
        <CircleCheck size={20} class="text-icon-success" />
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
        <X size={16} />
        <span class="sr-only">Dismiss “{toast.title}”</span>
      </button>
    </div>
  {/each}
</div>
