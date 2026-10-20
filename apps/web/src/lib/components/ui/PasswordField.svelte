<script lang="ts">
  import Visibility from '~icons/material-symbols/visibility-outline-rounded';
  import VisibilityOff from '~icons/material-symbols/visibility-off-outline-rounded';

  import TextField from './TextField.svelte';
  import { onMount, type ComponentProps } from 'svelte';

  type Props = Omit<ComponentProps<typeof TextField>, 'type' | 'suffix' | 'trailing'> & {
    /** The reveal button's name, such as "Show password". */
    showLabel: string;
  };

  let { showLabel, ...rest }: Props = $props();

  let shown = $state(false);
  // The button needs the page's script, so it only appears once that runs:
  // without it, the field is a plain password field.
  let interactive = $state(false);
  onMount(() => {
    interactive = true;
  });
</script>

<!--
  A password field with a button that shows what was typed, since a long
  password is easy to get wrong unseen. The button keeps its name and says
  whether it is on with aria-pressed. The field never spellchecks, so a shown
  password is not sent to a spelling service; data-password lets a refused
  form clear it while it is shown.
-->
<TextField
  {...rest}
  type={shown ? 'text' : 'password'}
  spellcheck="false"
  autocapitalize="none"
  data-password
>
  {#snippet trailing()}
    {#if interactive}
      <button
        type="button"
        aria-pressed={shown}
        aria-controls={rest.id}
        onclick={() => {
          shown = !shown;
        }}
        class="mr-100 inline-flex size-[2rem] shrink-0 cursor-pointer items-center justify-center rounded-200 text-icon-secondary transition-colors duration-300 hover:bg-surface-container-hover hover:text-icon-primary"
      >
        <span class="sr-only">{showLabel}</span>
        {#if shown}
          <VisibilityOff width={20} height={20} aria-hidden="true" />
        {:else}
          <Visibility width={20} height={20} aria-hidden="true" />
        {/if}
      </button>
    {/if}
  {/snippet}
</TextField>
