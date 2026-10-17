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
  <title>Create an Account | Spillway</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-[28rem] flex-col gap-600 px-600 py-1200">
  <div class="flex flex-col gap-200">
    <h1 class="ts-secondary-heading-s text-text-primary-default">Create an Account</h1>
    <p class="ts-primary-body-m text-text-secondary-default">
      An account keeps your alerts. It needs an email address and a password, nothing else.
    </p>
  </div>

  {#if problems.length > 0}
    <ErrorSummary title="Check the form" items={problems} />
  {/if}

  <form
    method="POST"
    novalidate
    use:enhance={submit}
    class="flex flex-col gap-700 rounded-300 border border-stroke-neutral-default bg-surface-container-primary p-600"
  >
    <TextField
      id={ids.email}
      label="Email address"
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
      label="Password"
      name="password"
      type="password"
      autocomplete="new-password"
      minlength={limits.passwordMinLength}
      maxlength={limits.passwordMaxLength}
      required
      hint={passwordHint}
      error={accountError('password', form?.errors.password)}
    />
    <Button type="submit" aria-disabled={pending || undefined}>
      {pending ? 'Creating Account…' : 'Create Account'}
    </Button>
  </form>

  <p class="ts-primary-body-m text-text-secondary-default">
    Already have an account?
    <a
      href={redirectTo === null
        ? resolve('/sign-in')
        : resolve(`/sign-in?${new URLSearchParams({ redirectTo }).toString()}`)}
      class="ts-primary-subheading-s text-text-primary-link hover:text-text-primary-link-hover"
    >
      Sign in
    </a>
  </p>
</div>
