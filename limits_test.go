package roll

import (
	"reflect"
	"testing"
)

func TestApplyLimit(t *testing.T) {
	tests := []struct {
		name string
		op   LimitOp
		in   []int
		want []int
	}{
		{name: "keep highest drops excess duplicate", op: LimitOp{Amount: 3, Type: KeepHighest}, in: []int{6, 2, 2, 2}, want: []int{6, 2, 2}},
		{name: "keep highest single", op: LimitOp{Amount: 1, Type: KeepHighest}, in: []int{6, 2, 2, 2}, want: []int{6}},
		{name: "keep lowest drops excess duplicate", op: LimitOp{Amount: 2, Type: KeepLowest}, in: []int{6, 2, 2, 2}, want: []int{2, 2}},
		{name: "drop highest removes one duplicate max", op: LimitOp{Amount: 1, Type: DropHighest}, in: []int{6, 6, 4, 6}, want: []int{6, 6, 4}},
		{name: "drop lowest removes one duplicate min", op: LimitOp{Amount: 1, Type: DropLowest}, in: []int{5, 4, 4, 4}, want: []int{5, 4, 4}},
		{name: "keep highest over count keeps all", op: LimitOp{Amount: 5, Type: KeepHighest}, in: []int{6, 2, 2, 2}, want: []int{6, 2, 2, 2}},
		{name: "keep highest all tied", op: LimitOp{Amount: 2, Type: KeepHighest}, in: []int{3, 3, 3, 3}, want: []int{3, 3}},
		{name: "keep lowest all tied", op: LimitOp{Amount: 3, Type: KeepLowest}, in: []int{3, 3, 3, 3}, want: []int{3, 3, 3}},
		{name: "drop highest all tied", op: LimitOp{Amount: 2, Type: DropHighest}, in: []int{3, 3, 3, 3}, want: []int{3, 3}},
		{name: "drop lowest all tied", op: LimitOp{Amount: 1, Type: DropLowest}, in: []int{3, 3, 3, 3}, want: []int{3, 3, 3}},
		{name: "keep highest zero drops all", op: LimitOp{Amount: 0, Type: KeepHighest}, in: []int{6, 2, 2, 2}, want: []int{}},
		{name: "keep lowest with distinct values", op: LimitOp{Amount: 2, Type: KeepLowest}, in: []int{5, 3, 3, 1}, want: []int{3, 1}},
		{name: "drop lowest with distinct values", op: LimitOp{Amount: 1, Type: DropLowest}, in: []int{5, 3, 3, 1}, want: []int{5, 3, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &Result{}
			for _, v := range tt.in {
				result.Results = append(result.Results, DieRoll{Result: v})
			}

			applyLimit(&tt.op, result)

			got := make([]int, len(result.Results))
			for i, roll := range result.Results {
				got[i] = roll.Result
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Fatalf("results mismatch: got=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestApplyLimitDoesNotMutateInputOrder(t *testing.T) {
	result := &Result{Results: []DieRoll{{Result: 6}, {Result: 2}, {Result: 2}, {Result: 2}}}

	// A nil limit is a no-op and must not reorder the underlying slice.
	applyLimit(nil, result)

	got := make([]int, len(result.Results))
	for i, roll := range result.Results {
		got[i] = roll.Result
	}
	if want := []int{6, 2, 2, 2}; !reflect.DeepEqual(want, got) {
		t.Fatalf("nil limit mutated results: got=%v want=%v", got, want)
	}
}
