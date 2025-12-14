#!/usr/bin/env node

const os = require('os');
const path = require('path');
const fs = require('fs');

// Platform mapping from Node.js to package names
const PLATFORM_MAP = {
  'darwin-arm64': '@claudex/darwin-arm64',
  'darwin-x64': '@claudex/darwin-x64',
  'linux-x64': '@claudex/linux-x64',
  'linux-arm64': '@claudex/linux-arm64'
};

// Binary names to link
const BINARIES = ['claudex', 'claudex-hooks'];

function getPlatformKey() {
  const platform = os.platform();
  const arch = os.arch();

  // Normalize architecture names
  let normalizedArch = arch;
  if (arch === 'x64' || arch === 'amd64') {
    normalizedArch = 'x64';
  } else if (arch === 'arm64' || arch === 'aarch64') {
    normalizedArch = 'arm64';
  }

  return `${platform}-${normalizedArch}`;
}

function findPlatformPackage(platformKey) {
  const packageName = PLATFORM_MAP[platformKey];

  if (!packageName) {
    return null;
  }

  // Try to find the platform package in node_modules
  const packagePath = path.join(__dirname, 'node_modules', packageName);

  if (fs.existsSync(packagePath)) {
    return packagePath;
  }

  // Also check parent node_modules (for global installs)
  const parentPackagePath = path.join(__dirname, '..', packageName);

  if (fs.existsSync(parentPackagePath)) {
    return parentPackagePath;
  }

  return null;
}

function createSymlink(target, linkPath) {
  // Remove existing symlink/file if it exists
  if (fs.existsSync(linkPath)) {
    fs.unlinkSync(linkPath);
  }

  // Create symlink (relative path for portability)
  const relativePath = path.relative(path.dirname(linkPath), target);
  fs.symlinkSync(relativePath, linkPath);

  // Make it executable
  try {
    fs.chmodSync(linkPath, 0o755);
  } catch (err) {
    // Ignore chmod errors on Windows
  }
}

function main() {
  const platformKey = getPlatformKey();

  console.log(`Installing claudex for ${platformKey}...`);

  const platformPackage = findPlatformPackage(platformKey);

  if (!platformPackage) {
    const supportedPlatforms = Object.keys(PLATFORM_MAP).join(', ');
    console.error(`
Error: Platform ${platformKey} is not supported.

Supported platforms: ${supportedPlatforms}

If you believe this platform should be supported, please file an issue at:
https://github.com/maikel/claudex/issues
`);
    process.exit(1);
  }

  console.log(`Found platform package at: ${platformPackage}`);

  // Create symlinks for each binary
  const binDir = path.join(__dirname, 'bin');

  // Ensure bin directory exists
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  let hasErrors = false;

  BINARIES.forEach(binaryName => {
    const sourceBinary = path.join(platformPackage, 'bin', binaryName);
    const targetLink = path.join(binDir, binaryName);

    if (!fs.existsSync(sourceBinary)) {
      console.error(`Warning: Binary ${binaryName} not found in platform package`);
      hasErrors = true;
      return;
    }

    try {
      createSymlink(sourceBinary, targetLink);
      console.log(`Linked ${binaryName}`);
    } catch (err) {
      console.error(`Error linking ${binaryName}: ${err.message}`);
      hasErrors = true;
    }
  });

  if (hasErrors) {
    console.error('\nInstallation completed with warnings. Some binaries may not be available.');
    process.exit(1);
  }

  console.log('\n✓ claudex installed successfully!\n\nTry running: claudex --help\n');
}

// Run postinstall
try {
  main();
} catch (err) {
  console.error(`Installation failed: ${err.message}`);
  process.exit(1);
}
