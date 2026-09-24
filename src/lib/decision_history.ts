export type DecisionHistoryFieldStatus = "known" | "partial" | "unknown";

export interface DecisionHistoryField {
  status: DecisionHistoryFieldStatus;
  value: string;
  detail: string;
}

export type DecisionHistoryMetricTone = "accent" | "positive" | "negative" | "neutral";

export interface DecisionHistoryMetric {
  label: string;
  value: string;
  detail?: string;
  tone?: DecisionHistoryMetricTone;
}

export interface DecisionHistoryTimelineItem {
  label: string;
  value: string;
  detail: string;
}

export interface DecisionHistoryBranch {
  id: string;
  number: string;
  title: string;
  period: string;
  decision: string;
  headline: string;
  metrics: DecisionHistoryMetric[];
  timeline?: DecisionHistoryTimelineItem[];
  outcome: string;
  sourceRefs: string[];
  caveats: string[];
}

export interface DecisionHistoryTheme {
  id: string;
  title: string;
  period: string;
  decision: string;
  headline: string;
  summary: string;
  metrics: DecisionHistoryMetric[];
  branches: DecisionHistoryBranch[];
  sourceRefs: string[];
  caveats: string[];
}

export interface DecisionHistoryOverview {
  title: string;
  subtitle: string;
  headline: string;
  valuationDate: string;
  metrics: DecisionHistoryMetric[];
  note: string;
}

export interface DecisionHistoryResult {
  lastPostingDate?: string;
  asOf: string;
  invested: number;
  withdrawn: number;
  currentValue: number;
  netResult: number;
  fees: number;
  taxes: number;
  taxRefund: number;
  benchmarkRate: number;
  benchmarkValue: number;
  actualComparableValue: number;
  difference: number;
  payoutCarry: number;
  coverage: string;
  method: string;
  fxEffect?: number;
  conversionCost?: number;
  bondCoupons?: number;
  bondPriceAndRedemption?: number;
  fundingRoutes?: {
    route: string;
    receipts: number;
    firstDate: string;
    lastDate: string;
    rubSent: number;
    usdReceived: number;
    effectiveRate: number;
    referenceRate: number;
    gapRub: number;
    gapPercent: number;
  }[];
  segments?: { label: string; invested: number; netResult: number; difference: number }[];
  details: {
    date: string;
    label: string;
    amount: number;
    source: string;
    quantity?: number;
    currency?: string;
    rate?: number;
  }[];
}

export interface DecisionHistoryCase {
  result?: DecisionHistoryResult;
  id: string;
  title: string;
  period: string;
  decision: string;
  context: string;
  amount: {
    label: string;
    currency: string;
  };
  outcome: string;
  currentStatus: string;
  sourceRefs: string[];
  caveats: string[];
  attribution: {
    fx: DecisionHistoryField;
    fees: DecisionHistoryField;
    taxes: DecisionHistoryField;
    rubCounterfactual: DecisionHistoryField;
  };
}

export interface PrivateDecisionHistoryPayload {
  schemaVersion: "decision-history/v1";
  overview?: DecisionHistoryOverview;
  themes?: DecisionHistoryTheme[];
  cases: DecisionHistoryCase[];
}

export type DecisionHistoryLoadResult =
  | { kind: "ready"; payload: PrivateDecisionHistoryPayload }
  | { kind: "unavailable"; reason: string };

