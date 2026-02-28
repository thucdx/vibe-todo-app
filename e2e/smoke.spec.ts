/**
 * Smoke test suite — covers the five critical user flows:
 *   1. First-time PIN setup
 *   2. Login with the correct PIN
 *   3. Create a task
 *   4. Toggle a task done/undone
 *   5. Delete a task
 *
 * Prerequisites:
 *   - `docker compose up --build` is running (backend + DB + frontend dev server)
 *   - The database is EMPTY (fresh volume) so the PIN has not been set.
 *     Run `docker compose down -v && docker compose up --build` for a clean slate.
 */
import { test, expect } from '@playwright/test'

const TEST_PIN = '1234'
const TEST_TASK = 'Smoke test task'

// ─── 1. Setup PIN ──────────────────────────────────────────────────────────────
test.describe('first-time PIN setup', () => {
  test.beforeAll(async ({ request }) => {
    await request.post('/api/v1/test/reset')
  })

  test('shows Create PIN and completes setup', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByText('Create PIN')).toBeVisible()

    await page.getByPlaceholder(/PIN/i).fill(TEST_PIN)
    await page.getByRole('button', { name: /create/i }).click()

    // After setup the login form should appear immediately.
    await expect(page.getByText('Unlock')).toBeVisible()
  })
})

// ─── 2. Login ─────────────────────────────────────────────────────────────────
test('login with correct PIN shows task view', async ({ page }) => {
  await page.goto('/')

  // The PIN was created in the previous test; the app may be fresh or in locked state.
  // If the create-PIN form appears, set it up first.
  const createBtn = page.getByRole('button', { name: /create pin/i })
  if (await createBtn.isVisible().catch(() => false)) {
    await page.getByPlaceholder(/PIN/i).fill(TEST_PIN)
    await createBtn.click()
  }

  await page.getByPlaceholder(/PIN/i).fill(TEST_PIN)
  await page.getByRole('button', { name: /unlock/i }).click()

  await expect(page.getByText('Add task')).toBeVisible()
})

// ─── 3. Create a task ─────────────────────────────────────────────────────────
test('create a task and see it in the list', async ({ page }) => {
  await page.goto('/')
  await loginIfNeeded(page)

  await page.getByText('Add task').click()
  await expect(page.getByText('New Task')).toBeVisible()

  await page.getByPlaceholder(/what needs to be done/i).fill(TEST_TASK)
  await page.getByRole('button', { name: /create/i }).click()

  await expect(page.getByText(TEST_TASK).first()).toBeVisible()
})

// ─── 4. Toggle task done ───────────────────────────────────────────────────────
test('toggle task done and undone', async ({ page }) => {
  await page.goto('/')
  await loginIfNeeded(page)

  // Make sure the task exists (idempotent create).
  await ensureTask(page, TEST_TASK)

  const checkbox = page.locator(`text=${TEST_TASK}`).first().locator('..').getByRole('checkbox')
  await expect(checkbox).not.toBeChecked()

  await checkbox.click()
  await expect(checkbox).toBeChecked()

  await checkbox.click()
  await expect(checkbox).not.toBeChecked()
})

// ─── 5. Delete a task ─────────────────────────────────────────────────────────
test('delete task removes it from the list', async ({ page }) => {
  await page.goto('/')
  await loginIfNeeded(page)

  await ensureTask(page, TEST_TASK)

  // Accept the confirm dialog automatically.
  page.on('dialog', dialog => dialog.accept())

  const countBefore = await page.getByText(TEST_TASK).count()
  const taskRow = page.locator(`text=${TEST_TASK}`).first().locator('..')
  await taskRow.hover()
  await taskRow.getByTitle('Delete').click()

  await expect(page.getByText(TEST_TASK)).toHaveCount(countBefore - 1)
})

// ─── Helpers ─────────────────────────────────────────────────────────────────

async function loginIfNeeded(page: import('@playwright/test').Page) {
  const isLoggedIn = await page.getByText('Add task').isVisible().catch(() => false)
  if (isLoggedIn) return

  // Set up PIN if not yet configured.
  const needsSetup = await page.getByText('Create PIN').isVisible().catch(() => false)
  if (needsSetup) {
    await page.getByPlaceholder(/PIN/i).fill(TEST_PIN)
    await page.getByRole('button', { name: /create/i }).click()
  }

  await page.getByPlaceholder(/PIN/i).fill(TEST_PIN)
  await page.getByRole('button', { name: /unlock/i }).click()
  await page.getByText('Add task').waitFor()
}

async function ensureTask(page: import('@playwright/test').Page, title: string) {
  const count = await page.getByText(title).count()
  if (count > 0) return

  await page.getByText('Add task').click()
  await page.getByPlaceholder(/what needs to be done/i).fill(title)
  await page.getByRole('button', { name: /create/i }).click()
  await page.getByText(title).first().waitFor()
}
