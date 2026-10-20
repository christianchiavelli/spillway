<script lang="ts">
  import ErrorIcon from '~icons/material-symbols/error-outline-rounded';

  type Props = {
    /** What the field's aria-describedby names. */
    id: string;
    hint?: string | undefined;
    error?: string | undefined;
  };

  let { id, hint, error }: Props = $props();
</script>

<!--
  The line under a field, lined up with what is typed in it. A hint is in the
  body's tone, so it never reads as something wrong; an error replaces it in
  bold red, with an icon, so the two differ by more than colour. The line is
  kept when empty, so a message that comes or goes never moves the form.
-->
<div class="min-h-[1rem] pl-300">
  {#if error !== undefined}
    <p {id} class="flex gap-100 ts-primary-subheading-xs text-text-semantic-error">
      <ErrorIcon width={14} height={14} aria-hidden="true" class="mt-[0.0625rem] shrink-0" />
      {error}
    </p>
  {:else if hint !== undefined}
    <p {id} class="ts-primary-body-s text-text-secondary-default">{hint}</p>
  {/if}
</div>
