// The numbers and dates of the pages. Days arrive as 'YYYY-MM-DD', calendar
// days in Brasília, and are written as dates, with no zone to move them;
// instants arrive in UTC and are written in Brasília's time, the ONS's clock.
// The formats are fixed, so the server and the browser write the same text and
// hydration finds nothing to change.

const numbers = (digits: number) =>
  new Intl.NumberFormat('en-US', { minimumFractionDigits: digits, maximumFractionDigits: digits });
const whole = numbers(0);
const tenths = numbers(1);

/** "1,234": rounded, grouped in thousands. */
export const int = (n: number) => whole.format(Math.round(n));

/** "12.3": one decimal. */
export const one = (n: number) => tenths.format(n);

/** MWh written as GWh: one decimal under a hundred, none above. */
export const gwh = (mwh: number) => (mwh / 1000 >= 100 ? int(mwh / 1000) : one(mwh / 1000));

/** MWh written as TWh, with one decimal. */
export const twh = (mwh: number) => (mwh > 0 && mwh < 50_000 ? 'Under 0.1' : one(mwh / 1e6));

/** Energy for a label: none, MWh under a gigawatt hour, GWh above. */
export function energy(mwh: number): string {
  if (mwh < 0.5) return 'None';
  return mwh < 1000 ? `${int(mwh)} MWh` : `${gwh(mwh)} GWh`;
}

/** "1,500 MWh", or "Under 1 MWh" for the little that rounds to nothing. */
export const mwh = (value: number) =>
  value > 0 && value < 0.5 ? 'Under 1 MWh' : `${int(value)} MWh`;

/** A share: one decimal under 10%, none above. */
export function pct(share: number): string {
  const p = share * 100;
  return `${p >= 10 ? int(p) : one(p)}%`;
}

/** What was lost over what could have been generated. */
export const shareOf = (e: { curtailedMwh: number; generatedMwh: number }) =>
  e.curtailedMwh + e.generatedMwh > 0 ? e.curtailedMwh / (e.curtailedMwh + e.generatedMwh) : 0;

/** How a period compares with the one before, in words. */
export function compared(now: number, before: number, name: string): string {
  if (before < 0.5) return now < 0.5 ? `None, as on ${name}` : `None on ${name}`;
  const ratio = now / before;
  if (ratio >= 2) return `${one(ratio)} times as much as ${name}`;
  if (ratio >= 1) return `${int((ratio - 1) * 100)}% more than ${name}`;
  return `${int((1 - ratio) * 100)}% less than ${name}`;
}

const dayFormat = (options: Intl.DateTimeFormatOptions) =>
  new Intl.DateTimeFormat('en-US', { timeZone: 'UTC', ...options });
const longDayFormat = dayFormat({
  weekday: 'long',
  month: 'long',
  day: 'numeric',
  year: 'numeric',
});
const weekdayFormat = dayFormat({ weekday: 'long' });
const weekdayShortFormat = dayFormat({ weekday: 'short' });
const shortDayFormat = dayFormat({ month: 'short', day: 'numeric' });
const shortDayYearFormat = dayFormat({ month: 'short', day: 'numeric', year: 'numeric' });
const dayAndMonthFormat = dayFormat({ month: 'long', day: 'numeric' });
const monthFormat = dayFormat({ month: 'long' });
const monthYearFormat = dayFormat({ month: 'long', year: 'numeric' });
const shortMonthYearFormat = dayFormat({ month: 'short', year: 'numeric' });

const asDate = (day: string) => new Date(`${day}T00:00:00Z`);

/** "Friday, October 16, 2026" */
export const longDay = (day: string) => longDayFormat.format(asDate(day));
/** "Friday" */
export const weekday = (day: string) => weekdayFormat.format(asDate(day));
/** "Fri" */
export const weekdayShort = (day: string) => weekdayShortFormat.format(asDate(day));
/** "Oct 16" */
export const shortDay = (day: string) => shortDayFormat.format(asDate(day));
/** "Oct 16, 2026" */
export const shortDayYear = (day: string) => shortDayYearFormat.format(asDate(day));
/** "October 16" */
export const dayAndMonth = (day: string) => dayAndMonthFormat.format(asDate(day));
/** "October" */
export const monthName = (day: string) => monthFormat.format(asDate(day));
/** "October 2026" */
export const monthYear = (day: string) => monthYearFormat.format(asDate(day));
/** "Oct 2026" */
export const shortMonthYear = (day: string) => shortMonthYearFormat.format(asDate(day));
/** The day of the month, 16. */
export const dayOfMonth = (day: string) => Number(day.slice(8, 10));

const clock = new Intl.DateTimeFormat('en-GB', {
  hour: '2-digit',
  minute: '2-digit',
  hourCycle: 'h23',
  timeZone: 'America/Sao_Paulo',
});
const brasiliaDay = new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Sao_Paulo' });

/** "13:30", an instant on Brasília's clock. */
export const time = (instant: string) => clock.format(new Date(instant));

/** The day an instant falls on in Brasília, as 'YYYY-MM-DD'. */
export const dayOf = (instant: string) => brasiliaDay.format(new Date(instant));

/** "ONS data to October 16, 23:30": how far the figures go. */
export const publishedThrough = (instant: string) =>
  `ONS data to ${dayAndMonth(dayOf(instant))}, ${time(instant)}`;
