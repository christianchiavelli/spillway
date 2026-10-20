<script lang="ts">
  import PasswordField from '#lib/components/ui/PasswordField.svelte';
  import TextField from '#lib/components/ui/TextField.svelte';
  import { type FieldFor, limits, passwordHint } from '#lib/features/account/account.ts';
  import AccountForm from '#lib/features/account/AccountForm.svelte';
  import { m } from '#lib/paraglide/messages.js';

  import type { PageProps } from './$types';

  let { form }: PageProps = $props();
</script>

<svelte:head>
  <title>{m.sign_up_head()} | Spillway</title>
</svelte:head>

<AccountForm
  title={m.sign_up_title()}
  lede={m.sign_up_lede()}
  submitLabel={m.sign_up_submit()}
  pendingLabel={m.sign_up_pending()}
  errors={form?.errors}
  other={{ path: '/sign-in', question: m.sign_up_have_one(), action: m.sign_up_sign_in() }}
>
  {#snippet fields(field: FieldFor)}
    <!-- "username", not "email": what a password manager saves beside the new password. -->
    <TextField
      {...field('email')}
      id="email"
      label={m.form_email()}
      type="email"
      autocomplete="username"
      spellcheck="false"
      maxlength={limits.emailMaxLength}
      value={form?.values.email}
      required
    />
    <PasswordField
      {...field('password')}
      id="new-password"
      label={m.form_password()}
      showLabel={m.form_show_password()}
      autocomplete="new-password"
      minlength={limits.passwordMinLength}
      maxlength={limits.passwordMaxLength}
      required
      hint={passwordHint()}
    />
  {/snippet}
</AccountForm>
