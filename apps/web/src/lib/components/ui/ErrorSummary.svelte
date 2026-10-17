<script lang="ts">
  import { CircleAlert } from '@lucide/svelte';

  type Item = { href: string; text: string };
  type Props = {
    title: string;
    /** One link per field, to its input. Empty for a problem with the whole form. */
    items?: readonly Item[];
    detail?: string | undefined;
  };

  let { title, items = [], detail }: Props = $props();
</script>

<!--
  What went wrong with the form, before the form. role="alert" reads it out as
  soon as it appears, and after a full page load without JavaScript it is the
  first thing on the page; each link moves to the field it names.
  It is drawn in the design system's error tones.
-->
<div
  role="alert"
  tabindex="-1"
  data-error-summary
  class="flex gap-400 rounded-300 border border-stroke-error bg-visual-light-red p-500"
>
  <CircleAlert size={20} class="mt-[0.125rem] shrink-0 text-text-semantic-error" />
  <div class="flex flex-col gap-200">
    <p class="ts-primary-subheading-s text-text-primary-default">{title}</p>
    {#if detail}
      <p class="ts-primary-body-m text-text-primary-default">{detail}</p>
    {/if}
    {#if items.length > 0}
      <ul class="flex flex-col gap-100">
        {#each items as item (item.href)}
          <li>
            <!-- eslint-disable svelte/no-navigation-without-resolve -- a #fragment on this page, not a route -->
            <a
              href={item.href}
              class="ts-primary-body-m text-text-semantic-error underline underline-offset-2 hover:no-underline"
            >
              {item.text}
            </a>
            <!-- eslint-enable svelte/no-navigation-without-resolve -->
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>
