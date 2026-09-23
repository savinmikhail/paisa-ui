export const PUBLIC_FIXTURE_ROUTES = [
  { pathname: "/expense/monthly", fixture: "expense-baseline" },
  { pathname: "/assets/networth", fixture: "networth-ticket-04" },
  { pathname: "/assets/decision-history", fixture: "synthetic" },
  { pathname: "/assets/exposure", fixture: "synthetic" },
  { pathname: "/resilience", fixture: "resilience" },
  { pathname: "/assets/money-flow", fixture: "money-flow" }
] as const;

export const PUBLIC_FIXTURE_IDS = PUBLIC_FIXTURE_ROUTES.map((route) => route.fixture);

export type PublicFixtureId = (typeof PUBLIC_FIXTURE_ROUTES)[number]["fixture"];

export function isPublicFixtureRoute(pathname: string, value: string | null): value is PublicFixtureId {
  return PUBLIC_FIXTURE_ROUTES.some((route) => route.pathname === pathname && route.fixture === value);
}
