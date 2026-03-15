'use strict';

const { spawn } = require('child_process');
const { getBinaryPath } = require('./platform');

/**
 * Run the Engram binary with the given arguments.
 * @param {string[]} args - Command-line arguments
 * @returns {Promise<number>} Exit code
 */
function run(args) {
  return new Promise((resolve, reject) => {
    const binary = getBinaryPath();
    const child = spawn(binary, args, { stdio: 'inherit' });

    child.on('error', (err) => {
      if (err.code === 'ENOENT') {
        reject(new Error(`Engram binary not found at: ${binary}`));
      } else {
        reject(err);
      }
    });

    child.on('exit', (code) => resolve(code || 0));
  });
}

module.exports = { run };
