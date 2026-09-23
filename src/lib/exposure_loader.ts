import type { ExposureReport } from "$lib/exposure";

export type ExposureSource = "private" | "synthetic";
export type ExposureRequest = (
  route: "/api/exposure" | "/api/exposure?source=synthetic"
) => Promise<unknown>;

export function exposureRoute(
  source: ExposureSource
): "/api/exposure" | "/api/exposure?source=synthetic" {
  return source === "synthetic" ? "/api/exposure?source=synthetic" : "/api/exposure";
}

function isExposureReport(value: unknown): value is ExposureReport {
  if (!value || typeof value !== "object") return false;
  const report = value as Partial<ExposureReport>;
  return (
    report.contract_version === "exposure.v1" &&
    (report.source === "private-adapter" || report.source === "synthetic-fixture") &&
    typeof report.as_of === "string" &&
    typeof report.base_currency === "string" &&
    Array.isArray(report.assets)
  );
}

export async function loadExposureReport(
  source: ExposureSource = "private",
  request: ExposureRequest
): Promise<ExposureReport> {
  const report = await request(exposureRoute(source));
  if (!isExposureReport(report)) {
    throw new Error("Exposure API returned an invalid exposure.v1 report");
  }
  return report;
}
