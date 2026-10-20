/**
 * Records the data the Web Vitals are measured on: the curtailment tables of
 * the stack `docker compose up` started, once its sync has loaded every
 * month, dumped as SQL into `sums.sql.gz`.
 *
 *   pnpm run vitals:record
 *
 * The ONS's data alone: no account, session or alert leaves the machine. The
 * dump restores into the schema as the migrations leave it, so it is recorded
 * again when one changes these tables, and a baseline accepted after it.
 */
import { spawn } from 'node:child_process';
import { createWriteStream } from 'node:fs';
import { pipeline } from 'node:stream/promises';
import { fileURLToPath } from 'node:url';
import { createGzip } from 'node:zlib';

const TABLES = ['plants', 'plant_days', 'subsystem_days', 'reason_days', 'half_hours', 'ons_files'];

const dump = spawn(
  'docker',
  [
    ...['compose', 'exec', '-T', 'db', 'pg_dump'],
    ...['--username', 'spillway', '--dbname', 'spillway', '--data-only'],
    ...TABLES.flatMap((table) => ['--table', table]),
  ],
  { cwd: fileURLToPath(new URL('../..', import.meta.url)), stdio: ['ignore', 'pipe', 'inherit'] },
);
const exited = new Promise<number | null>((resolve) => dump.on('close', resolve));
await pipeline(
  dump.stdout,
  createGzip({ level: 9 }),
  createWriteStream(new URL('sums.sql.gz', import.meta.url)),
);
const code = await exited;
if (code !== 0) throw new Error(`pg_dump exited with ${String(code)}`);
console.log('Recorded the sums into e2e/vitals/sums.sql.gz.');
