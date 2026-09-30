package main

import (
	"math"
	"testing"
)

func TestRotateClockwise(t *testing.T) {
	got := rotateClockwise([][]float64{{1, 2, 3}, {4, 5, 6}})
	want := [][]float64{{4, 1}, {5, 2}, {6, 3}}
	for i := range want {
		for j := range want[i] {
			if got[i][j] != want[i][j] {
				t.Fatalf("rotated[%d][%d] = %v, want %v", i, j, got[i][j], want[i][j])
			}
		}
	}
}

func TestQR(t *testing.T) {
	cases := [][][]float64{
		{{1, 2}, {3, 4}, {5, 6}},
		{{1, 2, 3}, {4, 5, 6}},
		{{2, 0}, {0, 3}},
		{{0, 0}, {0, 0}},
	}
	for _, input := range cases {
		q, r, err := factorQR(input)
		if err != nil {
			t.Fatal(err)
		}
		m, n, k := len(input), len(input[0]), min(len(input), len(input[0]))
		if len(q) != m || len(q[0]) != k || len(r) != k || len(r[0]) != n {
			t.Fatalf("unexpected dimensions for %dx%d input", m, n)
		}
		for i := 0; i < m; i++ {
			for j := 0; j < n; j++ {
				var value float64
				for h := 0; h < k; h++ {
					value += q[i][h] * r[h][j]
				}
				if math.Abs(value-input[i][j]) > 1e-9 {
					t.Fatalf("Q*R[%d][%d] = %v, want %v", i, j, value, input[i][j])
				}
			}
		}
		for i := 0; i < k; i++ {
			for j := 0; j < k; j++ {
				var value float64
				for h := 0; h < m; h++ {
					value += q[h][i] * q[h][j]
				}
				want := 0.0
				if i == j {
					want = 1
				}
				if math.Abs(value-want) > 1e-9 {
					t.Fatalf("Q^T Q[%d][%d] = %v, want %v", i, j, value, want)
				}
			}
		}
	}
}
