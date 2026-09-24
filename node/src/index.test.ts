import assert from "node:assert/strict"
import { test } from "node:test"

import { ready } from "./index.ts"

test("placeholder", () => {
  assert.equal(ready(), true)
})
