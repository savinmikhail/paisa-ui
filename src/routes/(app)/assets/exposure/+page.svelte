<script lang="ts">
  import {
    FAILURE_DOMAIN_FIELDS,
    affectedAssets,
    aggregateExposure,
    buildExposureRelationMap,
    bucketKey,
    fieldLabels,
    fieldValueLabel,
    metadataFor,
    type ExposureBucket,
    type ExposureRelation,
    type ExposureReport,
    type ExposureSummary,
    type FailureDomainField
  } from "$lib/exposure";
  import { syntheticExposureReport } from "$lib/exposure.fixture";
  import { loadExposureReport, type ExposureSource } from "$lib/exposure_loader";
  import { ajax, formatCurrency, formatPercentage } from "$lib/utils";
  import { onMount } from "svelte";

  const configuredSource: ExposureSource =
    import.meta.env.VITE_PAISA_EXPOSURE_SOURCE === "synthetic" ? "synthetic" : "private";
  let report: ExposureReport = null;
  let summary: ExposureSummary = null;
  let loadError = "";
  const fieldOptions = FAILURE_DOMAIN_FIELDS.map((field) => ({
    field,
    label: fieldLabels[field]
  }));

  let selectedField: FailureDomainField = "institution";
  let selectedBucketKey = "institution:Northstar Cooperative";

  $: visibleBuckets = summary?.buckets.filter((bucket) => bucket.field === selectedField) || [];
  $: if (!visibleBuckets.some((bucket) => bucket.key === selectedBucketKey)) {
    selectedBucketKey = visibleBuckets[0]?.key || "";
  }
  $: selectedBucket =
    summary?.buckets.find((bucket) => bucket.key === selectedBucketKey) ?? visibleBuckets[0];
  $: drilledAssets = selectedBucket
    ? affectedAssets(report, selectedBucket.field, selectedBucket.value)
    : [];
  $: unknownCount = summary?.buckets.filter((bucket) => bucket.unknown).length || 0;
  $: topOverlaps = summary?.overlaps.slice(0, 8) || [];
  $: relationMap = summary
    ? buildExposureRelationMap(summary, "jurisdiction", "institution", 3)
    : null;
  $: activeRelationKey = relationMap?.relations.find(
    (relation) =>
      (relation.left.field === selectedBucket?.field &&
        relation.left.value === selectedBucket?.value) ||
      (relation.right.field === selectedBucket?.field &&
        relation.right.value === selectedBucket?.value)
  )?.key;

  onMount(async () => {
    const fixtureMode = new URLSearchParams(window.location.search).get("fixture") === "synthetic";
    try {
      report = fixtureMode
        ? syntheticExposureReport
        : await loadExposureReport(configuredSource, (route) => ajax(route));
      summary = aggregateExposure(report);
    } catch (error) {
      loadError = error instanceof Error ? error.message : "Exposure API could not be loaded";
    }
  });

  function selectField(field: FailureDomainField) {
    selectedField = field;
  }

  function selectBucket(bucket: ExposureBucket) {
    selectedBucketKey = bucket.key;
  }

  function selectRelation(relation: ExposureRelation) {
    selectedField = relation.right.field;
    selectedBucketKey = bucketKey(relation.right.field, relation.right.value);
  }

  function relationAssetLabels(relation: ExposureRelation) {
    return relation.asset_ids
      .map((assetId) => report?.assets.find((asset) => asset.id === assetId)?.label)
      .filter(Boolean)
      .join(" · ");
  }

  function confidenceClass(confidence: string) {
    return confidence === "high"
      ? "is-success"
      : confidence === "medium"
        ? "is-info"
        : confidence === "low"
          ? "is-warning"
          : "is-light";
  }
</script>

<svelte:head>
  <title>Exposure · Paisa</title>
</svelte:head>

