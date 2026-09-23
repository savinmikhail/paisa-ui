import assert from "node:assert/strict";
import { describe, test } from "node:test";
import { isPublicFixtureRoute, PUBLIC_FIXTURE_ROUTES } from "./fixture_routes";

describe("public fixture route contract", () => {
  test("keeps every documented screenshot fixture in the read-only allowlist", () => {
    assert.deepEqual(PUBLIC_FIXTURE_ROUTES, [
      { pathname: "/expense/monthly", fixture: "expense-baseline" },
      { pathname: "/assets/networth", fixture: "networth-ticket-04" },
      { pathname: "/assets/decision-history", fixture: "synthetic" },
      { pathname: "/assets/exposure", fixture: "synthetic" },
      { pathname: "/resilience", fixture: "resilience" },
      { pathname: "/assets/money-flow", fixture: "money-flow" }
    ]);
  });

  test("does not broaden the config bypass to another route or query value", () => {
    assert.equal(isPublicFixtureRoute("/expense/monthly", "expense-baseline"), true);
    assert.equal(isPublicFixtureRoute("/assets/exposure", "synthetic"), true);
    assert.equal(isPublicFixtureRoute("/assets/decision-history", "private"), false);
    assert.equal(isPublicFixtureRoute("/assets/exposure", "../../data"), false);
    assert.equal(isPublicFixtureRoute("/assets/networth", "synthetic"), false);
    assert.equal(isPublicFixtureRoute("/assets/exposure", null), false);
  });
});
