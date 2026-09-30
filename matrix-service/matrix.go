package main

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/blas/blas64"
	"gonum.org/v1/gonum/lapack/lapack64"
)

const maxDimension = 100

func validateMatrix(matrix [][]float64) error {
	if len(matrix) == 0 || len(matrix) > maxDimension {
		return fmt.Errorf("matrix must have between 1 and %d rows", maxDimension)
	}
	width := len(matrix[0])
	if width == 0 || width > maxDimension {
		return fmt.Errorf("matrix must have between 1 and %d columns", maxDimension)
	}
	for _, row := range matrix {
		if len(row) != width {
			return fmt.Errorf("matrix must be rectangular")
		}
		for _, value := range row {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return fmt.Errorf("matrix values must be finite")
			}
		}
	}
	return nil
}

func rotateClockwise(matrix [][]float64) [][]float64 {
	rows, cols := len(matrix), len(matrix[0])
	rotated := make([][]float64, cols)
	for col := range rotated {
		rotated[col] = make([]float64, rows)
		for row := 0; row < rows; row++ {
			rotated[col][row] = matrix[rows-1-row][col]
		}
	}
	return rotated
}

// factorQR returns the reduced QR factorization of any nonempty m×n matrix.
// Q is m×k and R is k×n, where k=min(m,n).
func factorQR(matrix [][]float64) ([][]float64, [][]float64, error) {
	m, n := len(matrix), len(matrix[0])
	k := min(m, n)
	matrixData := make([]float64, m*n)
	for i, row := range matrix {
		copy(matrixData[i*n:(i+1)*n], row)
	}
	tau := make([]float64, k)
	work := make([]float64, max(m, n))
	// Geqrf stores R above the diagonal and Q's reflectors below it.
	lapack64.Geqrf(blas64.General{Rows: m, Cols: n, Stride: n, Data: matrixData}, tau, work, len(work))

	r := make([][]float64, k)
	for i := range r {
		r[i] = make([]float64, n)
		for j := i; j < n; j++ {
			r[i][j] = matrixData[i*n+j]
		}
	}

	qData := make([]float64, m*k)
	for i := 0; i < m; i++ {
		copy(qData[i*k:(i+1)*k], matrixData[i*n:i*n+k])
	}
	// Orgqr expands the reflectors into the reduced Q matrix.
	lapack64.Orgqr(blas64.General{Rows: m, Cols: k, Stride: k, Data: qData}, tau, work, len(work))
	q := make([][]float64, m)
	for i := range q {
		q[i] = qData[i*k : (i+1)*k]
	}
	for _, rows := range [][][]float64{q, r} {
		for _, row := range rows {
			for _, value := range row {
				if math.IsNaN(value) || math.IsInf(value, 0) {
					return nil, nil, fmt.Errorf("matrix values exceed the supported numeric range")
				}
			}
		}
	}
	return q, r, nil
}
