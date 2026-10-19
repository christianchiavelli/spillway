# 26. A Portuguese edition, under /pt-br

- Status: accepted
- Date: 2026-10-19

## Context

Spillway reads the data of Brazil's grid operator, and the people who follow curtailment most closely, the plants' owners, the trading desks and the press, read Portuguese first. The app was in English only, every word written into its components.

## Decision

- **English at the root, Portuguese under `/pt-br`.** `/plants` and `/pt-br/plants` are the same page in two languages. The address alone chooses the edition: no cookie and no `Accept-Language`, so a link always opens what it says, a cache needs no `Vary`, and a page shared from one edition opens in it.
- **Paraglide JS 2.29.0**, with the inlang message format 4.5.4, the library SvelteKit's own add-on sets up. Messages live in `messages/en.json` and `messages/pt-BR.json`, compiled at build time into one function each: a page carries only the messages it uses, a message that does not exist fails the type check, and plurals and Portuguese's genders are variants of the message, not code.
- **The edition is read once per request.** `hooks.ts` takes the prefix off before routing, so one route serves both, and a server handle keeps the request's locale for everything it renders, `<html lang>` included. A component never passes the locale on.
- **Labels are functions or getters**, never strings built when a module loads: a module is loaded once for every request, in both editions.
- **Every link, form and redirect goes through `localized()`**, which puts the reader's prefix on a path `resolve()` returned. The address to return to after signing in is kept as asked for, prefix and all.
- **An EN | PT switch in the header**, two links to the same page in each edition, with its query kept. They load the page in full, since the messages are read as a page renders and do not change under it.
- **Numbers and dates as each country writes them**, through `Intl` in `en-US` and `pt-BR`: 1,234.5 and 1.234,5, October 17 and 17 de outubro. Two things are the same in both editions: the clock, 24 hours, and the dates in the address, `2026-10-17`.
- **A threshold is still typed in digits only.** "1.500" is fifteen hundred in Portuguese and one and a half in English; refusing both separators is clearer than guessing.
- **Each page names its other edition** with `hreflang` links, English as the default for every other language.

## Consequences

- A text added to the app is two messages, one per edition. Paraglide would fall back to English for a missing one, so a unit test holds the two files to the same keys.
- Portuguese runs about a fifth longer: labels wrap where the English ones did not, and the screens were checked again at a laptop's width and a phone's.
- Some sentences are written twice rather than translated word by word. Portuguese puts "no" or "na" before a weekday, so a comparison names the day it compares with as "na sexta-feira", and an adjective agrees with its noun, so "usinas eólicas" is its own message.
- The API stays in English: it answers with codes, and the web app words them.
- The switch costs a full page load.

## Alternatives considered

- **Portuguese by `Accept-Language` or a cookie:** an address that opens different pages for different people, which a link, a cache and a crawler all get wrong.
- **A prefix for both, `/en` and `/pt-br`:** every address the app has had would move, for nothing a reader gains.
- **Translation at run time, with a dictionary object:** every message in every page's bundle, and a missing key found by a reader rather than the type check.
- **A domain for each edition:** two hosts to run for a portfolio app that runs on one laptop.
