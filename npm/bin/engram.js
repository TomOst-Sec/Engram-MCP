#!/usr/bin/env node
'use strict';

const { spawn } = require('child_process');
const { getBinaryPath } = require('../lib/platform');

const binary = getBinaryPath();
const child = spawn(binary, process.argv.slice(2), { stdio: 'inherit' });

child.on('error', (err) => {
  if (err.code === 'ENOENT') {
    console.error('Engram binary not found at:', binary);
    console.error('Run: go install github.com/TomOst-Sec/colony-project/cmd/engram@latest');
    process.exit(1);
  }
  console.error('Failed to start Engram:', err.message);
  process.exit(1);
});

child.on('exit', (code) => process.exit(code || 0));
