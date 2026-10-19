<script lang="ts">
  import { type ButtonVariants, buttonStyles } from './button-styles.ts';
  import type { Icon as IconComponent } from './icon.ts';
  import type { Snippet } from 'svelte';
  import type { HTMLButtonAttributes } from 'svelte/elements';

  type Props = Omit<HTMLButtonAttributes, 'class' | 'children'> &
    ButtonVariants & {
      children: Snippet;
      /** Drawn before the label, at the size that goes with the button. */
      icon?: IconComponent | undefined;
      class?: string | undefined;
    };

  let {
    variant,
    size = 'lg',
    icon: Icon,
    type = 'button',
    children,
    class: className,
    ...rest
  }: Props = $props();
</script>

<button {type} class={buttonStyles({ variant, size, class: className })} {...rest}>
  {#if Icon}
    <Icon width={size === 'sm' ? 16 : 20} height={size === 'sm' ? 16 : 20} />
  {/if}
  {@render children()}
</button>
