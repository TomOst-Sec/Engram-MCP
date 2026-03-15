'use strict';

const path = require('path');

const PLATFORM_MAP = {
  'darwin-x64': 'darwin-amd64',
  'darwin-arm64': 'darwin-arm64',
  'linux-x64': 'linux-amd64',
  'linux-arm64': 'linux-arm64',
  'win32-x64': 'windows-amd64',
};

function getPlatform() {
  const key = `${process.platform}-${process.arch}`;
  return PLATFORM_MAP[key] || null;
}

function getBinaryName() {
  return process.platform === 'win32' ? 'engram.exe' : 'engram';
}

function getBinaryPath() {
  return path.join(__dirname, '..', 'bin', getBinaryName());
}

module.exports = { getPlatform, getBinaryName, getBinaryPath, PLATFORM_MAP };
