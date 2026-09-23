export const trailingSlash = "never";

import type { LayoutLoad } from "./$types";
import { ajax, configUpdated, setNow } from "$lib/utils";

export const load = (async ({ url }) => {
  if (url.searchParams.get("fixture") === "networth-ticket-04") {
    globalThis.USER_CONFIG = {
      default_currency: "USD",
      readonly: true,
      locale: "en-US",
      journal_path: "fixture",
      display_precision: 0,
      db_path: "fixture",
      financial_year_starting_month: 1,
      amount_alignment_column: 52,
      week_starting_day: 0,
      goals: {},
      accounts: []
    } as UserConfig;
    configUpdated();
    return {};
  }

  const { config, now } = await ajax("/api/config");
  if (now) {
    setNow(now);
  }
  globalThis.USER_CONFIG = config;
  configUpdated();
  return {};
}) satisfies LayoutLoad;