export const syntheticDecisionHistory: PrivateDecisionHistoryPayload = {
  schemaVersion: "decision-history/v1",
  cases: [
    {
      id: "aldar-broker-account",
      title: "Брокерский счёт Aldar",
      period: "2022-04 → 2022-12",
      decision: "Открыть брокерский счёт, чтобы добавить долю активов в AED.",
      context:
        "Диверсифицироваться за пределами отечественной банковской инфраструктуры, отдельно проверив хранение активов, доступ и ликвидность.",
      amount: { label: "420k", currency: "Эквивалент в RUB" },
      outcome:
        "Счёт остался частью карты распределения; исходный тезис о диверсификации подтверждён лишь частично.",
      currentStatus: "На контроле",
      sourceRefs: ["DH-SYN-001 decision note", "DH-SYN-001 account summary"],
      caveats: [
        "В публичном контракте нет проверенного ряда полной доходности.",
        "Комиссии и налоги на уровне выписок пока не входят в этот фикстурный набор."
      ],
      attribution: {
        fx: {
          status: "partial",
          value: "Движение AED / RUB разделено только по направлению",
          detail: "Точный путь конвертации оставлен приватному адаптеру."
        },
        fees: {
          status: "unknown",
          value: "Не предоставлено",
          detail: "Для брокерских и переводных комиссий нужны подтверждающие источники."
        },
        taxes: {
          status: "unknown",
          value: "Не предоставлено",
          detail: "Налоговый режим не выводится из записи о решении."
        },
        rubCounterfactual: {
          status: "unknown",
          value: "Не рассчитан",
          detail: "Нужно выбрать альтернативу в RUB и зафиксировать горизонт сравнения."
        }
      }
    },
    {
      id: "iis-contribution",
      title: "ИИС",
      period: "2021-01 → 2024-12",
      decision:
        "Использовать заблокированный инвестиционный счёт для многолетнего распределения капитала.",
      context:
        "Обменять краткосрочный доступ на льготную налоговую оболочку и более длинный горизонт для облигаций и акций.",
      amount: { label: "650k", currency: "Внесено в RUB" },
      outcome:
        "Период пополнения зафиксирован; возможность вывода и налоговый эффект остаются отдельными вопросами.",
      currentStatus: "Заблокировано / требуется проверка",
      sourceRefs: ["DH-SYN-002 decision note", "DH-SYN-002 contribution ledger"],
      caveats: [
        "Публичный фикстурный набор не утверждает право на какой-либо налоговый вычет.",
        "Текущая доступность счёта указана как статус, а не как вывод."
      ],
      attribution: {
        fx: {
          status: "known",
          value: "Решение в RUB",
          detail: "Для синтетической базовой суммы валютная атрибуция не требуется."
        },
        fees: {
          status: "partial",
          value: "Комиссии за счёт исключены",
          detail: "Издержки на уровне инструментов не входят в этот пример контракта."
        },
        taxes: {
          status: "partial",
          value: "Режим оболочки отмечен, результат неизвестен",
          detail: "Приватный адаптер должен приложить проверенное налоговое подтверждение."
        },
        rubCounterfactual: {
          status: "unknown",
          value: "Не рассчитан",
          detail: "Сопоставимый ликвидный вариант распределения в RUB не выбран."
        }
      }
    },
    {
      id: "large-ofz-purchase",
      title: "Крупная покупка ОФЗ",
      period: "2023-10-06",
      decision: "Сделать концентрированную покупку ОФЗ в период высоких ставок.",
      context:
        "Предпочесть понятный купонный денежный поток и известный срок погашения добавлению ещё одной позиции в акциях.",
      amount: { label: "1.8m", currency: "Покупка в RUB" },
      outcome:
        "Денежный поток и переоценка по рынку разошлись; фактический результат требует сверки на уровне инструмента.",
      currentStatus: "Требует сверки",
      sourceRefs: ["DH-SYN-003 decision note", "DH-SYN-003 bond lot summary"],
      caveats: [
        "Купон, накопленный доход и поступления от продажи намеренно не раскрываются здесь подробно.",
        "Экран фиксирует границу решения, но не оценивает сделку."
      ],
      attribution: {
        fx: {
          status: "known",
          value: "Инструмент в RUB",
          detail: "Валютный эффект не входит в синтетический путь инструмента."
        },
        fees: {
          status: "partial",
          value: "Брокерские издержки ожидаются",
          detail: "Контракт предполагает сумму комиссии со ссылкой на источник."
        },
        taxes: {
          status: "partial",
          value: "Налоговый режим лота ожидается",
          detail: "Налог на купон и выбытие отделён от движения цены."
        },
        rubCounterfactual: {
          status: "partial",
          value: "Денежная альтернатива выбрана для сравнения",
          detail: "Перед расчётом нужно зафиксировать ставку, длительность и горизонт."
        }
      }
    },
    {
      id: "summer-rub-to-usd-shift",
      title: "Летний переход из RUB в USD",
      period: "2024-06 → 2024-08",
      decision: "Перевести транш из RUB в USD в течение лета.",
      context:
        "Создать резерв в твёрдой валюте, сохранив достаточно ликвидности в RUB для ближайших обязательств.",
      amount: { label: "1.25m", currency: "Эквивалент в RUB" },
      outcome:
        "Валютная экспозиция изменилась; результат решения нельзя отделить от валютного эффекта, комиссий, налогов и времени операции.",
      currentStatus: "На контроле",
      sourceRefs: ["DH-SYN-004 decision note", "DH-SYN-004 conversion summary"],
      caveats: [
        "Фикстурный набор не содержит квитанций об обмене, данных контрагентов или идентификаторов счетов.",
        "Движение валюты не считается прибылью без выбранного сценария сравнения."
      ],
      attribution: {
        fx: {
          status: "partial",
          value: "Направление зафиксировано; путь курса не указан",
          detail: "Приватный адаптер предоставляет датированные подтверждения конвертации."
        },
        fees: {
          status: "unknown",
          value: "Не предоставлено",
          detail: "Для спреда и комиссий за перевод нужны подтверждающие источники."
        },
        taxes: {
          status: "unknown",
          value: "Не предоставлено",
          detail: "Налоговый результат не выводится из конвертации валюты."
        },
        rubCounterfactual: {
          status: "partial",
          value: "Резерв в RUB выбран для сравнения",
          detail: "Горизонт и ценность ликвидности ещё нужно зафиксировать."
        }
      }
    }
  ]
};

