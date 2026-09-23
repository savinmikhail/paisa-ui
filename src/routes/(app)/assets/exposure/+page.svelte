<script lang="ts">
  import {
    FAILURE_DOMAIN_FIELDS,
    affectedAssets,
    aggregateExposure,
    fieldLabels,
    fieldValueLabel,
    metadataFor,
    type ExposureBucket,
    type ExposureReport,
    type ExposureSummary,
    type FailureDomainField
  } from "$lib/exposure";
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

  onMount(async () => {
    try {
      report = await loadExposureReport(configuredSource, (route) => ajax(route));
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
        <div>{report?.contract_version || "exposure.v1"} · read-only</div>
      </div>
    </div>

    <div
      class="notification is-info is-light invertable exposure-notice"
      data-testid="source-notice"
    >
      <span class="icon mr-2"><i class="fas fa-flask" aria-hidden="true" /></span>
      {#if configuredSource === "synthetic"}
        <span
          ><strong>Synthetic fixture.</strong> No private Ledger metadata is loaded in this capture.</span
        >
      {:else}
        <span
          ><strong>Private metadata adapter.</strong> Read-only metadata is loaded through the configured
          Paisa API source.</span
        >
      {/if}
    </div>

    {#if loadError}
      <div class="message invertable is-danger" data-testid="exposure-error">
        <div class="message-header">Exposure is unavailable</div>
        <div class="message-body">{loadError}</div>
      </div>
    {:else if !report || !summary}
      <div class="box has-text-centered has-text-grey" data-testid="exposure-loading">
        Loading read-only exposure metadata…
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
                    ></tr
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
                  </article>
                {/each}
              </div>
            {:else}
              <p class="has-text-grey">Select a domain value to see affected assets.</p>
            {/if}
          </div>

          <div class="box exposure-panel contract-panel">
            <h2 class="title is-5 mb-2">Reporting contract</h2>
            <p class="is-size-7 has-text-grey mb-3">
              Manually curated metadata is read-only and carries provenance on every field. Missing
              values stay unknown; they are never inferred from the Ledger.
            </p>
            <div class="contract-list">
              <div><span>Source</span><strong>{report.source}</strong></div>
              <div><span>Verification cutoff</span><strong>{report.as_of}</strong></div>
              <div><span>Currency</span><strong>{report.base_currency}</strong></div>
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

  .exposure-notice {
    display: flex;
    align-items: center;
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
