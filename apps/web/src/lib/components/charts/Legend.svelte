<script lang="ts">
  import type { Series } from './geometry.ts';

  type Props = {
    series: readonly Series[];
    /** A square for bars and columns, a short line for lines. */
    mark?: 'bar' | 'line';
  };

  let { series, mark = 'bar' }: Props = $props();
</script>

<ul
  aria-label="Legend"
  class="flex items-center gap-500 ts-primary-body-s text-text-secondary-default"
>
  {#each series as s (s.label)}
    <li class="flex items-center gap-200">
      {#if mark === 'bar'}
        <span class={['size-[0.625rem] rounded-100', s.swatch]}></span>
      {:else}
        <svg width="12" height="2" aria-hidden="true" class="shrink-0">
          <rect width="12" height="2" rx="1" class={s.fill} />
        </svg>
      {/if}
      {s.label}
    </li>
  {/each}
</ul>
