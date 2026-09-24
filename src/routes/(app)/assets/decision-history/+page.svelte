<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import {
    privateDecisionHistoryAdapter,
    syntheticDecisionHistory,
    type DecisionHistoryCase
  } from "$lib/decision_history";
  let loadState: "loading" | "ready" | "unavailable" = "loading";
  let decisionCases: DecisionHistoryCase[] = [];
  let isFixture = false;
  const rub = (n: number) =>
    new Intl.NumberFormat("ru-RU", {
      style: "currency",
      currency: "RUB",
      maximumFractionDigits: 0
    }).format(n);
  const signed = (n: number) => `${n > 0 ? "+" : ""}${rub(n)}`;
  const number = (n: number) =>
    new Intl.NumberFormat("ru-RU", { maximumFractionDigits: 2 }).format(n);
  const tone = (n: number) => (n > 0 ? "gain" : n < 0 ? "loss" : "");
  const when = (s: string) => new Intl.DateTimeFormat("ru-RU").format(new Date(`${s}T12:00:00`));
  onMount(async () => {
    isFixture = $page.url.searchParams.get("fixture") === "synthetic";
    if (isFixture) {
      decisionCases = syntheticDecisionHistory.cases;
      loadState = "ready";
      return;
    }
    try {
      const response = await privateDecisionHistoryAdapter.load(fetch);
      if (response.kind === "ready") {
        decisionCases = response.payload.cases;
        loadState = "ready";
      } else loadState = "unavailable";
    } catch {
      loadState = "unavailable";
    }
  });
  function retryLoading() {
    window.location.reload();
  }
</script>

