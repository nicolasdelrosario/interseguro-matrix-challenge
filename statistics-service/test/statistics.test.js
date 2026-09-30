import test from 'node:test';
import assert from 'node:assert/strict';
import { once } from 'node:events';
import { app } from '../src/app.js';
import { calculateStatistics } from '../src/statistics.js';

test('calculates statistics across Q and R', () => {
  assert.deepEqual(calculateStatistics([[[1, 0], [0, 1]], [[2, 3], [0, 4]]]), {
    min: 0,
    max: 4,
    sum: 11,
    average: 1.375,
    hasDiagonalMatrix: true,
  });
});

test('only square matrices can be diagonal', () => {
  assert.equal(calculateStatistics([[[1, 0, 0], [0, 1, 0]]]).hasDiagonalMatrix, false);
  assert.equal(calculateStatistics([[[1, 1e-10], [0, 2]]]).hasDiagonalMatrix, true);
  assert.equal(calculateStatistics([[[1, 0.01], [0, 2]]]).hasDiagonalMatrix, false);
});

test('serves statistics at the versioned route', async () => {
  const server = app.listen(0);
  try {
    await once(server, 'listening');
    const response = await fetch(`http://127.0.0.1:${server.address().port}/api/v1/statistics`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ matrices: [[[1, 2], [3, 4]]] }),
    });
    assert.equal(response.status, 200);
    assert.deepEqual(await response.json(), {
      min: 1,
      max: 4,
      sum: 10,
      average: 2.5,
      hasDiagonalMatrix: false,
    });
  } finally {
    await new Promise(resolve => server.close(resolve));
  }
});

test('serves Scalar documentation', async () => {
  const server = app.listen(0);
  try {
    await once(server, 'listening');
    const response = await fetch(`http://127.0.0.1:${server.address().port}/docs`);
    assert.equal(response.status, 200);
    assert.match(response.headers.get('content-type'), /^text\/html/);
  } finally {
    await new Promise(resolve => server.close(resolve));
  }
});
