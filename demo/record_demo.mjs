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
await filesPage.waitForTimeout(1800);

const grafanaPage = await context.newPage();
await grafanaPage.goto(`${grafanaURL}/dashboards`, { waitUntil: "domcontentloaded" });
await grafanaPage.waitForTimeout(1000);

const dashboard = await firstDashboard(grafanaPage);
await grafanaPage.goto(`${grafanaURL}${dashboard.url}?orgId=1&from=now-6h&to=now&timezone=browser`, {
  waitUntil: "domcontentloaded",
});

await grafanaPage.waitForFunction(
  (title) => document.body.innerText.includes(title),
  dashboard.title,
  { timeout: 60000 },
);
await grafanaPage.bringToFront();
await grafanaPage.waitForTimeout(5500);
await filesPage.bringToFront();
await filesPage.waitForTimeout(1700);
await grafanaPage.bringToFront();
await grafanaPage.waitForTimeout(5500);
await browser.close();

async function firstDashboard(page) {
  for (let i = 0; i < 60; i++) {
    const response = await page.request.get(`${grafanaURL}/api/search?type=dash-db`);
    const dashboards = await response.json();
    const dashboard = dashboards.find((item) => item.type === "dash-db" && item.url);
    if (dashboard) return dashboard;
    await page.waitForTimeout(1000);
  }
  throw new Error("no provisioned dashboards found");
}
