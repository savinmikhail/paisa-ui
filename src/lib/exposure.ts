export const FAILURE_DOMAIN_FIELDS = [
  "institution",
  "broker",
  "custodian",
  "depositary",
  "jurisdiction",
  "correspondent_bank",
  "withdrawal_route",
  "key_control"
] as const;

export type FailureDomainField = (typeof FAILURE_DOMAIN_FIELDS)[number];
export type Confidence = "high" | "medium" | "low" | "unknown";

export interface CuratedMetadata {
  value: string | null;
  source: string | null;
  verified_at: string | null;
  confidence: Confidence;
}

export type FailureDomainMetadata = Partial<Record<FailureDomainField, CuratedMetadata | null>>;

export interface ExposureAsset {
  id: string;
  label: string;
  kind: string;
  market_value: number;
  currency: string;
  failure_domains: FailureDomainMetadata;
}

export interface ExposureReport {
  contract_version: "exposure.v1";
  as_of: string;
  base_currency: string;
  source: "synthetic-fixture" | "private-adapter";
  assets: ExposureAsset[];
}

export interface ExposureBucket {
  key: string;
  field: FailureDomainField;
  value: string | null;
  label: string;
  unknown: boolean;
  asset_ids: string[];
  asset_count: number;
  market_value: number;
  share: number;
  verified_count: number;
  low_confidence_count: number;
}

export interface ExposureOverlap {
  left_field: FailureDomainField;
  left_value: string;
  right_field: FailureDomainField;
  right_value: string;
  asset_ids: string[];
  asset_count: number;
  market_value: number;
}

export interface ExposureSummary {
  total_market_value: number;
  bucket_count: number;
  filled_cells: number;
  total_cells: number;
  coverage: number;
  buckets: ExposureBucket[];
  overlaps: ExposureOverlap[];
}

export const fieldLabels: Record<FailureDomainField, string> = {
  institution: "Institution",
  broker: "Broker",
  custodian: "Custodian",
  depositary: "Depositary",
  jurisdiction: "Jurisdiction",
  correspondent_bank: "Correspondent bank",
  withdrawal_route: "Withdrawal route",
  key_control: "Key control"
};

const UNKNOWN_LABEL = "Unknown";

export function metadataFor(
  asset: ExposureAsset,
  field: FailureDomainField
): CuratedMetadata | null {
  return asset.failure_domains[field] ?? null;
}

export function bucketKey(field: FailureDomainField, value: string | null): string {
  return `${field}:${value?.trim() || "unknown"}`;
}

function knownValue(metadata: CuratedMetadata | null): string | null {
  const value = metadata?.value?.trim();
  return value || null;
}

export function aggregateExposure(report: ExposureReport): ExposureSummary {
  const totalMarketValue = report.assets.reduce((total, asset) => total + asset.market_value, 0);
  const buckets = new Map<string, ExposureBucket>();
  let filledCells = 0;

  for (const asset of report.assets) {
    for (const field of FAILURE_DOMAIN_FIELDS) {
      const metadata = metadataFor(asset, field);
      const value = knownValue(metadata);
      if (value) filledCells += 1;

      const key = bucketKey(field, value);
      const bucket = buckets.get(key) ?? {
        key,
        field,
        value,
        label: value ?? UNKNOWN_LABEL,
        unknown: !value,
        asset_ids: [],
        asset_count: 0,
        market_value: 0,
        share: 0,
        verified_count: 0,
        low_confidence_count: 0
      };

      if (!bucket.asset_ids.includes(asset.id)) {
        bucket.asset_ids.push(asset.id);
        bucket.asset_count += 1;
        bucket.market_value += asset.market_value;
      }
      if (metadata?.verified_at) bucket.verified_count += 1;
      if (metadata?.confidence === "low" || metadata?.confidence === "unknown") {
        bucket.low_confidence_count += 1;
      }
      buckets.set(key, bucket);
    }
  }

  const bucketList = [...buckets.values()]
    .map((bucket) => ({
      ...bucket,
      share: totalMarketValue === 0 ? 0 : bucket.market_value / totalMarketValue
    }))
    .sort((left, right) => {
      if (left.field !== right.field) {
        return (
          FAILURE_DOMAIN_FIELDS.indexOf(left.field) - FAILURE_DOMAIN_FIELDS.indexOf(right.field)
        );
      }
      if (left.unknown !== right.unknown) return left.unknown ? 1 : -1;
      return right.market_value - left.market_value || left.label.localeCompare(right.label);
    });

  const overlaps: ExposureOverlap[] = [];
  for (let leftIndex = 0; leftIndex < FAILURE_DOMAIN_FIELDS.length; leftIndex += 1) {
    const leftField = FAILURE_DOMAIN_FIELDS[leftIndex];
    for (
      let rightIndex = leftIndex + 1;
      rightIndex < FAILURE_DOMAIN_FIELDS.length;
      rightIndex += 1
    ) {
      const rightField = FAILURE_DOMAIN_FIELDS[rightIndex];
      const pairings = new Map<string, ExposureOverlap>();

      for (const asset of report.assets) {
        const leftValue = knownValue(metadataFor(asset, leftField));
        const rightValue = knownValue(metadataFor(asset, rightField));
        if (!leftValue || !rightValue) continue;

        const key = `${leftValue}\u0000${rightValue}`;
        const overlap = pairings.get(key) ?? {
          left_field: leftField,
          left_value: leftValue,
          right_field: rightField,
          right_value: rightValue,
          asset_ids: [],
          asset_count: 0,
          market_value: 0
        };
        overlap.asset_ids.push(asset.id);
        overlap.asset_count += 1;
        overlap.market_value += asset.market_value;
        pairings.set(key, overlap);
      }

      overlaps.push(...pairings.values());
    }
  }

  overlaps.sort(
    (left, right) =>
      right.market_value - left.market_value ||
      `${left.left_value}:${left.right_value}`.localeCompare(
        `${right.left_value}:${right.right_value}`
      )
  );

  return {
    total_market_value: totalMarketValue,
    bucket_count: bucketList.length,
    filled_cells: filledCells,
    total_cells: report.assets.length * FAILURE_DOMAIN_FIELDS.length,
    coverage:
      report.assets.length === 0
        ? 0
        : filledCells / (report.assets.length * FAILURE_DOMAIN_FIELDS.length),
    buckets: bucketList,
    overlaps
  };
}

export function affectedAssets(
  report: ExposureReport,
  field: FailureDomainField,
  value: string | null
): ExposureAsset[] {
  return report.assets.filter(
    (asset) => knownValue(metadataFor(asset, field)) === (value?.trim() || null)
  );
}

export function fieldValueLabel(field: FailureDomainField, value: string | null): string {
  return `${fieldLabels[field]} · ${value || UNKNOWN_LABEL}`;
}
