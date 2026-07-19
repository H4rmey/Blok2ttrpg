const puppeteer = require('puppeteer-core');

async function runTest() {
  const browser = await puppeteer.launch({
    executablePath: '/usr/bin/google-chrome',
    args: ['--no-sandbox', '--headless=new', '--disable-gpu'],
    headless: 'new'
  });

  const page = await browser.newPage();

  await page.goto('http://localhost:18081/characters/4c6132bedf8d702b/abilities/new', { waitUntil: 'networkidle2' });
  await page.waitForTimeout(1000);

  // Step 5: Screenshot before any changes
  await page.screenshot({ path: '/tmp/kilo/step_before.png' });
  console.log('Screenshot saved: /tmp/kilo/step_before.png');

  // Step 2: Select ability type "Execution"
  await page.select('#ability-type-select', 'Execution');
  await page.waitForTimeout(500);

  // Step 3: Click "+ Add Enactment"
  await page.click('#add-enactment-btn');
  await page.waitForTimeout(1000);

  // Step 4: Change enact type to "Enact Persistent Effect"
  const enactSelect = await page.$('.enact-type-select');
  if (enactSelect) {
    await enactSelect.select('Enact Persistent Effect');
    await page.waitForTimeout(1000);
  }
  await page.screenshot({ path: '/tmp/kilo/step_enact_persistent.png' });
  console.log('Screenshot saved: /tmp/kilo/step_enact_persistent.png');

  // Count solutions before
  let solRows = await page.$$('[data-generic-list="solution"] [data-row]');
  let solCountBefore = solRows.length;

  // Check solution select options
  const solSelect = await page.$('[data-generic-list="solution"] select');
  let solOptionsBefore = [];
  if (solSelect) {
    const opts = await solSelect.$$('option');
    for (let o of opts) {
      const t = await page.evaluate(el => el.textContent, o);
      if (t && !t.includes('-- Select')) solOptionsBefore.push(t);
    }
    console.log(`Solution dropdown: ${solOptionsBefore.length} options - ${solOptionsBefore.join(', ')}`);
  }

  // Step 6: Click "+ Solution" button
  await page.evaluate(() => {
    const path = document.querySelector('[data-generic-list="solution"]');
    if (path) {
      const header = path.previousElementSibling;
      if (header) {
        const btn = header.querySelector('button');
        if (btn) btn.click();
      }
    }
  });
  await page.waitForTimeout(500);

  solRows = await page.$$('[data-generic-list="solution"] [data-row]');
  let solCountAfter = solRows.length;
  console.log(`\nSolutions add: clicked once → expected ${solCountBefore + 1}, actual ${solCountAfter}`);
  await page.screenshot({ path: '/tmp/kilo/step_solutions_added.png' });
  console.log('Screenshot saved: /tmp/kilo/step_solutions_added.png');

  // Step 7: Click remove on a solution row
  const solRemoveBtns = await page.$$('[data-generic-list="solution"] [data-row] button');
  if (solRemoveBtns.length > 0) {
    await solRemoveBtns[0].click();
    await page.waitForTimeout(500);
  }

  solRows = await page.$$('[data-generic-list="solution"] [data-row]');
  let solCountRemoved = solRows.length;
  console.log(`Solutions remove: clicked once → expected ${solCountAfter - 1}, actual ${solCountRemoved}`);
  await page.screenshot({ path: '/tmp/kilo/step_solutions_removed.png' });
  console.log('Screenshot saved: /tmp/kilo/step_solutions_removed.png');

  // Step 8: Inspect counter_trait dropdown in validation card
  const validationCard = await page.$('.validation-card');
  if (validationCard) {
    await validationCard.evaluate(el => el.scrollIntoView());
  }
  await page.waitForTimeout(500);

  let counterRows = await page.$$('[data-generic-list="counter_trait"] [data-row]');
  let counterCountBefore = counterRows.length;

  const counterTraitSelect = await page.$('[data-generic-list="counter_trait"] select[name*="counter_trait__value"]');
  let counterOptions = [];
  if (counterTraitSelect) {
    const opts = await counterTraitSelect.$$('option');
    for (let o of opts) {
      const t = await page.evaluate(el => el.textContent, o);
      if (t && !t.includes('-- Select')) counterOptions.push(t);
    }
    console.log(`\nCounter trait dropdown: ${counterOptions.length} <option> elements`);
    console.log('Option values:', counterOptions);
  }

  // Step 9: Click "+ Counter" button
  await page.evaluate(() => {
    const path = document.querySelector('[data-generic-list="counter_trait"]');
    if (path) {
      const header = path.previousElementSibling;
      if (header) {
        const btn = header.querySelector('button');
        if (btn) btn.click();
      }
    }
  });
  await page.waitForTimeout(500);

  counterRows = await page.$$('[data-generic-list="counter_trait"] [data-row]');
  let counterCountAfter = counterRows.length;
  console.log(`Counter add: clicked once → expected ${counterCountBefore + 1}, actual ${counterCountAfter}`);
  await page.screenshot({ path: '/tmp/kilo/step_counter_added.png' });
  console.log('Screenshot saved: /tmp/kilo/step_counter_added.png');

  // Step 10: Click remove on a counter row
  const counterRemoveBtns = await page.$$('[data-generic-list="counter_trait"] [data-row] button');
  if (counterRemoveBtns.length > 0) {
    await counterRemoveBtns[counterRemoveBtns.length - 1].click();
    await page.waitForTimeout(500);
  }

  counterRows = await page.$$('[data-generic-list="counter_trait"] [data-row]');
  let counterCountRemoved = counterRows.length;
  console.log(`Counter remove: clicked once → expected ${counterCountAfter - 1}, actual ${counterCountRemoved}`);
  await page.screenshot({ path: '/tmp/kilo/step_after.png' });
  console.log('Screenshot saved: /tmp/kilo/step_after.png');

  // Step 11: Phase knockout cost test
  await page.select('#ability-type-select', 'Phase');
  await page.waitForTimeout(1000);

  const koRows = await page.$$('[data-generic-list="knockout"] [data-row]');
  console.log(`\nKnockout rows in Phase: ${koRows.length}`);

  const buildCost = await page.$eval('#total-cost', el => el.textContent);
  console.log(`Build cost after Phase selection: ${buildCost}`);

  const koSelect = await page.$('[data-generic-list="knockout"] select');
  let koOptions = [];
  if (koSelect) {
    const opts = await koSelect.$$('option');
    for (let o of opts) {
      const t = await page.evaluate(el => el.textContent, o);
      if (t && !t.includes('-- Select')) koOptions.push(t);
    }
    console.log(`Knockout dropdown: ${koOptions.length} options`);
    console.log('Values:', koOptions);
  }
  await page.screenshot({ path: '/tmp/kilo/step_phase.png' });
  console.log('Screenshot saved: /tmp/kilo/step_phase.png');

  await browser.close();

  console.log('\n=== SUMMARY ===');
  console.log(`For solutions add: clicked once → expected ${solCountBefore + 1} more row, actual ${solCountAfter - solCountBefore}`);
  console.log(`For solutions remove: clicked once → expected ${solCountAfter - solCountRemoved} less row, actual ${(solCountAfter - solCountRemoved) === 1 ? 'correct' : 'incorrect'}`);
  console.log(`For counter_trait dropdown: ${counterOptions.length} <option> elements (excluding -- Select --)`);
  console.log(`For counter add: clicked once → expected ${counterCountBefore + 1} more row, actual ${counterCountAfter - counterCountBefore}`);
  console.log(`For counter remove: clicked once → expected ${counterCountAfter - counterCountRemoved} less row, actual ${(counterCountAfter - counterCountRemoved) === 1 ? 'correct' : 'incorrect'}`);
}

runTest().catch(console.error);