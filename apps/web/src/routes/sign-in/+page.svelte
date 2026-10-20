<script lang="ts">
  import ErrorSummary from '#lib/components/ui/ErrorSummary.svelte';
  import PasswordField from '#lib/components/ui/PasswordField.svelte';
  import TextField from '#lib/components/ui/TextField.svelte';
  import type { FieldFor } from '#lib/features/account/account.ts';
  import AccountForm from '#lib/features/account/AccountForm.svelte';
  import { m } from '#lib/paraglide/messages.js';

  import type { PageProps } from './$types';

  let { form }: PageProps = $props();
</script>

<svelte:head>
  <title>{m.sign_in_head()} | Spillway</title>
</svelte:head>

<AccountForm
  title={m.sign_in_title()}
  lede={m.sign_in_lede()}
  submitLabel={m.sign_in_title()}
  pendingLabel={m.sign_in_pending()}
  errors={form?.errors}
  other={{ path: '/sign-up', question: m.sign_in_new(), action: m.sign_in_create() }}
>
  {#snippet alert()}
    {#if form?.refused === 'credentials'}
      <ErrorSummary title={m.sign_in_wrong()} detail={m.sign_in_wrong_detail()} />
    {:else if form?.refused === 'throttled'}
      <ErrorSummary
        title={m.sign_in_throttled()}
        detail={form.wait === undefined
          ? m.sign_in_wait()
          : m.sign_in_wait_seconds({ seconds: form.wait })}
      />
    {/if}
  {/snippet}

  {#snippet fields(field: FieldFor)}
    <!-- "username", not "email": what a password manager fills in beside the password. -->
    <TextField
      {...field('email')}
      id="email"
      label={m.form_email()}
      type="email"
      autocomplete="username"
      spellcheck="false"
      value={form?.values.email}
      required
    />
    <PasswordField
      {...field('password')}
      id="password"
      label={m.form_password()}
      showLabel={m.form_show_password()}
      autocomplete="current-password"
      required
    />
  {/snippet}
</AccountForm>
