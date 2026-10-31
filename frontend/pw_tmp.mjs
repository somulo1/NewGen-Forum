import { chromium } from 'playwright';
const b = await chromium.launch({ executablePath: '/usr/bin/chromium' });
for (const [w,h] of [[1400,900],[390,800]]) {
const pg = await b.newPage({ viewport: { width: w, height: h } });
await pg.goto('http://localhost:3000/login'); await pg.waitForTimeout(2000);
await pg.fill('input[type=email],input[name=email]', 'samwel.omulo@company.com');
await pg.fill('input[type=password]', 'Hub!Demo2026');
await pg.keyboard.press('Enter'); await pg.waitForTimeout(6000);
console.log(w, pg.url());
await pg.screenshot({ path: `/tmp/claude-1000/-home-vaultum-Documents-NewGen-Forum/31f56a7c-7446-444a-84f2-d985d775dc41/scratchpad/a${w}.png` });
await pg.close(); }
await b.close();
