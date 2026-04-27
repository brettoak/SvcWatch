const puppeteer = require('puppeteer');

(async () => {
  const browser = await puppeteer.launch();
  const page = await browser.newPage();
  
  page.on('console', msg => {
    if (msg.type() === 'error' || msg.type() === 'warning') {
      console.log(msg.text());
    }
  });

  page.on('pageerror', err => {
    console.log('Page Error: ', err.toString());
  });

  await page.goto('http://localhost:5173/login');
  await page.type('input[type="email"]', 'admin@example.com');
  await page.type('input[type="password"]', 'admin123');
  await Promise.all([
    page.waitForNavigation(),
    page.click('button[type="submit"]') // wait, is the button type submit? Just click the button containing "Sign In"
  ]);

  await page.goto('http://localhost:5173/logs');
  await page.waitForSelector('button');
  
  // click Show Advanced Filters
  const buttons = await page.$$('button');
  for (const btn of buttons) {
    const text = await page.evaluate(el => el.textContent, btn);
    if (text.includes('Show Advanced Filters')) {
      await btn.click();
      break;
    }
  }

  await page.waitForTimeout(2000);
  await browser.close();
})();
