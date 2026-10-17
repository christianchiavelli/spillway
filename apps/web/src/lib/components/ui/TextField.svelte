<script lang="ts">
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
  };

  const uid = $props.id();

  let {
    label,
    name,
    value,
    hint,
    error,
    suffix,
    id = uid,
    type = 'text',
    required = false,
    ...rest
  }: Props = $props();

  const messageId = $derived(`${id}-message`);
  const message = $derived(error ?? hint);
</script>

<!--
  The field: the label sits notched into the border and always shows, never a
  placeholder that leaves as you type. The message is tied to the input with
  aria-describedby, and focus draws the same ring as every other control.
-->
<div class="flex flex-col gap-100">
  <div
    class={[
      'relative flex h-[2.625rem] items-center rounded-200 border bg-surface-container-primary transition-colors duration-300',
      'has-[input:focus-visible]:outline-2 has-[input:focus-visible]:outline-offset-2 has-[input:focus-visible]:outline-stroke-primary-default',
      error ? 'border-stroke-error' : 'border-stroke-control hover:border-stroke-control-hover',
    ]}
  >
    <label
      for={id}
      class={[
        'absolute -top-300 left-300 bg-surface-container-primary px-100 ts-primary-subheading-xs',
        error ? 'text-text-semantic-error' : 'text-text-secondary-default',
      ]}
    >
      {label}{#if required}<span aria-hidden="true">&nbsp;*</span>{/if}
    </label>
    <input
      {id}
      {name}
      {type}
      {required}
      value={value ?? ''}
      aria-invalid={error === undefined ? undefined : true}
      aria-describedby={message === undefined ? undefined : messageId}
      class="h-full w-full min-w-0 rounded-200 bg-transparent px-300 ts-primary-body-l text-text-primary-default outline-none placeholder:text-text-primary-disabled"
      {...rest}
    />
    {#if suffix}
      <span aria-hidden="true" class="pr-300 ts-primary-body-m text-text-secondary-default"
        >{suffix}</span
      >
    {/if}
  </div>
  {#if message !== undefined}
    <p
      id={messageId}
      class={[
        'pl-200 ts-primary-subheading-xs',
        error === undefined ? 'text-text-secondary-default' : 'text-text-semantic-error',
      ]}
    >
      {message}
    </p>
  {/if}
</div>
