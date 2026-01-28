#!/usr/bin/env node

import { build } from 'vite';

try {
  await build({
    configFile: './vite.config.ts',
    mode: 'production'
  });
  console.log('✅ Build successful!');
} catch (error) {
  console.error('❌ Build failed:', error);
  process.exit(1);
}
