/**
 * One JSON object per line, in the shape the Go API writes with slog, so the
 * two services' logs read and query the same way.
 */
type Level = 'INFO' | 'WARN' | 'ERROR';

function write(level: Level, msg: string, attrs: Record<string, unknown>) {
  const line = JSON.stringify(
    { time: new Date().toISOString(), level, msg, ...attrs },
    (_key, value: unknown) =>
      value instanceof Error
        ? { name: value.name, message: value.message, stack: value.stack }
        : value,
  );
  if (level === 'INFO') console.log(line);
  else console.error(line);
}

export const log = {
  info(msg: string, attrs: Record<string, unknown> = {}) {
    write('INFO', msg, attrs);
  },
  warn(msg: string, attrs: Record<string, unknown> = {}) {
    write('WARN', msg, attrs);
  },
  error(msg: string, attrs: Record<string, unknown> = {}) {
    write('ERROR', msg, attrs);
  },
};
