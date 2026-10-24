#!/usr/bin/env node
import { spawnSync } from 'node:child_process'

const pnpm = process.platform === 'win32' ? 'pnpm.cmd' : 'pnpm'

function run(args) {
  console.log(`> pnpm ${args.join(' ')}`)
  const result = spawnSync(pnpm, args, { stdio: 'inherit', shell: process.platform === 'win32' })
  if (result.status !== 0) process.exit(result.status ?? 1)
}

run(['run', 'build:web'])
run(['run', 'build:admin'])
run(['run', 'build:plugins'])
run(['exec', 'node', 'scripts/copy-web-dist.mjs'])
run(['run', 'build:server'])
