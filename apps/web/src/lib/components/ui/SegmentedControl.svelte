<script lang="ts">
  import type { LucideIcon } from '@lucide/svelte';

  type Option = { value: string; label: string; icon?: LucideIcon };
  type Props = {
    legend: string;
    name: string;
    options: readonly Option[];
    value?: string | undefined;
    error?: string | undefined;
    required?: boolean;
    /** The first option's id, which a link to the field moves focus to. */
    id?: string;
  };

  const uid = $props.id();
  let { legend, name, options, value, error, required = false, id = uid }: Props = $props();

  const messageId = $derived(`${id}-message`);
</script>

<!--
  A segmented control, built from native radio buttons: arrow keys move
  between options, a screen reader reads a group of choices, and the choice is
  posted with the form even without JavaScript. It wears the text field's
  frame, its legend notched into the border where a field's label sits, so a
  row of fields lines up whatever kind each one is, and the options sit clear
  of the legend's lower half. Each radio is see-through and laid over its
  option, so a click, a tap or a screen reader's touch lands on the radio
  itself, where the option is drawn.
-->
<fieldset
  class="relative flex min-w-0 flex-col gap-100"
  aria-describedby={error === undefined ? undefined : messageId}
>
  <legend
    class={[
      'absolute -top-300 left-300 z-20 bg-surface-container-primary px-100 ts-primary-subheading-xs',
      error ? 'text-text-semantic-error' : 'text-text-secondary-default',
    ]}
  >
    {legend}{#if required}<span aria-hidden="true">&nbsp;*</span>{/if}
  </legend>
  <div
    class={[
      'flex h-[2.625rem] w-fit items-stretch gap-100 rounded-200 border bg-surface-container-primary px-[0.1875rem] py-[0.4375rem] transition-colors duration-300',
      error ? 'border-stroke-error' : 'border-stroke-control hover:border-stroke-control-hover',
    ]}
  >
    {#each options as option, i (option.value)}
      <label class="relative flex">
        <!--
          defaultChecked, not checked: it is the checked attribute, which a
          form reset goes back to. Svelte sets checked as a property only, so
          the reset after saving an alert would leave no option chosen.
        -->
        <input
          id={i === 0 ? id : `${id}-${option.value}`}
          type="radio"
          class="peer absolute inset-0 z-10 m-0 cursor-pointer appearance-none rounded-100 opacity-0"
          {name}
          value={option.value}
          defaultChecked={value === option.value}
          {required}
        />
        <!--
          The hover colour is for the options not chosen: on the chosen one it
          would put blue text on the blue fill. Forced colours drop the fill,
          so there an outline marks the choice.
        -->
        <span
          class={[
            'flex items-center gap-200 rounded-100 px-[0.875rem] ts-primary-button-l transition-colors duration-300',
            'text-text-secondary-default peer-not-checked:peer-hover:text-text-primary-default',
            'peer-checked:bg-surface-primary-default peer-checked:text-text-on-dark-primary',
            'forced-colors:peer-checked:outline-2',
            'peer-focus-visible:outline-2 peer-focus-visible:outline-offset-1 peer-focus-visible:outline-stroke-primary-default',
          ]}
        >
          {#if option.icon}
            <option.icon size={16} />
          {/if}
          {option.label}
        </span>
      </label>
    {/each}
  </div>
  {#if error !== undefined}
    <p id={messageId} class="pl-200 ts-primary-subheading-xs text-text-semantic-error">
      {error}
    </p>
  {/if}
</fieldset>
