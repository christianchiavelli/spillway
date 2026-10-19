<script lang="ts">
  import type { Icon } from './icon.ts';

  type Option = { value: string; label: string; icon?: Icon };
  type Props = {
    legend: string;
    name: string;
    options: readonly Option[];
    value?: string | undefined;
    error?: string | undefined;
    required?: boolean;
    /** The first option's id, which a link to the field moves focus to. */
    id?: string;
    /** Small, for a switch in a card's header. */
    size?: 'md' | 'sm';
    /** A legend only a screen reader needs, where the options say it all. */
    hideLegend?: boolean;
  };

  const uid = $props.id();
  let {
    legend,
    name,
    options,
    value,
    error,
    required = false,
    id = uid,
    size = 'md',
    hideLegend = false,
  }: Props = $props();

  const messageId = $derived(`${id}-message`);
</script>

<!--
  A segmented control: a pill of options, the chosen one filled. It is built
  from native radio buttons, so arrow keys move between options, a screen
  reader reads a group of choices, and the choice is posted with the form even
  without JavaScript. Each radio is see-through and laid over its option, so a
  click, a tap or a screen reader's touch lands on the radio itself.
-->
<fieldset
  class="flex w-full min-w-0 flex-col mobile:w-auto"
  aria-describedby={error === undefined ? undefined : messageId}
>
  <legend
    class={[
      hideLegend ? 'sr-only' : 'mb-300 ts-primary-subheading-xs',
      error ? 'text-text-semantic-error' : 'text-text-secondary-default',
    ]}
  >
    {legend}{#if required}<span aria-hidden="true">&nbsp;*</span>{/if}
  </legend>
  <!--
    On a phone the pill fills its row in equal parts, so it lines up with the
    fields under it, or a card's switch with the title over it; from the
    mobile breakpoint up it is as wide as its labels.
  -->
  <div
    class={[
      'flex w-full rounded-full border bg-grey-25 p-[0.125rem] mobile:w-fit',
      error ? 'border-stroke-error' : 'border-stroke-neutral-default',
    ]}
  >
    {#each options as option, i (option.value)}
      <label class="relative flex flex-1 mobile:flex-none">
        <!--
          defaultChecked, not checked: it is the checked attribute, which a
          form reset goes back to. Svelte sets checked as a property only, so
          the reset after saving an alert would leave no option chosen.
        -->
        <input
          id={i === 0 ? id : `${id}-${option.value}`}
          type="radio"
          class="peer absolute inset-0 z-10 m-0 cursor-pointer appearance-none rounded-full opacity-0"
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
            'flex flex-1 items-center justify-center gap-200 rounded-full transition-colors duration-300',
            size === 'sm'
              ? 'px-[0.625rem] py-[0.25rem] ts-primary-button-s'
              : 'px-[0.875rem] py-[0.375rem] ts-primary-button-l',
            'text-text-secondary-default peer-not-checked:peer-hover:text-text-primary-default',
            'peer-checked:bg-surface-primary-default peer-checked:text-text-on-dark-primary',
            'forced-colors:peer-checked:outline-2',
            'peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-stroke-primary-default',
          ]}
        >
          {#if option.icon}
            <option.icon width={16} height={16} />
          {/if}
          {option.label}
        </span>
      </label>
    {/each}
  </div>
  {#if error !== undefined}
    <p id={messageId} class="mt-100 pl-200 ts-primary-subheading-xs text-text-semantic-error">
      {error}
    </p>
  {/if}
</fieldset>
