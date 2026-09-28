import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { describe, test } from "node:test";

const pageSource = await readFile(new URL("./+page.svelte", import.meta.url), "utf8");

describe("decision history page copy contract", () => {
  test("keeps fixture selection separate from the live Russian UI", () => {
    assert.match(pageSource, /let isFixture = false/);
    assert.match(pageSource, /if \(isFixture\) \{/);
    assert.doesNotMatch(pageSource, /isSynthetic/);
    assert.doesNotMatch(pageSource, /isFixture\s*\?/);
    assert.doesNotMatch(
      pageSource,
      /Decision history|Loading decision history|public UI checkout|Synthetic Fixture|Public UI contract/
    );
  });

  test("keeps loading and unavailable states actionable in Russian", () => {
    assert.match(pageSource, /Загрузка истории решений/);
    assert.match(pageSource, /История решений недоступна/);
    assert.match(pageSource, /Проверьте подключение и повторите попытку/);
    assert.match(pageSource, /Повторить загрузку/);
  });

  test("shows stoozing earnings and credit interest as separate facts", () => {
    assert.match(pageSource, /Заработано/);
    assert.match(pageSource, /Проценты банку/);
    assert.match(pageSource, /Средний покрытый долг/);
    assert.match(pageSource, /Platinum и Drive/);
    assert.match(pageSource, /Незавершённый\s+период не прогнозируется/);
  });

  test("shows Bybit yield period, working principal and simple APR", () => {
    assert.match(pageSource, /Что дал Bybit/);
    assert.match(pageSource, /Средняя сумма в работе/);
    assert.match(pageSource, /Максимум в работе/);
    assert.match(pageSource, /простой годовой эквивалент/i);
    assert.match(pageSource, /фактическому времени каждой подписки и погашения/);
  });
});
