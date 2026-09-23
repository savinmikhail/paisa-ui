import assert from "node:assert/strict";
import { test } from "node:test";
import dayjs from "dayjs";
import * as d3 from "d3";
import { createServer, type ViteDevServer } from "vite";

let server: ViteDevServer | undefined;

async function selectionToInclusiveDateRange() {
  server ??= await createServer({ configFile: "./vite.config.js", server: { middlewareMode: true } });
  const module = await server.ssrLoadModule("/src/lib/networth.ts");
  return module.selectionToInclusiveDateRange as typeof import("./networth").selectionToInclusiveDateRange;
}

const dates = [dayjs("2024-01-01"), dayjs("2024-01-10"), dayjs("2024-01-20")];
const scale = d3.scaleTime().domain([dates[0].toDate(), dates[2].toDate()]).range([0, 100]);

test.after(async () => {
  await server?.close();
});

test("selectionToInclusiveDateRange maps a direct drag to inclusive dates", async () => {
  const range = (await selectionToInclusiveDateRange())([0, 50], scale, dates);

  assert.deepEqual([range?.from.format("YYYY-MM-DD"), range?.to.format("YYYY-MM-DD")], [
    "2024-01-01",
    "2024-01-10"
  ]);
});

test("selectionToInclusiveDateRange normalizes a reverse drag", async () => {
  const range = (await selectionToInclusiveDateRange())([50, 0], scale, dates);

  assert.deepEqual([range?.from.format("YYYY-MM-DD"), range?.to.format("YYYY-MM-DD")], [
    "2024-01-01",
    "2024-01-10"
  ]);
});

test("selectionToInclusiveDateRange returns null for empty dates", async () => {
  assert.equal((await selectionToInclusiveDateRange())([0, 50], scale, []), null);
});
