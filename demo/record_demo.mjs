import { chromium } from "playwright";

const grafanaURL = process.env.GRAFANA_URL;
const fileURL = process.env.DASHBOARD_FILE_URL || "file:///Users/admin/dashboard/";

if (!grafanaURL) throw new Error("GRAFANA_URL is required");

const browser = await chromium.launch({
  headless: false,
  args: ["--window-size=1440,900", "--start-maximized", "--no-sandbox"],
});
const context = await browser.newContext({
  viewport: { width: 1365, height: 760 },
});

const filesPage = await context.newPage();
await filesPage.goto(fileURL);
await filesPage.bringToFront();
await filesPage.waitForTimeout(900);
await filesPage.getByRole("link", { name: "sre/" }).click();
await filesPage.waitForTimeout(700);
await filesPage.getByRole("link", { name: "cluster-health.json" }).click();
await filesPage.waitForFunction(() => document.body.innerText.includes("SRE Cluster Health"), { timeout: 10000 });
await filesPage.waitForTimeout(1500);

const grafanaPage = await context.newPage();
await grafanaPage.goto(`${grafanaURL}/dashboards`, { waitUntil: "domcontentloaded" });
await grafanaPage.bringToFront();
await grafanaPage.waitForTimeout(800);

await clickText(grafanaPage, "sre");
await grafanaPage.waitForTimeout(900);
await clickText(grafanaPage, "SRE Cluster Health");

await grafanaPage.waitForFunction(
  (title) => document.body.innerText.includes(title),
  "SRE Cluster Health",
  { timeout: 60000 },
);
await grafanaPage.bringToFront();
await grafanaPage.waitForTimeout(5500);
await browser.close();

async function clickText(page, text) {
  for (let i = 0; i < 30; i++) {
    const locator = page.getByText(text, { exact: true }).first();
    if (await locator.count()) {
      await locator.click();
      return;
    }
    await page.waitForTimeout(500);
  }
  throw new Error(`text not found: ${text}`);
}
