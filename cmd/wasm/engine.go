package main

import (
	"github.com/darkliquid/roll"
)

// DieResult contains the individual die evaluation outcome.
type DieResult struct {
	Result int    `json:"result"`
	Symbol string `json:"symbol"`
}

// RollResponse represents the structured evaluation result sent to the browser.
type RollResponse struct {
	OK           bool        `json:"ok"`
	Input        string      `json:"input,omitempty"`
	Rendered     string      `json:"rendered,omitempty"`
	Total        int         `json:"total,omitempty"`
	Successes    int         `json:"successes,omitempty"`
	HasSuccesses bool        `json:"hasSuccesses,omitempty"`
	Results      []DieResult `json:"results,omitempty"`
	Text         string      `json:"text,omitempty"`
	Error        string      `json:"error,omitempty"`
}

// EvaluateRoll parses and evaluates a dice expression string.
func EvaluateRoll(expr string) RollResponse {
	program, err := roll.CompileString(expr)
	if err != nil {
		return RollResponse{
			OK:    false,
			Error: err.Error(),
		}
	}

	result, err := roll.EvaluateProgram(program)
	if err != nil {
		return RollResponse{
			OK:    false,
			Error: err.Error(),
		}
	}

	text, _ := roll.ParseString(expr)

	var diceResults []DieResult
	for _, r := range result.Results {
		diceResults = append(diceResults, DieResult{
			Result: r.Result,
			Symbol: r.Symbol,
		})
	}

	hasSuccesses := false
	for _, term := range program.DiceTerms {
		if term.Success != nil || term.Failure != nil {
			hasSuccesses = true
			break
		}
	}
	if !hasSuccesses {
		for _, term := range program.GroupTerms {
			if term.Success != nil || term.Failure != nil {
				hasSuccesses = true
				break
			}
		}
	}

	return RollResponse{
		OK:           true,
		Input:        expr,
		Rendered:     program.String(),
		Total:        result.Total,
		Successes:    result.Successes,
		HasSuccesses: hasSuccesses,
		Results:      diceResults,
		Text:         text,
	}
}
