/**
 * Waits for the sync service to load every month: on a fresh stack the data
 * pages say the ONS's data is on its way for a minute or two, and the tests
 * are about the pages with data in them. The first file is October 2021's.
 */
export default async function waitForData() {
  const base = process.env['E2E_BASE_URL'] ?? 'http://localhost:8090';
  const deadline = Date.now() + 10 * 60 * 1000;
  for (;;) {
    const page = await fetch(`${base}/history`)
      .then((r) => r.text())
      .catch(() => '');
    if (page.includes('2021, from October')) return;
    if (Date.now() > deadline) throw new Error(`No data at ${base} after ten minutes.`);
    await new Promise((resolve) => setTimeout(resolve, 5000));
  }
}
