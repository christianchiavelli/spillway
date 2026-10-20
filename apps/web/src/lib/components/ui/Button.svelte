<script lang="ts">
  import ProgressActivity from '~icons/material-symbols/progress-activity';

  import { type ButtonVariants, buttonStyles } from './button-styles.ts';
  import type { Icon as IconComponent } from './icon.ts';
  import type { Snippet } from 'svelte';
  import type { HTMLButtonAttributes } from 'svelte/elements';

  type Props = Omit<HTMLButtonAttributes, 'class' | 'children'> &
    ButtonVariants & {
      children: Snippet;
      /** Drawn before the label, at the size that goes with the button. */
      icon?: IconComponent | undefined;
      /**
       * While what it started is under way: a spinner takes the icon's place
       * and the button refuses another press, without leaving the tab order
       * as a disabled one would.
       */
      pending?: boolean;
      class?: string | undefined;
    };

  let {
    variant,
    size = 'lg',
    icon: Icon,
    pending = false,
    type = 'button',
    children,
    class: className,
    ...rest
  }: Props = $props();

  const iconSize = $derived(size === 'sm' ? 16 : 20);
</script>

<button
  {type}
  class={buttonStyles({ variant, size, class: className })}
  aria-disabled={pending || undefined}
  {...rest}
>
  {#if pending}
    <!-- Still under reduced motion: the label says the same thing. -->
    <ProgressActivity
      width={iconSize}
      height={iconSize}
      aria-hidden="true"
      class="shrink-0 motion-safe:animate-spin"
    />
  {:else if Icon}
    <Icon width={iconSize} height={iconSize} />
  {/if}
  {@render children()}
</button>