syntheticDecisionHistory.cases.forEach((item, index) => {
  const difference = [2400, -1800, -3200, 4100][index];
  const netResult = 8000 + difference;
  item.result = {
    asOf: "2025-12-31",
    invested: 100000,
    withdrawn: 5000,
    currentValue: 103000 + difference,
    netResult,
    fees: 200,
    taxes: 300,
    taxRefund: 0,
    benchmarkRate: 9,
    benchmarkValue: 108200,
    actualComparableValue: 108200 + difference,
    difference,
    payoutCarry: 200,
    coverage: "Иллюстративный расчёт",
    method: "Синтетические суммы. Одинаковые даты вложений и ставка сравнения.",
    details: [{ date: "2025-01-01", label: "Взнос", amount: 100000, source: "SYN-001" }]
  };
  if (item.id === "large-ofz-purchase") {
    item.result.bondCoupons = 9000;
    item.result.bondPriceAndRedemption = netResult - 9000;
  }
  if (item.id === "summer-rub-to-usd-shift") {
    item.result.conversionCost = 7000;
    item.result.fxEffect = netResult + item.result.conversionCost + item.result.fees;
    item.result.fundingRoutes = [
      {
        route: "Банк А → Банк Б",
        receipts: 2,
        firstDate: "2025-06-01",
        lastDate: "2025-06-15",
        rubSent: 100000,
        usdReceived: 1200,
        effectiveRate: 83.33,
        referenceRate: 80,
        gapRub: 4000,
        gapPercent: 4.17
      }
    ];
  }
});

export const privateDecisionHistoryAdapter = {
  async load(fetcher: typeof fetch): Promise<DecisionHistoryLoadResult> {
    const response = await fetcher("/api/decision_history");
    if (!response.ok) {
      return {
        kind: "unavailable",
        reason:
          response.status === 404
            ? "Private adapter is not connected."
            : `Adapter returned ${response.status}.`
      };
    }

    const payload = (await response.json()) as PrivateDecisionHistoryPayload;
    if (payload.schemaVersion !== "decision-history/v1" || !Array.isArray(payload.cases)) {
      return { kind: "unavailable", reason: "Adapter returned an unsupported contract." };
    }

    return { kind: "ready", payload };
  }
};
