import express from 'express';
import { readFileSync } from 'node:fs';
import { calculateStatistics } from './statistics.js';

const openapiDocument = readFileSync(new URL('../openapi.yaml', import.meta.url));
const docsPage = readFileSync(new URL('../docs.html', import.meta.url));

export const app = express();
app.disable('x-powered-by');
app.use(express.json({ limit: '1mb' }));

app.get('/health', (_request, response) => response.json({ status: 'ok' }));
app.get('/openapi.yaml', (_request, response) => {
  response.type('application/yaml').send(openapiDocument);
});
app.get('/docs', (_request, response) => {
  response.type('html').send(docsPage);
});
const api = express.Router();
api.post('/statistics', (request, response) => {
  try {
    response.json(calculateStatistics(request.body?.matrices));
  } catch (error) {
    response.status(400).json({ error: error.message });
  }
});
app.use('/api/v1', api);
app.use((error, _request, response, _next) => {
  const message = error.type === 'entity.too.large'
    ? 'request body too large'
    : 'invalid JSON body';
  response.status(error.status || 400).json({ error: message });
});
