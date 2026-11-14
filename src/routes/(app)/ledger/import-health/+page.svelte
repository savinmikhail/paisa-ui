<script lang="ts">
  import { onMount } from "svelte";
  import { ajax } from "$lib/utils";

  type Account = { account: string; last_operation: string; confirmed_through: string | null; imported_at?: string | null; last_sync?: string | null };
  type Service = { service: string; last_success: string | null };
  type Report = { generated_at: string; accounts: Account[]; services: Service[]; statements?: { bank: string; period_end: string; imported_at: string }[] | null; issues: { id: string }[] };

  const serviceNames: Record<string, string> = {
    bybit: "Bybit",
    "tbank-business": "Т-Банк Бизнес",
    "tbank-invest": "Т-Инвестиции",
    "tbank-mail": "Т-Банк · выписки из почты",
    freedom: "Freedom",
    debank: "DeBank",
    "wallet-history": "Криптокошельки"
  };
  const formatter = new Intl.DateTimeFormat("ru-RU", { dateStyle: "medium" });
  let report: Report | null = null;
  let loading = true;
  let failed = false;
  const statementBanks = [
    { name: "Halyk", accounts: ["Assets:Checking:Halyk:"] },
    { name: "Bakai", accounts: ["Assets:Checking:Bakai:"] },
    { name: "Т-Банк", accounts: ["Assets:Checking:TBank:Black", "Assets:Savings:TBank:Savings Account:RUB", "Liabilities:CreditCards:TBank:"] },
    { name: "Сбер", accounts: ["Assets:Checking:Sber:MIR Classic:"] },
    { name: "Альфа", accounts: ["Assets:Checking:Alfa:"] },
    { name: "Арцахбанк", accounts: ["Assets:Checking:Artsakhbank:"] }
  ];

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

  $: banks = statementBanks.map((bank) => ({
    name: bank.name,
    accounts: (report?.accounts ?? []).filter(({ account }) =>
      bank.accounts.some((prefix) => account.startsWith(prefix))
    )
  }));

  function accountLabel(account: string) {
    return account.split(":").slice(3).join(" · ") || account.split(":").slice(2).join(" · ");
  }
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
      <p class="mb-5">Отчёт от {date(report.generated_at)}</p>
      <div class="is-flex is-flex-wrap-wrap is-align-items-baseline is-justify-content-space-between mb-3 header">
        <h2 class="title is-5 mb-0">Ручные выписки</h2>
        <a href="/ledger/import">Загрузить выписку</a>
      </div>
      <div class="table-container mb-5">
        <table class="table is-fullwidth statements">
          <thead><tr><th>Банк и счёт</th><th>Импортировано по</th></tr></thead>
          {#each banks as bank (bank.name)}
            <tbody>
              <tr><th scope="rowgroup" colspan="2">{bank.name}</th></tr>
              {#each bank.accounts as account (account.account)}
                <tr>
                  <th scope="row" class="account-label">{accountLabel(account.account)}</th>
                  <td>
                    {#if account.confirmed_through}
                      {date(account.confirmed_through)} · {age(account.confirmed_through)} дн. назад
                    {:else if account.imported_at}
                      Период неизвестен <small>Последняя загрузка {date(account.imported_at)}</small>
                    {:else}
                      Период неизвестен <small>Последняя операция {date(account.last_operation)}</small>
                    {/if}
                  </td>
                </tr>
              {:else}
                <tr><td colspan="2">Нет данных в отчёте</td></tr>
              {/each}
            </tbody>
          {/each}
        </table>
      </div>
      <h2 class="title is-5">Автоматические загрузки</h2>
      <p class="mb-3">Запустить вручную <code>make sync</code></p>
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
    {/if}
  </div>
</section>

<style>
  .header { gap: 0.75rem; }
  .statements .account-label { padding-inline-start: 1.25rem; font-weight: normal; }
  .statements small { display: block; }
  td { font-variant-numeric: tabular-nums; }
</style>
