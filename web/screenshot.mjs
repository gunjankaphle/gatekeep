import { chromium } from 'playwright';

async function takeScreenshots() {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.setViewportSize({ width: 1920, height: 1080 });

  const pages = [
    { name: 'dashboard', url: 'http://localhost:5173/' },
    { name: 'logs', url: 'http://localhost:5173/logs' },
    { name: 'roles', url: 'http://localhost:5173/roles' },
    { name: 'diff', url: 'http://localhost:5173/diff' },
  ];

  for (const { name, url } of pages) {
    console.log(`Capturing ${name}...`);
    await page.goto(url, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1000); // Wait for animations
    await page.screenshot({
      path: `screenshots/${name}.png`,
      fullPage: true
    });
    console.log(`✓ Saved screenshots/${name}.png`);
  }

  await browser.close();
  console.log('\n✅ All screenshots captured!');
}

takeScreenshots().catch(console.error);
