'use strict';

const https = require('https');
const http = require('http');
const fs = require('fs');
const path = require('path');
const { getPlatform, getBinaryName } = require('./platform');

const REPO = 'TomOst-Sec/colony-project';

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const proto = url.startsWith('https') ? https : http;
    proto.get(url, (res) => {
      // Follow redirects (GitHub releases redirect to S3)
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        return download(res.headers.location, dest).then(resolve).catch(reject);
      }
      if (res.statusCode !== 200) {
        return reject(new Error(`Download failed: HTTP ${res.statusCode}`));
      }
      const file = fs.createWriteStream(dest);
      res.pipe(file);
      file.on('finish', () => file.close(resolve));
      file.on('error', reject);
    }).on('error', reject);
  });
}

async function install() {
  const platform = getPlatform();
  if (!platform) {
    console.error(`Unsupported platform: ${process.platform} ${process.arch}`);
    console.error('Build from source: go install github.com/TomOst-Sec/colony-project/cmd/engram@latest');
    process.exit(1);
  }

  const version = require('../package.json').version;
  const binaryName = getBinaryName();
  const url = `https://github.com/${REPO}/releases/download/v${version}/engram-${platform}`;
  const dest = path.join(__dirname, '..', 'bin', binaryName);

  console.log(`Downloading Engram v${version} for ${platform}...`);

  try {
    await download(url, dest);

    // Set executable permissions on unix
    if (process.platform !== 'win32') {
      fs.chmodSync(dest, 0o755);
    }

    console.log('Engram installed successfully!');
  } catch (err) {
    console.error(`Failed to download Engram binary: ${err.message}`);
    console.error('');
    console.error('Binary not yet published. Build from source:');
    console.error('  go install github.com/TomOst-Sec/colony-project/cmd/engram@latest');
    // Don't exit with error — allow npm install to succeed even without binary
  }
}

install();
