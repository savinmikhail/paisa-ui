import assert from "node:assert/strict";
import { test } from "node:test";
import { build } from "esbuild";
import ts from "typescript";
import { readFile } from "node:fs/promises";

test("route preloading does not activate the foreground page spinner", async () => {
  const result = await build({
    entryPoints: ["src/routes/(app)/+layout.ts"],
    bundle: true,
    write: false,
    format: "esm",
    plugins: [
      {
        name: "layout-dependencies",
        setup(build) {
          build.onLoad({ filter: /\.ts$/ }, async (args) => ({
            contents: ts.transpileModule(await readFile(args.path, "utf8"), {
              compilerOptions: { module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2022 }
            }).outputText,
            loader: "js"
          }));
          build.onResolve({ filter: /^\$lib\// }, (args) => ({
            path: args.path,
            namespace: "layout-test"
          }));
          build.onLoad({ filter: /.*/, namespace: "layout-test" }, (args) => ({
            contents: args.path.endsWith("/utils")
              ? `
        export const foregroundRequests = [];
        export async function ajax(route, options) {
          if (!options?.background) foregroundRequests.push(route);
          return {config: {}, now: null};
        }
        export function configUpdated() {}
        export function setNow() {}
      `
              : "export function isPublicFixtureRoute() { return false; }"
          }));
        }
      }
    ],
    footer: { js: "export { foregroundRequests };" }
  });
  // Expose the same dependency instance used by the real route loader.
  const code = result.outputFiles[0].text;
  const module = await import(
    `data:text/javascript;base64,${Buffer.from(code).toString("base64")}`
  );
  await module.load({ url: new URL("http://localhost/income") });
  assert.deepEqual(module.foregroundRequests, []);
});
