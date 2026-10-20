<script lang="ts">
  import type { SubmitFunction } from '$app/forms';
  import { enhance } from '$app/forms';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import Button from '#lib/components/ui/Button.svelte';
  import {
    accountError,
    type AccountField,
    accountFields,
    type FieldFor,
  } from '#lib/features/account/account.ts';
  import { enhanceForm } from '#lib/forms/enhance.ts';
  import { type FieldErrors, firstInvalid, fromValidity } from '#lib/forms/errors.ts';
  import { localized } from '#lib/i18n.ts';
  import type { Snippet } from 'svelte';

  type Props = {
    title: string;
    lede: string;
    submitLabel: string;
    /** The button's label while the form is sent. */
    pendingLabel: string;
    /** The server's verdict on each field, from its last answer. */
    errors: FieldErrors<AccountField> | undefined;
    /** The other way in, sign-up from sign-in and back, and the line that leads there. */
    other: { path: '/sign-in' | '/sign-up'; question: string; action: string };
    /** A refusal of the whole form, such as a wrong password, above the fields. */
    alert?: Snippet | undefined;
    fields: Snippet<[FieldFor]>;
  };

  let { title, lede, submitLabel, pendingLabel, errors, other, alert, fields }: Props = $props();
  const headingId = $props.id();

  // The browser's verdict on each field it has checked since the server last
  // answered, null for a field found right: a field checked here shows its
  // own, any other the server's. A new answer brings a new errors object,
  // which starts this over.
  let judged = $derived({
    answer: errors,
    verdicts: {} as Partial<Record<AccountField, string | null>>,
  });

  const shown = (name: AccountField) =>
    name in judged.verdicts
      ? (judged.verdicts[name] ?? undefined)
      : accountError(name, errors?.[name]);

  function judge(name: AccountField, input: HTMLInputElement) {
    const verdict = accountError(name, fromValidity(input.validity));
    judged = { ...judged, verdicts: { ...judged.verdicts, [name]: verdict ?? null } };
    return verdict;
  }

  // Without JavaScript, a refused form loads again with the first wrong field
  // focused, where the script would have put it.
  const firstWrong = $derived(firstInvalid(errors, accountFields));

  // Nothing is judged while it is typed the first time: a field is checked
  // when it is left, unless it was left empty on the way through, and only
  // once it shows an error does each keystroke check it again, so the error
  // goes as soon as it is put right.
  const field: FieldFor = (name) => ({
    name,
    error: shown(name),
    autofocus: name === firstWrong,
    onblur: (event) => {
      if (event.currentTarget.value !== '' || shown(name) !== undefined) {
        judge(name, event.currentTarget);
      }
    },
    oninput: (event) => {
      if (shown(name) !== undefined) judge(name, event.currentTarget);
    },
  });

  let sending = $state(false);
  const send = enhanceForm({
    onpending: (pending) => {
      sending = pending;
    },
  });

  // Checked here before it goes, with the rules the fields carry from the
  // contract: a form the server would refuse is not sent, and focus goes to
  // the first field to put right. The server checks it again either way.
  const submit: SubmitFunction = (input) => {
    let first: HTMLInputElement | undefined;
    for (const name of accountFields) {
      const element = input.formElement.elements.namedItem(name);
      if (element instanceof HTMLInputElement && judge(name, element) !== undefined) {
        first ??= element;
      }
    }
    if (first !== undefined) {
      input.cancel();
      first.focus();
      return;
    }
    return send(input);
  };

  // The query goes along, so the other way in still leads to the page asked for.
  const otherHref = $derived.by(() => {
    const redirectTo = page.url.searchParams.get('redirectTo');
    return localized(
      redirectTo === null
        ? resolve(other.path)
        : resolve(`${other.path}?${new URLSearchParams({ redirectTo }).toString()}`),
    );
  });
</script>

<!--
  Sign-in and sign-up, one form drawn the same way: the title and its line,
  the fields and the button in one card, and under it, centred, the way to the
  other one.
-->
<div
  class="mx-auto flex w-full max-w-[28rem] flex-col gap-600 px-500 py-1000 mobile:px-600 mobile:py-1200"
>
  <form
    method="POST"
    novalidate
    use:enhance={submit}
    aria-labelledby={headingId}
    class="flex flex-col gap-600 rounded-300 border border-stroke-neutral-default bg-surface-container-primary p-600 mobile:p-700"
  >
    <div class="flex flex-col gap-300">
      <h1 id={headingId} class="ts-secondary-heading-s text-text-primary-default">{title}</h1>
      <p class="ts-primary-body-m text-pretty text-text-secondary-default">{lede}</p>
    </div>

    {@render alert?.()}

    <div class="flex flex-col gap-600">
      {@render fields(field)}
    </div>

    <Button type="submit" size="block" pending={sending}>
      {sending ? pendingLabel : submitLabel}
    </Button>
  </form>

  <p class="text-center ts-primary-body-l text-text-secondary-default">
    {other.question}
    <a
      href={otherHref}
      class="rounded-100 ts-primary-subheading-m whitespace-nowrap text-text-primary-link underline-offset-4 hover:text-text-primary-link-hover hover:underline active:text-text-primary-default"
    >
      {other.action}
    </a>
  </p>
</div>
