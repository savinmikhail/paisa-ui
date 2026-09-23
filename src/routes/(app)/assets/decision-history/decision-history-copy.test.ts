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
});
