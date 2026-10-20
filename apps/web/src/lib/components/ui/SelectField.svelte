<script lang="ts">
  import ArrowDown from '~icons/material-symbols/keyboard-arrow-down-rounded';

  import FieldMessage from './FieldMessage.svelte';
  import type { HTMLSelectAttributes } from 'svelte/elements';

  type Option = { value: string; label: string };
  type Props = Omit<HTMLSelectAttributes, 'class' | 'value' | 'children'> & {
    label: string;
    name: string;
    options: readonly Option[];
    value?: string | undefined;
    /** The first, empty choice, which a required field cannot be sent with. */
    placeholder: string;
    hint?: string | undefined;
    error?: string | undefined;
  };

  const uid = $props.id();

  let {
    label,
    name,
    options,
    value = '',
    placeholder,
    hint,
    error,
    id = uid,
    required = false,
    ...rest
  }: Props = $props();

  const messageId = $derived(`${id}-message`);
  const described = $derived(error !== undefined || hint !== undefined);
</script>

<!--
  A native select in the frame of the text field, not a list drawn by a
  script: it works in a form posted without JavaScript, and the platform gives
  it keyboard and screen reader support on every device.
-->
<div class="flex flex-col gap-100">
  <div
    class={[
      'relative flex h-[2.625rem] items-center rounded-200 border bg-surface-container-primary transition-colors duration-300',
      'has-[select:focus-visible]:outline-2 has-[select:focus-visible]:outline-offset-2 has-[select:focus-visible]:outline-stroke-primary-default',
      error ? 'border-stroke-error' : 'border-stroke-control hover:border-stroke-control-hover',
    ]}
  >
    <label
      for={id}
      class={[
        'absolute -top-300 left-200 bg-surface-container-primary px-200 ts-primary-subheading-xs',
        error ? 'text-text-semantic-error' : 'text-text-secondary-default',
      ]}
    >
      {label}{#if required}<span aria-hidden="true">&nbsp;*</span>{/if}
    </label>
    <select
      {id}
      {name}
      {required}
      aria-invalid={error === undefined ? undefined : true}
      aria-describedby={described ? messageId : undefined}
      class="select-menu h-full w-full min-w-0 cursor-pointer rounded-200 bg-transparent pr-800 pl-300 ts-primary-body-l text-text-primary-default outline-none"
      {...rest}
    >
      <option value="" disabled selected={value === ''}>{placeholder}</option>
      {#each options as option (option.value)}
        <option value={option.value} selected={value === option.value}>{option.label}</option>
      {/each}
    </select>
    <ArrowDown
      width={16}
      height={16}
      class="pointer-events-none absolute right-300 text-icon-secondary"
    />
  </div>
  <FieldMessage id={messageId} {hint} {error} />
</div>
