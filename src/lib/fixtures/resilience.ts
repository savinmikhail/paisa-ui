import rawFixture from "./resilience.json";
import { normalizeResilienceReport, type ResilienceReport } from "$lib/resilience";

export const RESILIENCE_FIXTURE: ResilienceReport = normalizeResilienceReport(rawFixture);

export function loadResilienceFixture(): ResilienceReport {
  return normalizeResilienceReport(rawFixture);
}
