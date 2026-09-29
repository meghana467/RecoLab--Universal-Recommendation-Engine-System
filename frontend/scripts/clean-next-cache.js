const fs = require('fs');
const path = require('path');

const cacheDir = path.join(__dirname, '..', '.next');

try {
  fs.rmSync(cacheDir, { recursive: true, force: true, maxRetries: 3, retryDelay: 100 });
} catch (error) {
  console.warn(`Could not clean .next cache: ${error.message}`);
}