<svelte:head><title>История решений · Paisa</title></svelte:head>
<section class="section decision-history-page">
  <div class="container is-fluid">
    <h1 class="title is-3">История решений</h1>
    <p class="intro">Сколько принесли решения и обогнали ли они накопительный счёт.</p>
    {#if loadState === "loading"}
      <p role="status" data-testid="decision-history-loading">Загрузка истории решений…</p>
    {:else if loadState === "unavailable"}
      <div data-testid="decision-history-unavailable">
        <h2 class="title is-4">История решений недоступна</h2>
        <p>Проверьте подключение и повторите попытку.</p>
        <button class="button is-link mt-4" type="button" on:click={retryLoading}
          >Повторить загрузку</button
        >
      </div>
    {:else}
      <div data-testid="decision-history-ready">
        {#if isFixture}<p class="fixture-note">Иллюстративные данные для проверки экрана.</p>{/if}
        {#if decisionCases.some((c) => c.result)}
          <div class="summary-wrap">
            <table class="results-summary">
              <caption
                >Результаты в рублях. Прибыль решения и преимущество перед альтернативой — разные
                показатели.</caption
              >
              <thead
                ><tr
                  ><th scope="col">Решение</th><th scope="col">Принесло</th><th scope="col"
                    >Против накопительного</th
                  ></tr
                ></thead
              >
              <tbody>
                {#each decisionCases as item}
                  {#if item.result}
                    <tr>
                      <th scope="row"
                        ><a href={`#${item.id}`}>{item.title}</a><small
                          >{item.result.coverage}</small
                        ></th
                      >
                      <td class={tone(item.result.netResult)}>{signed(item.result.netResult)}</td>
                      <td class={tone(item.result.difference)}
                        ><strong class={tone(item.result.difference)}
                          >{signed(item.result.difference)}</strong
                        ><small>при {item.result.benchmarkRate}% годовых</small></td
                      >
                    </tr>
                  {/if}
                {/each}
              </tbody>
            </table>
          </div>
          <p class="scope-note">
            Это денежная ретроспектива на указанные даты. Польза доступа к деньгам, ограничения
            вывода и будущая доходность оцениваются отдельно. Пересекающиеся решения не
            складываются.
          </p>
        {/if}
        {#each decisionCases as item}
          <article id={item.id} class="decision-case" data-testid={`decision-case-${item.id}`}>
            <header>
              <h2>{item.title}</h2>
              <p>{item.period}</p>
            </header>
            {#if item.result}
              {@const r = item.result}
              <div class="verdict">
                <h3 class={tone(r.difference)}>
                  {r.difference >= 0 ? "Выигрыш" : "Отставание"}
                  {rub(Math.abs(r.difference))} относительно счёта под {r.benchmarkRate}%
                </h3>
                <p>
                  Само решение принесло <strong class={tone(r.netResult)}
                    >{signed(r.netResult)}</strong
                  >
                  после учтённых издержек. Оценка на {when(r.asOf)}.
                </p>
                <p class="coverage">{r.coverage}</p>
                {#if r.lastPostingDate}<p class="detail-note">
                    Последняя операция в данных: {when(r.lastPostingDate)}. Оценка остатка по ценам
                    на {when(r.asOf)}.
                  </p>{/if}
              </div>
              <div class="result-layout">
                <div>
                  <h3>Из чего сложился результат</h3>
                  <dl class="reconciliation">
                    <div>
                      <dt>Вложено / исходная стоимость</dt>
                      <dd>{rub(r.invested)}</dd>
                    </div>
                    <div>
                      <dt>Осталось по рыночной оценке</dt>
                      <dd>{rub(r.currentValue)}</dd>
                    </div>
                    <div>
                      <dt>Выплачено / использовано</dt>
                      <dd>{rub(r.withdrawn)}</dd>
                    </div>
                    <div class="total">
                      <dt>Остаток + выплаты − вложения</dt>
                      <dd class={tone(r.netResult)}>{signed(r.netResult)}</dd>
                    </div>
                  </dl>
                  <p class="detail-note">
                    Комиссии: {rub(r.fees)}. Удержанный налог: {rub(r.taxes)}.{#if r.taxRefund}
                      Налоговый возврат: {rub(r.taxRefund)}.{/if} Эти суммы уже учтены в итоге.
                  </p>
                  {#if r.fxEffect !== undefined && r.conversionCost !== undefined}
                    <dl class="reconciliation attribution">
                      <div>
                        <dt>Изменение курса за время владения</dt>
                        <dd class={tone(r.fxEffect)}>{signed(r.fxEffect)}</dd>
                      </div>
                      <div>
                        <dt>Издержки конвертации в фактическом курсе</dt>
                        <dd>{rub(r.conversionCost)}</dd>
                      </div>
                    </dl>
                    <p class="detail-note">
                      Издержки — разница между фактической RUB-стоимостью и стоимостью валюты по
                      учётному курсу в дату получения. Они уже входят во вложения, повторно не
                      вычитаются.
                    </p>
                  {/if}
                </div>
                <div class="comparison">
                  <h3>Если бы деньги лежали на счёте</h3>
                  <div class="comparison-row">
                    <span>Решение, к дате оценки</span><strong
                      >{rub(r.actualComparableValue)}</strong
                    >
                  </div>
                  <div class="bar-track">
                    <div
                      class="bar actual"
                      style={`width:${Math.max(
                        0,
                        (r.actualComparableValue /
                          Math.max(r.actualComparableValue, r.benchmarkValue, 1)) *
                          100
                      )}%`}
                    ></div>
                  </div>
                  <div class="comparison-row">
                    <span>Накопительный счёт · {r.benchmarkRate}%</span><strong
                      >{rub(r.benchmarkValue)}</strong
                    >
                  </div>
                  <div class="bar-track">
                    <div
                      class="bar alternative"
                      style={`width:${Math.max(
                        0,
                        (r.benchmarkValue /
                          Math.max(r.actualComparableValue, r.benchmarkValue, 1)) *
                          100
                      )}%`}
                    ></div>
                  </div>
                  <p class="comparison-difference">
                    Разница <strong class={tone(r.difference)}>{signed(r.difference)}</strong>
                  </p>
                  <p class="detail-note">
                    Выплаты доведены до даты оценки по той же ставке: {rub(r.payoutCarry)} расчётных
                    процентов. Поэтому сравнимая стоимость отличается от остатка плюс выплаты.
                  </p>
                  <p class="detail-note">
                    Ставка постоянна во всём сценарии. Альтернатива до индивидуального налога на
                    проценты; результат решения после фактических удержаний.
                  </p>
                </div>
              </div>
              {#if r.segments?.length}
                <div class="table-wrap">
                  <table class="flow-table">
                    <caption>Внутри решения</caption><thead
                      ><tr
                        ><th>Направление</th><th>Вложено</th><th>Принесло</th><th>К альтернативе</th
                        ></tr
                      ></thead
                    ><tbody
                      >{#each r.segments as segment}<tr
                          ><th>{segment.label}</th><td>{rub(segment.invested)}</td><td
                            class={tone(segment.netResult)}>{signed(segment.netResult)}</td
                          ><td class={tone(segment.difference)}>{signed(segment.difference)}</td
                          ></tr
                        >{/each}</tbody
                    >
                  </table>
                </div>
              {/if}
              <details class="calculation-details">
                <summary>Операции и расчёт ({r.details.length})</summary>
                <p class="method">{r.method}</p>
                <div class="table-wrap">
                  <table class="flow-table">
                    <thead
                      ><tr
                        ><th scope="col">Дата</th><th scope="col">Операция</th><th scope="col"
                          >Сумма, RUB</th
                        ><th scope="col">Получено / курс</th></tr
                      ></thead
                    ><tbody>
                      {#each r.details as row}<tr
                          ><td>{when(row.date)}</td><td>{row.label}<small>{row.source}</small></td
                          ><td>{rub(row.amount)}</td><td
                            >{#if row.quantity !== undefined}{number(row.quantity)}
                              {row.currency}<small>{number(row.rate ?? 0)} RUB/{row.currency}</small
                              >{:else}—{/if}</td
                          ></tr
                        >{/each}
                    </tbody>
                  </table>
                </div>
              </details>
            {:else}
              <p>{item.decision}</p>
              <p>{item.outcome}</p>
              <p><strong>{item.amount.label}</strong> · {item.amount.currency}</p>
            {/if}
            <details class="source-details">
              <summary>Границы расчёта и источники</summary>
              <p>{item.decision}</p>
              <ul>
                {#each item.caveats as caveat}<li>{caveat}</li>{/each}
              </ul>
              <ul class="sources">
                {#each item.sourceRefs as source}<li>{source}</li>{/each}
              </ul>
            </details>
          </article>
        {/each}
      </div>
    {/if}
  </div>
</section>

<style lang="scss">
  .decision-history-page {
    --muted: hsl(var(--bc) / 0.76);
    --line: hsl(var(--bc) / 0.18);
    color: hsl(var(--bc));
    font-variant-numeric: tabular-nums;
  }
  .intro {
    margin: -0.5rem 0 2rem;
    max-width: 70ch;
    color: var(--muted);
  }
  .summary-wrap,
  .table-wrap {
    overflow-x: auto;
  }
  table {
    width: 100%;
    border-collapse: collapse;
  }
  caption {
    text-align: left;
    color: var(--muted);
    padding: 0 0 1rem;
  }
  th,
  td {
    text-align: left;
    padding: 1rem;
    border-bottom: 1px solid var(--line);
    vertical-align: top;
  }
  th {
    font-weight: 600;
  }
  thead th {
    color: var(--muted);
    font-size: 0.85rem;
  }
  td:not(:first-child) {
    white-space: nowrap;
  }
  .results-summary td {
    font-size: 1.1rem;
  }
  .results-summary th:first-child {
    width: 52%;
  }
  small {
    display: block;
    font-weight: 400;
    color: var(--muted);
    font-size: 0.8rem;
    margin-top: 0.4rem;
    max-width: 65ch;
    white-space: normal;
  }
  a {
    color: hsl(var(--bc));
    text-decoration: underline;
    text-underline-offset: 0.2em;
  }
  a:hover {
    color: hsl(var(--p));
  }
  .scope-note {
    color: var(--muted);
    max-width: 75ch;
    margin: 1rem 0 3rem;
    font-size: 0.9rem;
  }
  .decision-case {
    scroll-margin-top: 5rem;
    padding: 2rem 0 2.5rem;
    border-top: 1px solid var(--line);
  }
  header h2 {
    font-size: 1.6rem;
    font-weight: 650;
    line-height: 1.3;
    margin: 0 0 0.5rem;
    text-wrap: balance;
  }
  header p {
    color: var(--muted);
    font-size: 0.9rem;
  }
  h3 {
    font-size: 1rem;
    font-weight: 650;
    line-height: 1.4;
    margin-bottom: 1rem;
    text-wrap: balance;
  }
  .verdict {
    margin: 1.75rem 0 2rem;
  }
  .verdict h3 {
    font-size: 1.35rem;
    margin-bottom: 0.6rem;
  }
  .verdict p {
    max-width: 75ch;
  }
  .coverage {
    color: var(--muted);
    margin-top: 0.5rem;
    font-size: 0.9rem;
  }
  .gain {
    color: #40c9a0;
  }
  .loss {
    color: #ff969e;
  }
  :global([data-theme="light"]) .gain {
    color: #166c50;
  }
  :global([data-theme="light"]) .loss {
    color: #aa263a;
  }
  .result-layout {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 3rem;
  }
  .reconciliation div {
    display: flex;
    justify-content: space-between;
    gap: 1.5rem;
    padding: 0.6rem 0;
  }
  dt {
    color: var(--muted);
  }
  dd {
    white-space: nowrap;
    font-weight: 550;
  }
  .total {
    border-top: 1px solid var(--line);
    margin-top: 0.5rem;
  }
  .total dt {
    color: hsl(var(--bc));
  }
  .detail-note {
    color: var(--muted);
    font-size: 0.85rem;
    line-height: 1.55;
    max-width: 70ch;
    margin-top: 0.8rem;
  }
  .attribution {
    margin-top: 1.5rem;
    font-size: 0.9rem;
  }
  .comparison-row {
    display: flex;
    gap: 1rem;
    justify-content: space-between;
    font-size: 0.9rem;
  }
  .comparison-row strong {
    white-space: nowrap;
  }
  .bar-track {
    height: 10px;
    background: hsl(var(--bc) / 0.08);
    margin: 0.6rem 0 1.25rem;
  }
  .bar {
    height: 100%;
  }
  .actual {
    background: #9083ec;
  }
  .alternative {
    background: hsl(var(--bc) / 0.45);
  }
  .comparison-difference {
    font-size: 1.1rem;
    margin: 1.5rem 0 0;
  }
  details {
    border-top: 1px solid var(--line);
    margin-top: 1.5rem;
    padding-top: 1rem;
  }
  summary {
    cursor: pointer;
    font-weight: 550;
    width: fit-content;
  }
  summary:hover {
    text-decoration: underline;
    text-underline-offset: 0.2em;
  }
  summary:focus-visible,
  a:focus-visible,
  button:focus-visible {
    outline: 2px solid hsl(var(--p));
    outline-offset: 4px;
  }
  details p,
  details ul {
    margin-top: 1rem;
    max-width: 75ch;
    color: var(--muted);
  }
  details ul {
    list-style: disc;
    padding-left: 1.3rem;
  }
  details li + li {
    margin-top: 0.6rem;
  }
  .sources {
    font-size: 0.8rem;
    overflow-wrap: anywhere;
  }
  .flow-table {
    font-size: 0.85rem;
    margin-top: 1rem;
  }
  .flow-table td:nth-child(2) {
    white-space: normal;
    min-width: 12rem;
  }
  .fixture-note {
    padding: 0.75rem;
    margin-bottom: 1.5rem;
    background: hsl(var(--bc) / 0.08);
  }
  ::selection {
    background: hsl(var(--p) / 0.4);
  }
  @media (max-width: 1000px) {
    .result-layout {
      grid-template-columns: 1fr;
      gap: 2rem;
    }
  }
  @media (max-width: 600px) {
    .section {
      padding: 1.5rem 1rem;
    }
    .container.is-fluid {
      padding: 0;
    }
    .results-summary th,
    .results-summary td {
      padding: 0.7rem 0.6rem;
    }
    .results-summary {
      min-width: 0;
    }
    .results-summary caption {
      display: block;
      max-width: 100%;
    }
    .results-summary thead {
      position: absolute;
      width: 1px;
      height: 1px;
      overflow: hidden;
      clip-path: inset(50%);
    }
    .results-summary tbody {
      display: block;
    }
    .results-summary tbody tr {
      display: grid;
      grid-template-columns: 1fr 1fr;
      padding: 0.75rem 0;
      border-bottom: 1px solid var(--line);
    }
    .results-summary tbody th:first-child {
      width: auto;
      grid-column: 1 / -1;
    }
    .results-summary tbody th,
    .results-summary tbody td {
      border: 0;
      padding: 0.45rem 0;
      white-space: normal;
    }
    .results-summary tbody td::before {
      display: block;
      font-size: 0.75rem;
      color: var(--muted);
      margin-bottom: 0.3rem;
    }
    .results-summary tbody td:nth-child(2)::before {
      content: "Принесло";
    }
    .results-summary tbody td:nth-child(3)::before {
      content: "К альтернативе";
    }
    header h2 {
      font-size: 1.3rem;
    }
    .verdict h3 {
      font-size: 1.1rem;
    }
    .reconciliation {
      font-size: 0.9rem;
    }
    .comparison-row {
      flex-wrap: wrap;
    }
    .decision-case {
      padding: 1.5rem 0;
    }
  }
</style>
