# Bugua Character Art Design

## Goal

Add readable hexagram character art to the `/plugin/bugua` text output.

## Approved Output

Keep the existing text summary and add a six-line visual form under `卦象`, ordered from top line to bottom line:

```text
上九  ━━━━━  动
六五  ━  ━
九四  ━━━━━
六三  ━  ━  动
九二  ━━━━━
初六  ━  ━
```

Yang lines render as `━━━━━`; yin lines render as `━  ━`; moving lines append `动`.

## Architecture

The output is already centralized in `plugin/bugua/main.go` through `divinationResult.summary()`, `divinationResult.prompt()`, and `yaoValues.format()`. Extend `yaoValues.format()` so both the user-facing summary and AI prompt receive the same character art. Keep lookup, casting, and AI integration unchanged.

## Testing

Add focused Go tests in `plugin/bugua/bugua_test.go` for `yaoValues.format()` and for the summary containing the visual lines. Run `go test ./plugin/bugua`.
