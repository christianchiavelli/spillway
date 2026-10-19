<script lang="ts">
  import { enhance } from '$app/forms';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import Button from '#lib/components/ui/Button.svelte';
  import ErrorSummary from '#lib/components/ui/ErrorSummary.svelte';
  import TextField from '#lib/components/ui/TextField.svelte';
  import {
    accountError,
    accountFields,
    limits,
    passwordHint,
  } from '#lib/features/account/account.ts';
  import { enhanceForm } from '#lib/forms/enhance.ts';
  import { localized } from '#lib/i18n.ts';
  import { m } from '#lib/paraglide/messages.js';

  import type { PageProps } from './$types';

  let { form }: PageProps = $props();

  let pending = $state(false);
  const submit = enhanceForm({
    onpending: (p) => {
      pending = p;
    },
  });

  const redirectTo = $derived(page.url.searchParams.get('redirectTo'));
  const ids = { email: 'email', password: 'new-password' } as const;
  const problems = $derived(
    accountFields.flatMap((field) => {
      const text = accountError(field, form?.errors[field]);
      return text === undefined ? [] : [{ href: `#${ids[field]}`, text }];
    }),
  );
</script>

<svelte:head>
  <title>{m.sign_up_head()} | Spillway</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-[28rem] flex-col gap-600 px-600 py-1200">
  <div class="flex flex-col gap-200">
    <h1 class="ts-secondary-heading-s text-text-primary-default">{m.sign_up_title()}</h1>
    <p class="ts-primary-body-m text-text-secondary-default">
      {m.sign_up_lede()}
    </p>
  </div>

  {#if problems.length > 0}
    <ErrorSummary title={m.form_check()} items={problems} />
  {/if}

  <form
    method="POST"
    novalidate
    use:enhance={submit}
    class="flex flex-col gap-700 rounded-300 border border-stroke-neutral-default bg-surface-container-primary p-600"
  >
    <TextField
      id={ids.email}
      label={m.form_email()}
      name="email"
      type="email"
      autocomplete="email"
      spellcheck="false"
      maxlength={limits.emailMaxLength}
      value={form?.values.email}
      required
      error={accountError('email', form?.errors.email)}
    />
    <TextField
      id={ids.password}
      label={m.form_password()}
      name="password"
      type="password"
      autocomplete="new-password"
      minlength={limits.passwordMinLength}
      maxlength={limits.passwordMaxLength}
      required
      hint={passwordHint()}
      error={accountError('password', form?.errors.password)}
    />
    <Button type="submit" aria-disabled={pending || undefined}>
      {pending ? m.sign_up_pending() : m.sign_up_submit()}
    </Button>
  </form>

  <p class="ts-primary-body-m text-text-secondary-default">
    {m.sign_up_have_one()}
    <a
      href={localized(
        redirectTo === null
          ? resolve('/sign-in')
          : resolve(`/sign-in?${new URLSearchParams({ redirectTo }).toString()}`),
      )}
      class="ts-primary-subheading-s text-text-primary-link hover:text-text-primary-link-hover"
    >
      {m.sign_up_sign_in()}
    </a>
  </p>
</div>
