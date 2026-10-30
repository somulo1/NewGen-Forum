import { chromium } from 'playwright';
const b = await chromium.launch({ executablePath: '/usr/bin/chromium', args: ['--no-sandbox'] });
for (const [name, vp, mode] of [['full', { width: 1500, height: 900 }, 'Full screen'], ['mobile', { width: 420, height: 860 }, null]]) {
  const p = await b.newPage({ viewport: vp });
  await p.goto('http://localhost:3000/login');
  await p.waitForTimeout(1500);
  await p.fill('input[type=email]', 'rights-tester@example.com');
  await p.fill('input[type=password]', 'password');
  await p.keyboard.press('Enter');
  await p.waitForTimeout(9000);
  await p.click('button[aria-label="Forums0 new"], button:has-text("FORUMS")').catch(() => {});
  await p.waitForTimeout(600);
  await p.locator('nav button:has-text("Caption test"), aside button:has-text("Caption test")').first().click().catch(() => console.log('row fail'));
  await p.waitForTimeout(2500);
  if (mode) {
    await p.click('button[aria-label="Expand chat"]');
    await p.click(`[role=menuitem]:has-text("${mode}")`);
  }
  await p.waitForTimeout(2500);
  await p.screenshot({ path: `/tmp/chat-${name}.png` });
  await p.close();
}
await b.close();
