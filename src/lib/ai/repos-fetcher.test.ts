import assert from "node:assert/strict";
import * as fs from "node:fs";
import * as path from "node:path";
import { describe, it } from "node:test";

import { isAiMlRelated, parseGitHubTrending } from "./repos-fetcher";

describe("repos fetcher", () => {
  it("parses valid GitHub trending HTML successfully", () => {
    const fixturePath = path.join(__dirname, "__fixtures__", "github-trending.html");
    const html = fs.readFileSync(fixturePath, "utf-8");
    const repos = parseGitHubTrending(html);

    assert.equal(repos.length, 2);
    assert.equal(repos[0].repoFullName, "meta-llama/llama3");
    assert.equal(repos[0].url, "https://github.com/meta-llama/llama3");
    assert.equal(repos[0].description, "Llama 3 model code and model definitions. An advanced LLM architecture.");
    assert.equal(repos[0].language, "Python");
    assert.equal(repos[0].stars, 15234);
    assert.equal(repos[0].starsToday, 1245);

    assert.equal(repos[1].repoFullName, "owner/webscraper");
    assert.equal(repos[1].language, "TypeScript");
    assert.equal(repos[1].stars, 456);
    assert.equal(repos[1].starsToday, 23);
  });

  it("evaluates AI/ML relevance correctly", () => {
    assert.equal(
      isAiMlRelated(
        "meta-llama/llama3",
        "Llama 3 model code and model definitions. An advanced LLM architecture.",
      ),
      true,
    );
    assert.equal(isAiMlRelated("owner/my-cool-agent", "An agent system built on top of Claude."), true);
    assert.equal(isAiMlRelated("owner/simple-web-app", "Just a normal web dashboard with react and tailwind."), false);
  });

  it("handles empty or malformed HTML without throwing", () => {
    const repos = parseGitHubTrending("<html><body>Nothing here</body></html>");
    assert.equal(repos.length, 0);
  });
});
