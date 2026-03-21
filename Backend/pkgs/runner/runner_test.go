package runner

import (
	"testing"
)

func TestCompare(t *testing.T) {
	r := &runner{}

	tests := []struct {
		name         string
		expected     *QueryResult
		actual       *QueryResult
		orderMatters bool
		isCorrect    bool
	}{
		{
			name: "Match with orderMatters=true",
			expected: &QueryResult{
				RowCount: 2,
				Rows: [][]interface{}{
					{"1", "Alice"},
					{"2", "Bob"},
				},
			},
			actual: &QueryResult{
				RowCount: 2,
				Rows: [][]interface{}{
					{"1", "Alice"},
					{"2", "Bob"},
				},
			},
			orderMatters: true,
			isCorrect:    true,
		},
		{
			name: "Mismatch order with orderMatters=true",
			expected: &QueryResult{
				RowCount: 2,
				Rows: [][]interface{}{
					{"1", "Alice"},
					{"2", "Bob"},
				},
			},
			actual: &QueryResult{
				RowCount: 2,
				Rows: [][]interface{}{
					{"2", "Bob"},
					{"1", "Alice"},
				},
			},
			orderMatters: true,
			isCorrect:    false,
		},
		{
			name: "Mismatch order with orderMatters=false",
			expected: &QueryResult{
				RowCount: 2,
				Rows: [][]interface{}{
					{"1", "Alice"},
					{"2", "Bob"},
				},
			},
			actual: &QueryResult{
				RowCount: 2,
				Rows: [][]interface{}{
					{"2", "Bob"},
					{"1", "Alice"},
				},
			},
			orderMatters: false,
			isCorrect:    true,
		},
		{
			name: "Value mismatch",
			expected: &QueryResult{
				RowCount: 1,
				Rows: [][]interface{}{
					{"1", "Alice"},
				},
			},
			actual: &QueryResult{
				RowCount: 1,
				Rows: [][]interface{}{
					{"1", "Charlie"},
				},
			},
			orderMatters: false,
			isCorrect:    false,
		},
		{
			name: "Row count mismatch",
			expected: &QueryResult{
				RowCount: 2,
				Rows: [][]interface{}{
					{"1", "Alice"},
					{"2", "Bob"},
				},
			},
			actual: &QueryResult{
				RowCount: 1,
				Rows: [][]interface{}{
					{"1", "Alice"},
				},
			},
			orderMatters: false,
			isCorrect:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := r.Compare(tt.expected, tt.actual, tt.orderMatters)
			if res.IsCorrect != tt.isCorrect {
				t.Errorf("Compare() isCorrect = %v, want %v", res.IsCorrect, tt.isCorrect)
			}
		})
	}
}
