const MAX_DIMENSION = 100;

export function calculateStatistics(matrices) {
  if (!Array.isArray(matrices) || matrices.length < 1 || matrices.length > 2) {
    throw new Error('matrices must contain one or two matrices');
  }

  let min = Infinity;
  let max = -Infinity;
  let sum = 0;
  let count = 0;
  let hasDiagonalMatrix = false;

  for (const matrix of matrices) {
    if (!Array.isArray(matrix) || matrix.length < 1 || matrix.length > MAX_DIMENSION) {
      throw new Error('each matrix must have between 1 and 100 rows and columns');
    }
    if (!Array.isArray(matrix[0]) || matrix[0].length < 1 || matrix[0].length > MAX_DIMENSION) {
      throw new Error('each matrix must have between 1 and 100 rows and columns');
    }
    const width = matrix[0].length;
    let diagonal = matrix.length === width;
    let largestMagnitude = 0;
    for (const row of matrix) {
      if (!Array.isArray(row) || row.length !== width) {
        throw new Error('each matrix must be rectangular');
      }
      for (const value of row) {
        if (typeof value !== 'number' || !Number.isFinite(value)) {
          throw new Error('matrix values must be finite numbers');
        }
        largestMagnitude = Math.max(largestMagnitude, Math.abs(value));
        min = Math.min(min, value);
        max = Math.max(max, value);
        sum += value;
        count++;
      }
    }
    const tolerance = 1e-9 * Math.max(1, largestMagnitude);
    if (diagonal) {
      for (let i = 0; i < matrix.length && diagonal; i++) {
        for (let j = 0; j < width; j++) {
          if (i !== j && Math.abs(matrix[i][j]) > tolerance) {
            diagonal = false;
            break;
          }
        }
      }
    }
    hasDiagonalMatrix ||= diagonal;
  }
  const average = sum / count;
  if (!Number.isFinite(sum) || !Number.isFinite(average)) {
    throw new Error('matrix values exceed the supported numeric range');
  }
  return { min, max, sum, average, hasDiagonalMatrix };
}
