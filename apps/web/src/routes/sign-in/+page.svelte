<script lang="ts">
  import { enhance } from '$app/forms';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import Button from '#lib/components/ui/Button.svelte';
  import ErrorSummary from '#lib/components/ui/ErrorSummary.svelte';
  import TextField from '#lib/components/ui/TextField.svelte';
  import { accountError, accountFields } from '#lib/features/account/account.ts';
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
  const ids = { email: 'email', password: 'password' } as const;
  const problems = $derived(
    accountFields.flatMap((field) => {
      const text = accountError(field, form?.errors?.[field]);
      return text === undefined ? [] : [{ href: `#${ids[field]}`, text }];
    }),
  );
</script>

<svelte:head>
  <title>{m.sign_in_head()} | Spillway</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-[28rem] flex-col gap-600 px-600 py-1200">
  <div class="flex flex-col gap-200">
    <h1 class="ts-secondary-heading-s text-text-primary-default">{m.sign_in_title()}</h1>
    <p class="ts-primary-body-m text-text-secondary-default">
      {m.sign_in_lede()}
    </p>
  </div>

  {#if form?.refused === 'credentials'}
    <ErrorSummary title={m.sign_in_wrong()} detail={m.sign_in_wrong_detail()} />
  {:else if form?.refused === 'throttled'}
    <ErrorSummary
      title={m.sign_in_throttled()}
      detail={form.wait === undefined
        ? m.sign_in_wait()
        : m.sign_in_wait_seconds({ seconds: form.wait })}
    />
  {:else if problems.length > 0}
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
      value={form?.values.email}
      required
      error={accountError('email', form?.errors?.email)}
    />
    <TextField
      id={ids.password}
      label={m.form_password()}
      name="password"
      type="password"
      autocomplete="current-password"
      required
      error={accountError('password', form?.errors?.password)}
    />
    <Button type="submit" aria-disabled={pending || undefined}>
      {pending ? m.sign_in_pending() : m.sign_in_title()}
    </Button>
  </form>

  <!-- The query goes along, so signing up still leads to the page asked for. -->
  <p class="ts-primary-body-m text-text-secondary-default">
    {m.sign_in_new()}
    <a
      href={localized(
        redirectTo === null
          ? resolve('/sign-up')
          : resolve(`/sign-up?${new URLSearchParams({ redirectTo }).toString()}`),
      )}
      class="ts-primary-subheading-s text-text-primary-link hover:text-text-primary-link-hover"
    >
      {m.sign_in_create()}
    </a>
  </p>
</div>
