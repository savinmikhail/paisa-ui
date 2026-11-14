<script lang="ts">
  import { onMount } from "svelte";
  import { ajax } from "$lib/utils";

  type Account = { account: string; last_operation: string; confirmed_through: string | null };
  type Service = { service: string; last_success: string | null };
  type Report = { generated_at: string; accounts: Account[]; services: Service[]; statements?: { bank: string; period_end: string; imported_at: string }[] | null; issues: { id: string }[] };

  const serviceNames: Record<string, string> = {
    bybit: "Bybit",
    "tbank-business": "Т-Банк Бизнес",
    "tbank-invest": "Т-Инвестиции",
    "tbank-mail": "Т-Банк · выписки из почты",
    freedom: "Freedom",
    "wallet-history": "Криптокошельки"
  };
  const formatter = new Intl.DateTimeFormat("ru-RU", { dateStyle: "medium" });
  let report: Report | null = null;
  let loading = true;
  let failed = false;
  let search = "";

  function date(value: string) {
    return formatter.format(new Date(value));
  }

  function age(value: string) {
    const now = new Date();
    const then = new Date(value);
    const day = (d: Date) => Date.UTC(d.getFullYear(), d.getMonth(), d.getDate());
    return Math.max(0, (day(now) - day(then)) / 86400000);
  }

  async function load() {
    loading = true;
    failed = false;
    try {
      const result = await ajax("/api/import_health", { background: true });
      if (!("generated_at" in result)) throw new Error("Import health unavailable");
      report = result;
    } catch {
      failed = true;
    } finally {
      loading = false;
    }
  }

  $: accounts = (report?.accounts ?? [])
    .filter(({ account }) => account.toLowerCase().includes(search.trim().toLowerCase()))
    .sort((a, b) => (a.confirmed_through ?? a.last_operation).localeCompare(b.confirmed_through ?? b.last_operation));
  $: issues = new Set(report?.issues.map(({ id }) => id));

  onMount(() => { void load(); });
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="is-flex is-flex-wrap-wrap is-align-items-center is-justify-content-space-between mb-4 header">
      <h1 class="title is-3 mb-0">Состояние импорта</h1>
      <button class="button" disabled={loading} on:click={load}>Перечитать отчёт</button>
    </div>
    {#if loading}
      <p role="status">Загружаю отчёт…</p>
    {:else if failed}
      <p role="alert">Отчёт недоступен. Выполните <code>make sync</code> в каталоге Paisa и перечитайте отчёт.</p>
    {:else if report}
      <p class="mb-5">Отчёт от {date(report.generated_at)}. Чтобы запустить автоматические загрузки, выполните <code>make sync</code>.</p>
      <h2 class="title is-5">Автоматические загрузки</h2>
      <div class="table-container mb-5">
        <table class="table is-fullwidth">
          <thead><tr><th>Источник</th><th>Последний успех</th><th>Состояние</th></tr></thead>
          <tbody>
            {#each report.services as service (service.service)}
              <tr>
                <th scope="row">{serviceNames[service.service] ?? service.service}</th>
                <td>{service.last_success ? `${date(service.last_success)} · ${age(service.last_success)} дн. назад` : "Нет отметки"}</td>
                <td>{issues.has(`error:${service.service}`) ? "Ошибка загрузки" : issues.has(`stale:${service.service}`) ? "Давно не обновлялось" : service.last_success ? "Успешно" : "Проверить настройку"}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <h2 class="title is-5">Импортированные выписки</h2>
      {#if report.statements?.length}
        <div class="table-container mb-5">
          <table class="table is-fullwidth">
            <thead><tr><th>Банк</th><th>Импортировано по</th><th>Загружено</th></tr></thead>
            <tbody>
              {#each report.statements as statement (statement.bank)}
                <tr>
                  <th scope="row">{statement.bank}</th>
                  <td>{date(statement.period_end)} · {age(statement.period_end)} дн. назад</td>
                  <td>{date(statement.imported_at)}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else}
        <p class="mb-5">В отчёте пока нет данных о периодах выписок.</p>
      {/if}
      <h2 class="title is-5">Покрытие по счетам</h2>
      <p class="mb-3">Дата операции показывает последнюю запись в Ledger. Покрытие подтверждается по периоду выписки. <a href="/ledger/import">Загрузить выписку</a></p>
      <label class="label" for="account-search">Найти банк или счёт</label>
      <input class="input mb-4 search" id="account-search" type="search" bind:value={search} placeholder="Например, Halyk или Bakai" />
      {#if accounts.length === 0}
        <p>Счета не найдены.</p>
      {:else}
        <div class="table-container">
          <table class="table is-fullwidth">
            <thead><tr><th>Счёт</th><th>Последняя операция</th><th>Выписка проверена по</th><th>Действие</th></tr></thead>
            <tbody>
              {#each accounts as account (account.account)}
                <tr>
                  <th scope="row">{account.account.replace(/:/g, " / ")}</th>
                  <td>{date(account.last_operation)} · {age(account.last_operation)} дн. назад</td>
                  <td>{account.confirmed_through ? `${date(account.confirmed_through)} · ${age(account.confirmed_through)} дн. назад` : "Не подтверждено"}</td>
                  <td>{issues.has(`account:${account.account}`) ? "Проверить выписку" : "—"}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    {/if}
  </div>
</section>

<style>
  .header { gap: 0.75rem; }
  .search { max-width: 26rem; }
  td { font-variant-numeric: tabular-nums; }
</style>
