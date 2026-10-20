import { tv, type VariantProps } from 'tailwind-variants/lite';

/**
 * The button: primary, secondary and ghost, in two sizes, and a variant
 * for the dark header. In a module of its own so a link that acts as a button
 * wears the same classes on an `<a>`.
 *
 * The lite build, without tailwind-merge: no caller passes a class that fights
 * these, and the merge would cost the browser 20 kB for nothing.
 */
export const buttonStyles = tv({
  base: [
    'inline-flex shrink-0 cursor-pointer items-center justify-center gap-300 rounded-200 px-400 py-300 whitespace-nowrap',
    'transition-[background-color,box-shadow,color] duration-300 ease-out focus-visible:shadow-focus-lift aria-disabled:cursor-progress',
    'disabled:cursor-not-allowed disabled:bg-surface-container-primary disabled:text-text-primary-disabled disabled:ring-1 disabled:ring-stroke-neutral-default disabled:ring-inset',
  ],
  variants: {
    variant: {
      primary:
        'bg-surface-primary-default text-text-on-dark-primary hover:bg-surface-primary-hover focus-visible:bg-surface-primary-hover',
      secondary:
        'bg-surface-container-primary text-text-primary-default ring-1 ring-stroke-neutral-default ring-inset hover:ring-stroke-neutral-hover',
      ghost: 'text-text-primary-default hover:bg-surface-container-hover',
      'on-dark':
        'text-text-on-dark-primary ring-1 ring-stroke-neutral-on-dark ring-inset hover:bg-surface-primary-dark',
    },
    size: {
      lg: 'min-h-[2.25rem] ts-primary-button-l',
      sm: 'min-h-[2rem] ts-primary-button-s',
      /** A form's one action: its full width, and as tall as its fields. */
      block: 'min-h-[2.625rem] w-full ts-primary-button-l',
    },
  },
  defaultVariants: { variant: 'primary', size: 'lg' },
});

export type ButtonVariants = VariantProps<typeof buttonStyles>;