<section class="section exposure-screen" data-testid="exposure-screen">
  <div class="container is-fluid">
    <div class="exposure-heading">
      <div>
        <h1 class="title is-3 mb-2">Exposure</h1>
        <p class="subtitle is-6 mb-0">
          Where assets can fail together — across institutions, custody, routes, and controls.
        </p>
      </div>
      <div class="has-text-right is-size-7 has-text-grey">
        <div>As of {report?.as_of || "loading"}</div>
      </div>
    </div>

    {#if loadError}
      <div class="message invertable is-danger" data-testid="exposure-error">
        <div class="message-header">Exposure is unavailable</div>
        <div class="message-body">{loadError}</div>
      </div>
    {:else if !report || !summary}
      <div class="box has-text-centered has-text-grey" data-testid="exposure-loading">
        Loading exposure metadata…
      </div>
    {:else}
      <div class="columns is-variable is-5 mb-2">
        <div class="column is-7">
          <div class="exposure-summary box">
            <div class="summary-label">Reported exposure</div>
            <div class="summary-value">
              {formatCurrency(summary.total_market_value)}
              {report.base_currency}
            </div>
            <div class="summary-meta">
              {report.assets.length} assets · {summary.bucket_count} domain values
            </div>
          </div>
        </div>
        <div class="column is-5">
          <div class="exposure-summary box">
            <div class="summary-label">Metadata coverage</div>
            <div class="summary-value">{formatPercentage(summary.coverage, 0)}</div>
            <div class="summary-meta">
              {summary.filled_cells} of {summary.total_cells} curated fields · {unknownCount} unknown
              buckets
            </div>
          </div>
        </div>
      </div>

      <div class="box exposure-panel relation-map-panel" data-testid="exposure-relation-map">
        <div class="panel-heading-row relation-map-heading">
          <div>
            <h2 class="title is-5 mb-1">Shared exposure map</h2>
            <p class="has-text-grey is-size-7">
              The strongest Jurisdiction × Institution relationships, linked to the same assets.
              Select a relationship to open its institution drilldown.
            </p>
          </div>
          <div class="relation-map-legend" aria-label="Shared exposure map legend">
            <span><i class="legend-swatch legend-swatch-domain" aria-hidden="true" />Domain</span>
            <span
              ><i class="legend-swatch legend-swatch-overlap" aria-hidden="true" />Joint
              concentration</span
            >
          </div>
        </div>

        <p class="relation-map-note">
          Overlap means assets share both labels; it is a concentration signal, not a default
          probability or mathematical Venn area.
        </p>

        {#if relationMap?.relations.length}
          <div class="relation-map" role="group" aria-label="Top shared exposure relationships">
            {#each relationMap.relations as relation (relation.key)}
              <button
                class="relation-map-item"
                class:is-active={activeRelationKey === relation.key}
                type="button"
                aria-current={activeRelationKey === relation.key ? "true" : undefined}
                aria-controls="exposure-drilldown"
                aria-label={`${relation.left.label} in ${fieldLabels[relation.left.field]} and ${
                  relation.right.label
                } in ${fieldLabels[relation.right.field]}, ${formatCurrency(
                  relation.market_value
                )} ${report.base_currency}, ${formatPercentage(
                  relation.share,
                  0
                )} of reported exposure. Open institution drilldown.`}
                on:click={() => selectRelation(relation)}
              >
                <span class="relation-circles" aria-hidden="true">
                  <span class="relation-circle relation-circle-left">
                    <span class="relation-circle-field">{fieldLabels[relation.left.field]}</span>
                    <strong>{relation.left.label}</strong>
                    <span>{formatCurrency(relation.left.market_value)} {report.base_currency}</span>
                    <span>{formatPercentage(relation.left.share, 0)} of total</span>
                  </span>
                  <span class="relation-circle relation-circle-right">
                    <span class="relation-circle-field">{fieldLabels[relation.right.field]}</span>
                    <strong>{relation.right.label}</strong>
                    <span>{formatCurrency(relation.right.market_value)} {report.base_currency}</span
                    >
                    <span>{formatPercentage(relation.right.share, 0)} of total</span>
                  </span>
                  <span class="relation-overlap">
                    <span>Joint concentration</span>
                    <strong>{formatCurrency(relation.market_value)} {report.base_currency}</strong>
                    <span
                      >{formatPercentage(relation.share, 0)} · {relation.asset_count} assets</span
                    >
                  </span>
                </span>
                <span class="relation-map-footer">
                  <span>
                    <span class="relation-footer-label">Shared assets</span>
                    <span>{relationAssetLabels(relation)}</span>
                  </span>
                  <span class="relation-map-action"
                    >Open drilldown <span aria-hidden="true">→</span></span
                  >
                </span>
              </button>
            {/each}
          </div>
        {:else}
          <p class="has-text-grey">No known relationship is available for this map.</p>
        {/if}
      </div>

      <div class="columns is-variable is-5">
        <div class="column is-7">
          <div class="box exposure-panel">
            <div class="panel-heading-row">
              <div>
                <h2 class="title is-5 mb-1">Failure domains</h2>
                <p class="has-text-grey is-size-7">
                  Select a domain, then a value to inspect the affected assets.
                </p>
              </div>
              <span class="tag is-rounded is-light">{visibleBuckets.length} values</span>
            </div>

            <div
              class="field is-grouped is-grouped-multiline domain-tabs"
              role="tablist"
              aria-label="Failure domain types"
            >
              {#each fieldOptions as option}
                <p class="control">
                  <button
                    class="button is-small"
                    class:is-link={selectedField === option.field}
                    class:is-light={selectedField !== option.field}
                    aria-selected={selectedField === option.field}
                    role="tab"
                    on:click={() => selectField(option.field)}>{option.label}</button
                  >
                </p>
              {/each}
            </div>

            <div class="table-container">
              <table class="table is-fullwidth is-hoverable exposure-table">
                <thead>
                  <tr>
                    <th>{fieldLabels[selectedField]}</th>
                    <th class="has-text-right">Assets</th>
                    <th class="has-text-right">Exposure</th>
                    <th class="has-text-right">Share</th>
                    <th class="has-text-right">Verified</th>
                  </tr>
                </thead>
                <tbody>
                  {#each visibleBuckets as bucket}
                    <tr class:is-selected={selectedBucket?.key === bucket.key}>
                      <td>
                        <button class="bucket-button" on:click={() => selectBucket(bucket)}>
                          <span class:has-text-grey={bucket.unknown} class="has-text-weight-medium"
                            >{bucket.label}</span
                          >
                          {#if bucket.unknown}<span class="tag is-warning is-light ml-2"
                              >unknown</span
                            >{/if}
                        </button>
                      </td>
                      <td class="has-text-right">{bucket.asset_count}</td>
                      <td class="has-text-right"
                        >{formatCurrency(bucket.market_value)} {report.base_currency}</td
                      >
                      <td class="has-text-right">{formatPercentage(bucket.share, 0)}</td>
                      <td class="has-text-right">{bucket.verified_count}/{bucket.asset_count}</td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          </div>

          <div class="box exposure-panel">
            <div class="panel-heading-row">
              <div>
                <h2 class="title is-5 mb-1">Cross-domain overlaps</h2>
                <p class="has-text-grey is-size-7">
                  Known values shared by the same affected assets.
                </p>
              </div>
              <span class="tag is-rounded is-light">{summary.overlaps.length} pairings</span>
            </div>

            <div class="table-container">
              <table class="table is-fullwidth exposure-table">
                <thead>
                  <tr
                    ><th>Shared path</th><th class="has-text-right">Assets</th><th
                      class="has-text-right">Exposure</th
                    ><th class="has-text-right">Share</th></tr
                  >
                </thead>
                <tbody>
                  {#each topOverlaps as overlap}
                    <tr>
                      <td>
                        <div class="is-size-7 has-text-grey">
                          {fieldLabels[overlap.left_field]} × {fieldLabels[overlap.right_field]}
                        </div>
                        <div class="has-text-weight-medium">
                          {overlap.left_value} <span class="has-text-grey">→</span>
                          {overlap.right_value}
                        </div>
                      </td>
                      <td class="has-text-right">{overlap.asset_count}</td>
                      <td class="has-text-right"
                        >{formatCurrency(overlap.market_value)} {report.base_currency}</td
                      >
                      <td class="has-text-right">
                        {formatPercentage(overlap.market_value / summary.total_market_value, 0)}
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <div class="column is-5">
          <div class="box exposure-panel drilldown-panel" data-testid="exposure-drilldown">
            <div class="panel-heading-row">
              <div>
                <div class="has-text-grey is-size-7">Affected-asset drilldown</div>
                <h2 class="title is-5 mb-1">
                  {selectedBucket
                    ? fieldValueLabel(selectedBucket.field, selectedBucket.value)
                    : "No selection"}
                </h2>
              </div>
              {#if selectedBucket}<span class="tag is-link is-light"
                  >{selectedBucket.asset_count} assets</span
                >{/if}
            </div>

            {#if selectedBucket}
              <div class="drilldown-total">
                <span>{formatCurrency(selectedBucket.market_value)} {report.base_currency}</span>
                <span>{formatPercentage(selectedBucket.share, 0)} of reported exposure</span>
              </div>
              <div class="asset-list">
                {#each drilledAssets as asset}
                  <article class="asset-row">
                    <div class="asset-row-heading">
                      <div>
                        <strong>{asset.label}</strong>
                        <div class="is-size-7 has-text-grey">{asset.kind} · {asset.currency}</div>
                      </div>
                      <span class="has-text-weight-medium"
                        >{formatCurrency(asset.market_value)} {report.base_currency}</span
                      >
                    </div>
                    <dl class="metadata-grid">
                      {#each FAILURE_DOMAIN_FIELDS as field}
                        {@const metadata = metadataFor(asset, field)}
                        <div>
                          <dt>{fieldLabels[field]}</dt>
                          <dd>
                            {metadata?.value || "Unknown"}
                            {#if metadata}
                              <span
                                class="tag is-rounded is-small {confidenceClass(
                                  metadata.confidence
                                )} is-light ml-1">{metadata.confidence}</span
                              >
                            {/if}
                          </dd>
                        </div>
                      {/each}
                    </dl>
                    {#if report.source !== "synthetic-fixture"}
                      <div class="provenance-row">
                        <span
                          >{metadataFor(asset, selectedBucket.field)?.source ||
                            "No source recorded"}</span
                        >
                        <span
                          >verified {metadataFor(asset, selectedBucket.field)?.verified_at ||
                            "never"}</span
                        >
                      </div>
                    {/if}
                  </article>
                {/each}
              </div>
            {:else}
              <p class="has-text-grey">Select a domain value to see affected assets.</p>
            {/if}
          </div>

          <div class="box exposure-panel contract-panel">
            <h2 class="title is-5 mb-2">Metadata notes</h2>
            <p class="is-size-7 has-text-grey mb-3">
              Curated metadata carries provenance on every field. Missing values stay unknown; they
              are never inferred from the Ledger.
            </p>
            <div class="contract-list">
              <div><span>Verification cutoff</span><strong>{report.as_of}</strong></div>
              <div><span>Base currency</span><strong>{report.base_currency}</strong></div>
              <div>
                <span>Curated fields</span><strong
                  >{summary.filled_cells}/{summary.total_cells}</strong
                >
              </div>
            </div>
          </div>
        </div>
      </div>
    {/if}
  </div>
</section>

<style lang="scss">
  @import "bulma/sass/utilities/_all.sass";

  .exposure-screen {
    padding-top: 1.5rem;
  }

  .exposure-heading,
  .panel-heading-row,
  .asset-row-heading,
  .drilldown-total,
  .contract-list div {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
  }

  .exposure-heading {
    margin-bottom: 1.25rem;
  }

  .exposure-summary {
    height: 100%;
    padding: 1.15rem 1.25rem;
  }

  .summary-label {
    color: $grey;
    font-size: 0.75rem;
    font-weight: 700;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }

  .summary-value {
    font-size: 1.8rem;
    font-weight: 650;
    line-height: 1.2;
    margin-top: 0.35rem;
  }

  .summary-meta {
    color: $grey-dark;
    font-size: 0.8rem;
    margin-top: 0.35rem;
  }

  .exposure-panel {
    height: calc(100% - 1.5rem);
    padding: 1.25rem;
  }

  .relation-map-panel {
    height: auto;
    margin-bottom: 1rem;
  }

  .relation-map-heading {
    align-items: flex-start;
    margin-bottom: 0.65rem;
  }

  .relation-map-legend {
    align-items: center;
    color: $grey-dark;
    display: flex;
    flex-wrap: wrap;
    font-size: 0.72rem;
    gap: 0.75rem;
    justify-content: flex-end;
  }

  .relation-map-legend span {
    align-items: center;
    display: inline-flex;
    gap: 0.3rem;
    white-space: nowrap;
  }

  .legend-swatch {
    border-radius: 50%;
    display: inline-block;
    height: 0.65rem;
    width: 0.65rem;
  }

  .legend-swatch-domain {
    background: rgba($link, 0.16);
    border: 1px solid rgba($link, 0.65);
  }

  .legend-swatch-overlap {
    background: $link;
  }

  .relation-map-note {
    background: $white-ter;
    border-radius: $radius-small;
    color: $grey-dark;
    font-size: 0.78rem;
    line-height: 1.45;
    margin: 0 0 1rem;
    padding: 0.65rem 0.75rem;
  }

  .relation-map {
    display: grid;
    gap: 0.85rem;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .relation-map-item {
    background: transparent;
    border: 1px solid $grey-lighter;
    border-radius: $radius;
    color: inherit;
    cursor: pointer;
    display: block;
    padding: 1rem;
    text-align: left;
    transition:
      background-color 150ms ease-out,
      border-color 150ms ease-out;
    width: 100%;
  }

  .relation-map-item:hover {
    background: $white-ter;
    border-color: $link;
  }

  .relation-map-item.is-active {
    background: $link-light;
    border-color: $link;
  }

  .relation-map-item:focus-visible {
    outline: 2px solid $link;
    outline-offset: 3px;
  }

  .relation-circles {
    align-items: center;
    display: flex;
    min-height: 10.25rem;
    position: relative;
  }

  .relation-circle {
    align-items: center;
    background: rgba($link, 0.08);
    border: 1px solid rgba($link, 0.45);
    border-radius: 50%;
    display: flex;
    flex: 0 0 9.75rem;
    flex-direction: column;
    gap: 0.16rem;
    height: 9.75rem;
    justify-content: center;
    padding: 1rem 1.25rem;
    position: relative;
    text-align: center;
    z-index: 1;
  }

  .relation-circle-right {
    background: rgba($info, 0.08);
    border-color: rgba($info, 0.55);
    margin-left: -1.25rem;
  }

  .relation-circle-field,
  .relation-footer-label {
    color: $grey;
    font-size: 0.66rem;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .relation-circle strong {
    font-size: 0.8rem;
    line-height: 1.15;
    max-width: 8rem;
  }

  .relation-circle > span:not(.relation-circle-field) {
    color: $grey-dark;
    font-size: 0.7rem;
    line-height: 1.15;
  }

  .relation-overlap {
    align-items: center;
    background: $white;
    border: 1px solid $link;
    border-radius: $radius;
    box-shadow: 0 4px 10px rgba($black, 0.14);
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
    left: 50%;
    padding: 0.55rem 0.45rem;
    position: absolute;
    text-align: center;
    top: 50%;
    transform: translate(-50%, -50%);
    width: 6.8rem;
    z-index: 2;
  }

  .relation-overlap span:first-child {
    color: $link;
    font-size: 0.64rem;
    font-weight: 700;
    letter-spacing: 0.03em;
    text-transform: uppercase;
  }

  .relation-overlap strong {
    font-size: 0.78rem;
    line-height: 1.15;
  }

  .relation-overlap span:last-child {
    color: $grey-dark;
    font-size: 0.68rem;
  }

  .relation-map-footer {
    align-items: flex-end;
    border-top: 1px solid $grey-lighter;
    display: flex;
    font-size: 0.74rem;
    gap: 0.75rem;
    justify-content: space-between;
    margin-top: 0.75rem;
    padding-top: 0.7rem;
  }

  .relation-map-footer > span:first-child {
    display: grid;
    gap: 0.12rem;
    min-width: 0;
  }

  .relation-map-footer > span:first-child > span:last-child {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .relation-map-action {
    color: $link;
    flex: 0 0 auto;
    font-weight: 650;
    white-space: nowrap;
  }

  .panel-heading-row {
    margin-bottom: 1rem;
  }

  .domain-tabs {
    border-bottom: 1px solid $grey-lighter;
    margin-bottom: 0.75rem;
    padding-bottom: 0.65rem;
  }

  .exposure-table {
    margin-bottom: 0;

    th {
      color: $grey;
      font-size: 0.72rem;
      letter-spacing: 0.04em;
      text-transform: uppercase;
    }

    td,
    th {
      vertical-align: middle;
    }

    tr.is-selected td {
      background: $link-light;
    }
  }

  .bucket-button {
    background: transparent;
    border: 0;
    color: inherit;
    cursor: pointer;
    padding: 0;
    text-align: left;
  }

  .drilldown-panel {
    min-height: 26rem;
  }

  .drilldown-total {
    border-bottom: 1px solid $grey-lighter;
    color: $grey-dark;
    font-size: 0.8rem;
    margin-bottom: 0.5rem;
    padding-bottom: 0.75rem;
  }

  .drilldown-total span:first-child {
    color: $text;
    font-size: 1.15rem;
    font-weight: 650;
  }

  .asset-row {
    border-bottom: 1px solid $grey-lighter;
    padding: 0.9rem 0;
  }

  .asset-row:last-child {
    border-bottom: 0;
  }

  .metadata-grid {
    display: grid;
    gap: 0.65rem 0.8rem;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    margin-top: 0.85rem;
  }

  .metadata-grid dt {
    color: $grey;
    font-size: 0.68rem;
    letter-spacing: 0.03em;
    text-transform: uppercase;
  }

  .metadata-grid dd {
    font-size: 0.8rem;
    margin: 0.1rem 0 0;
  }

  .provenance-row {
    color: $grey;
    display: flex;
    font-family: $family-monospace;
    font-size: 0.68rem;
    gap: 0.8rem;
    justify-content: space-between;
    margin-top: 0.85rem;
  }

  .contract-panel {
    height: auto;
    min-height: 0;
  }

  .contract-list {
    display: grid;
    gap: 0.6rem;
    font-size: 0.8rem;
  }

  .contract-list span {
    color: $grey;
  }

  .contract-list strong {
    font-family: $family-monospace;
    font-size: 0.75rem;
    font-weight: 500;
  }

  :global(html[data-theme="dark"]) .relation-map-note {
    background: hsl(215, 18%, 20%);
    color: hsl(0, 0%, 74%);
  }

  :global(html[data-theme="dark"]) .relation-map-item.is-active {
    background: rgba(72, 116, 214, 0.2);
  }

  :global(html[data-theme="dark"]) .relation-overlap {
    background: hsl(215, 18%, 10%);
    color: hsl(0, 0%, 94%);
  }

  :global(html[data-theme="dark"]) .relation-circle > span:not(.relation-circle-field),
  :global(html[data-theme="dark"]) .relation-overlap span:last-child {
    color: hsl(0, 0%, 72%);
  }

  @include mobile {
    .exposure-heading,
    .panel-heading-row,
    .asset-row-heading,
    .drilldown-total {
      align-items: flex-start;
      flex-direction: column;
    }

    .exposure-heading > div:last-child {
      text-align: left;
    }

    .relation-map-heading {
      gap: 0.65rem;
    }

    .relation-map-legend {
      justify-content: flex-start;
    }

    .relation-map {
      grid-template-columns: 1fr;
    }

    .relation-circles {
      min-height: 9.25rem;
    }

    .relation-circle {
      flex-basis: 8.75rem;
      height: 8.75rem;
      padding-left: 0.75rem;
      padding-right: 0.75rem;
    }

    .relation-circle-right {
      margin-left: -1rem;
    }

    .relation-overlap {
      width: 6.2rem;
    }

    .relation-map-footer {
      align-items: flex-start;
      flex-direction: column;
      gap: 0.4rem;
    }

    .metadata-grid {
      grid-template-columns: 1fr;
    }

    .provenance-row {
      align-items: flex-start;
      flex-direction: column;
      gap: 0.2rem;
    }
  }
</style>
