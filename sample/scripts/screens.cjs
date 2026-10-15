const { chromium } = require('playwright');
(async () => {
  const b = await chromium.launch();
  const errs = [];
  const nav = (t) => async (p) => { await p.click(`nav[aria-label=Main] button:has-text("${t}")`); await p.waitForTimeout(500); };
  const shots = [
    ['01-home', null], ['02-community-wall', nav('Community wall')], ['03-meetings', nav('Meetings')], ['04-calendar', nav('Calendar')],
    ['05-communications', nav('Communications')], ['06-tickets', nav('Tickets')],
    ['07-ticket-details', async (p) => { await nav('Tickets')(p); await p.locator('main button:has-text("View")').first().click(); await p.waitForTimeout(500); }],
    ['08-projects', nav('Projects')],
    ['09-project-view', async (p) => { await nav('Projects')(p); await p.locator('main button:has-text("View")').first().click(); await p.waitForTimeout(500); }],
    ['10-people', nav('People')], ['11-approvals', nav('Approvals')], ['12-security', nav('Security')], ['13-roles', nav('Roles')],
    ['14-profile', async (p) => { await p.click('header button[aria-label="My profile"]'); await p.waitForTimeout(500); }],
    ['15-forum-full', async (p) => { await p.click('button[aria-label="Open forum full screen"]'); await p.waitForTimeout(500); }],
    ['16-private-chat', async (p) => { await p.click('nav[aria-label=Main] button:has-text("Brian Kiprop")'); await p.waitForTimeout(500); }],
    ['17-add-widget', async (p) => { await p.click('main button:has-text("Add widget")'); await p.waitForTimeout(500); }],
    ['18-department-dashboard', async (p) => { await p.locator('nav[aria-label=Main] button:has-text("My department")').first().click(); await p.waitForTimeout(500); }]
  ];
  for (const [vw, vh, tag] of [[1440, 900, 'desktop'], [390, 844, 'mobile']]) {
    for (const [name, fn] of shots) {
      const p = await b.newPage({ viewport: { width: vw, height: vh } });
      p.on('pageerror', (e) => errs.push(tag + name + ': ' + e.message));
      await p.goto('process.env.HUB_URL || 'http://localhost:3000/'', { waitUntil: 'networkidle' });
      await p.waitForTimeout(500);
      try {
        if (fn && tag === 'mobile' && /nav\[/.test(fn.toString() + '') ) {}
        if (fn) { if (tag === 'mobile') { const m = p.locator('button[aria-label="Open menu"]'); if (await m.isVisible()) { await m.click(); await p.waitForTimeout(350); } } await fn(p); }
      } catch (e) { errs.push(tag + name + ' step failed: ' + e.message.split('\n')[0]); }
      await p.screenshot({ path: `screens/${tag}-${name}.png` });
      await p.close();
    }
  }
  console.log(errs.join('\n') || 'no errors');
  await b.close();
})();
