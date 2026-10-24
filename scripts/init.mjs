#!/usr/bin/env node
import { spawnSync } from 'node:child_process'
import { toolVersions } from './tool-versions.mjs'

const go = process.platform === 'win32' ? 'go.exe' : 'go'
const pnpm = process.platform === 'win32' ? 'pnpm.cmd' : 'pnpm'

function run(command, args) {
  const result = spawnSync(command, args, {
    stdio: 'inherit',
    shell: process.platform === 'win32' && command.endsWith('.cmd'),
  })
  if (result.status !== 0) process.exit(result.status ?? 1)
}

for (const version of Object.values(toolVersions)) {
  console.log(`install ${version}`)
  run(go, ['install', version])
}

console.log('install frontend dependencies')
run(pnpm, ['install'])
