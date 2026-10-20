# 27. Sign-in and sign-up check each field as it is left

- Status: accepted; narrows the error summary of [ADR 0009](0009-a-dense-dashboard-look-accessibility-first.md) to the longer forms
- Date: 2026-10-19

## Context

The two account forms were checked by the server alone. A mistake showed after a round trip, twice: in an error summary at the top of the page, which pushed the form down as it appeared, and again under its field. The password's guidance sat under its field in the small bold type of an error, so it read as one before anything was typed. Each page drew its own copy of the same form, and the line to the other page sat under it, small and to one side.

## Decision

- **One form for both, `AccountForm`.** The title and its line, the fields and a full-width button share one card, and under it, centred, is the way to the other page. A refusal of the whole form, a wrong password or too many attempts, shows inside the card, above the fields.
- **The browser checks a field when it is left.** The fields already carry the contract's rules as attributes, `required`, `type="email"`, `minlength` and `maxlength`; the Constraint Validation API reads them, and each broken rule is named with the server's code, so the words are the same whichever side found it. A field left empty on the way through is not judged. Once a field shows an error, each keystroke checks it again, so the error goes as soon as it is put right. Nothing is judged while it is first typed.
- **A form the browser can tell is wrong is not sent.** Each wrong field says why, and focus goes to the first. The server checks everything again (ADR 0005), and what only it can know, an address already taken or a wrong password, comes back as before.
- **No error summary on these two forms.** With two fields, each error sits beside its field; a summary would only repeat it and move the form. Without JavaScript, the refused page loads with `autofocus` on the first wrong field, where the script would have put focus. The alert form, longer and beside its preview, keeps its summary.
- **A hint looks like a hint:** regular type in the body's grey. An error takes its place in bold red with an icon, so the two differ by more than colour. The password's hint says the one rule the server applies, 15 characters, and a tip, never a rule it does not enforce.
- **The line under each field is kept when empty**, so a message that comes or goes never moves the field below it.
- **A button shows the password**, since 15 characters typed unseen are easy to get wrong. It keeps its name and says whether it is on with `aria-pressed`, and appears once the script runs, since it needs it. The field never spellchecks, so a shown password is not sent to a spelling service, and a refused form clears it, shown or not.
- **A pending button shows a spinner** where its icon goes and refuses a second press with `aria-disabled`, staying in the tab order. Under reduced motion the spinner stands still; the label says the same.
- **`autocomplete="username"` on the address**, which is what a password manager saves and fills in beside the password.

## Consequences

- The rules are applied in three places, the browser, the web server and the API, all from the contract's limits.
- A field's error comes from the browser or from the server. The form keeps the browser's verdict on a field until the server answers again, which starts it over.
- A field with no hint is a line taller, and the alert form grows by that line under its subsystem.
- The text and select fields share the line under them, `FieldMessage`, so the alert form's hints and errors changed with these.

## Alternatives considered

- **Checking each key as it is typed:** an error before the address is finished.
- **The browser's own messages, `reportValidity()`:** bubbles that differ by browser, leave after a moment, and speak the browser's language, not the edition's.
- **Keeping the summary here:** the one or two errors already beside their fields, said again above them, with the form pushed down as it appears.
- **A validation library:** a few rules a field, all of which the browser already checks.
