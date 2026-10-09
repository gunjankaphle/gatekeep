import { test, expect } from '@playwright/test';

test('reads API data across dashboard, logs, roles and comparison', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', error => errors.push(error.message));
  await page.route('**/api/**', async route => {
    const url = new URL(route.request().url());
    const run = { id: 1, sync_id: 'test', status: 'success', started_at: '2026-10-08T12:00:00Z', total_operations: 1, successful_operations: 1, failed_operations: 0, duration_ms: 50 };
    const operation = { id: 1, operation_type: 'GRANT_OBJECT', target_object: 'LIVE_DB.PUBLIC.EVENTS', sql_statement: 'GRANT SELECT ON TABLE LIVE_DB.PUBLIC.EVENTS TO ROLE LIVE_A', status: 'success', executed_at: run.started_at, execution_time_ms: 50 };
    const responses: Record<string, unknown> = {
      '/api/health': { status: 'healthy', services: { database: 'ok' } },
      '/api/sync/history': { sync_runs: [run], page: 1, page_size: 100, total_count: 1 },
      '/api/sync/history/1': { ...run, operations: [operation] },
      '/api/roles/hierarchy': { roles: [{ name: 'LIVE_A', parent_roles: [] }, { name: 'LIVE_B', parent_roles: ['LIVE_A'] }], pagination: { total_pages: 1 } },
      '/api/roles/refresh/status': { status: 'success', last_refresh: run.started_at, error_message: null },
      '/api/roles/compare': { diff: { added: [], removed: [], unchanged: [{ granted_on: 'TABLE', object_name: operation.target_object, privilege: 'SELECT' }] } },
    };
    await route.fulfill({ json: responses[url.pathname] ?? {} });
  });
  await page.goto('/');
  await expect(page.getByText('Sync #1', { exact: true })).toBeVisible();
  await page.goto('/logs');
  await expect(page.getByText('LIVE_DB.PUBLIC.EVENTS', { exact: true })).toBeVisible();
  await page.goto('/roles');
  await page.getByRole('button', { name: 'Tree View' }).click();
  await expect(page.getByRole('button', { name: 'LIVE_A', exact: true })).toBeVisible();
  await page.goto('/diff');
  await page.locator('select').nth(0).selectOption('LIVE_A');
  await page.locator('select').nth(1).selectOption('LIVE_B');
  await expect(page.getByText('LIVE_DB.PUBLIC.EVENTS', { exact: true })).toBeVisible();
  expect(errors).toEqual([]);
});

test('shows missing database error instead of mock history', async ({ page }) => {
  await page.route('**/api/**', route => route.fulfill({ status: 503, json: { message: 'PostgreSQL is not configured; audit history unavailable' } }));
  await page.goto('/logs');
  await expect(page.getByRole('alert')).toHaveText('PostgreSQL is not configured; audit history unavailable');
});
