# Bugua Character Art Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add six-line hexagram character art to bugua text output while preserving existing labels and moving-line details.

**Architecture:** `plugin/bugua/main.go` already formats the output through `yaoValues.format()`. Update that formatter to render each yao with its name, visual line, text description, and moving marker. `summary()` and `prompt()` should continue to call this shared formatter.

**Tech Stack:** Go 1.26, standard library `strings`, existing `go test` package tests.

---

### Task 1: Add Failing Formatter Tests

**Files:**
- Modify: `plugin/bugua/bugua_test.go`
- Test: `plugin/bugua/bugua_test.go`

**Step 1: Write the failing test**

Add a test that calls:

```go
got := yaoValues{oldYin, youngYang, youngYin, oldYang, youngYin, oldYang}.format()
```

Assert that the output contains these lines:

```text
上九 ━━━━━ 阳爻 动
六五 ━  ━ 阴爻
九四 ━━━━━ 阳爻 动
六三 ━  ━ 阴爻
九二 ━━━━━ 阳爻
初六 ━  ━ 阴爻 动
```

**Step 2: Run test to verify it fails**

Run: `go test ./plugin/bugua -run TestFormatYaoValuesIncludesCharacterArt -count=1`

Expected: FAIL because existing `format()` does not render `━━━━━` or `━  ━`.

**Step 3: Write minimal implementation**

In `plugin/bugua/main.go`, add a small helper:

```go
func yaoLineArt(value int) string {
	if value == youngYang || value == oldYang {
		return "━━━━━"
	}
	return "━  ━"
}
```

Update `yaoValues.format()` to insert the line art between the line name and text.

**Step 4: Run test to verify it passes**

Run: `go test ./plugin/bugua -run TestFormatYaoValuesIncludesCharacterArt -count=1`

Expected: PASS.

### Task 2: Verify Summary Output

**Files:**
- Modify: `plugin/bugua/bugua_test.go`
- Test: `plugin/bugua/bugua_test.go`

**Step 1: Write the failing test**

Create a `divinationResult` with fixed yao values and assert `summary()` contains:

```text
卦象:
上九 ━━━━━ 阳爻 动
```

**Step 2: Run test to verify behavior**

Run: `go test ./plugin/bugua -run TestSummaryIncludesHexagramCharacterArt -count=1`

Expected: PASS after Task 1 because summary uses `yaoValues.format()`.

**Step 3: Run package tests**

Run: `go test ./plugin/bugua -count=1`

Expected: PASS.

### Task 3: Commit

**Files:**
- Modify: `plugin/bugua/main.go`
- Modify: `plugin/bugua/bugua_test.go`
- Create: `docs/plans/2026-06-09-bugua-character-art-design.md`
- Create: `docs/plans/2026-06-09-bugua-character-art.md`

**Step 1: Review diff**

Run: `git diff -- plugin/bugua/main.go plugin/bugua/bugua_test.go docs/plans/2026-06-09-bugua-character-art-design.md docs/plans/2026-06-09-bugua-character-art.md`

**Step 2: Commit scoped files**

```bash
git add plugin/bugua/main.go plugin/bugua/bugua_test.go docs/plans/2026-06-09-bugua-character-art-design.md docs/plans/2026-06-09-bugua-character-art.md
git commit -m "feat: add bugua hexagram character art"
```
