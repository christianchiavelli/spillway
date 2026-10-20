<script lang="ts">
  import FieldMessage from './FieldMessage.svelte';
  import type { Snippet } from 'svelte';
  import type { HTMLInputAttributes } from 'svelte/elements';

  type Props = Omit<HTMLInputAttributes, 'class' | 'value' | 'children'> & {
    label: string;
    name: string;
    value?: string | number | null | undefined;
    /** Shown under the field until there is an error to show instead. */
    hint?: string | undefined;
    error?: string | undefined;
    /** Text after the value, such as a unit. */
    suffix?: string | undefined;
    /** A control inside the frame, after the value, such as a password's reveal button. */
    trailing?: Snippet | undefined;
  };

  const uid = $props.id();

  let {
    label,
    name,
    value,
    hint,
    error,
    suffix,
    trailing,
    id = uid,
    type = 'text',
    required = false,
    ...rest
  }: Props = $props();

  const messageId = $derived(`${id}-message`);
  const described = $derived(error !== undefined || hint !== undefined);
</script>

<!--
  The field: the label sits notched into the border and always shows, never a
  placeholder that leaves as you type, and the label, the value and the message
  start on one line. The message is tied to the input with aria-describedby,
  and focus draws the same ring as every other control.
-->
<div class="group flex flex-col gap-100">
  <div
    class={[
      'relative flex h-[2.625rem] items-center rounded-200 border bg-surface-container-primary transition-colors duration-300',
      'has-[input:focus-visible]:outline-2 has-[input:focus-visible]:outline-offset-2 has-[input:focus-visible]:outline-stroke-primary-default',
      'has-[input:disabled]:cursor-not-allowed has-[input:disabled]:border-stroke-neutral-default',
      error ? 'border-stroke-error' : 'border-stroke-control hover:border-stroke-control-hover',
    ]}
  >
    <label
      for={id}
      class={[
        'absolute -top-300 left-200 bg-surface-container-primary px-200 ts-primary-subheading-xs group-has-[input:disabled]:text-text-primary-disabled',
        error ? 'text-text-semantic-error' : 'text-text-secondary-default',
      ]}
    >
      {label}{#if required}<span aria-hidden="true">&nbsp;*</span>{/if}
    </label>
    <!--
      defaultValue, not value: hydration sets a value property, and would wipe
      what someone typed before the page's JavaScript arrived. A default value
      only shows while the field is untouched.
    -->
    <input
      {id}
      {name}
      {type}
      {required}
      defaultValue={value ?? ''}
      aria-invalid={error === undefined ? undefined : true}
      aria-describedby={described ? messageId : undefined}
      class={[
        'h-full w-full min-w-0 rounded-200 bg-transparent px-300 ts-primary-body-l text-text-primary-default outline-none placeholder:text-text-primary-disabled',
        'disabled:cursor-not-allowed disabled:text-text-primary-disabled',
        // A browser paints a field it filled in its own blue, which the
        // notched label would cut across; the field keeps its white instead.
        'autofill:shadow-[inset_0_0_0_2rem_var(--color-surface-container-primary)] autofill:[-webkit-text-fill-color:var(--color-text-primary-default)]',
      ]}
      {...rest}
    />
    {#if suffix}
      <span aria-hidden="true" class="pr-300 ts-primary-body-m text-text-secondary-default"
        >{suffix}</span
      >
    {/if}
    {@render trailing?.()}
  </div>
  <FieldMessage id={messageId} {hint} {error} />
</div>
