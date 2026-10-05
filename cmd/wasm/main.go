//go:build js && wasm

package main

import (
	"syscall/js"
)

func rollDiceWrapper(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return map[string]any{
			"ok":    false,
			"error": "missing roll expression",
		}
	}

	expr := args[0].String()
	resp := EvaluateRoll(expr)

	if !resp.OK {
		return map[string]any{
			"ok":    false,
			"error": resp.Error,
		}
	}

	resultsSlice := make([]any, len(resp.Results))
	for i, r := range resp.Results {
		resultsSlice[i] = map[string]any{
			"result": r.Result,
			"symbol": r.Symbol,
		}
	}

	return map[string]any{
		"ok":           true,
		"input":        resp.Input,
		"rendered":     resp.Rendered,
		"total":        resp.Total,
		"successes":    resp.Successes,
		"hasSuccesses": resp.HasSuccesses,
		"results":      resultsSlice,
		"text":         resp.Text,
	}
}

func main() {
	js.Global().Set("rollDice", js.FuncOf(rollDiceWrapper))
	// Prevent Go program from exiting so JS callbacks remain active
	select {}
}
