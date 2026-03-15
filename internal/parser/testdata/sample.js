import { readFile } from 'fs/promises';
import path from 'path';

const axios = require('axios');

/**
 * Represents a database connection.
 */
class Database {
  constructor(url) {
    this.url = url;
    this.connected = false;
  }

  /** Connect to the database. */
  async connect() {
    this.connected = true;
    return this;
  }

  async query(sql, params) {
    return [];
  }
}

class CachedDatabase extends Database {
  constructor(url, cacheSize) {
    super(url);
    this.cacheSize = cacheSize;
  }
}

/**
 * Process incoming data and return results.
 * @param {Object} data - The input data
 * @returns {Object} Processed results
 */
function processData(data) {
  return { processed: true, data };
}

/** Format a greeting message */
const greet = (name) => {
  return `Hello, ${name}!`;
};

function helperFunc() {
  // internal
}

module.exports = { Database, processData };

export { greet };

describe('Database', () => {
  it('should connect', async () => {
    const db = new Database('localhost');
    await db.connect();
    expect(db.connected).toBe(true);
  });

  test('processData works', () => {
    expect(processData({ a: 1 }).processed).toBe(true);
  });
});
