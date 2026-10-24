#!/usr/bin/env node
import { spawnSync } from 'node:child_process'
import { mkdirSync } from 'node:fs'

const isWindows = process.platform === 'win32'
const output = isWindows ? '../bin/scaffold.exe' : '../bin/scaffold'
mkdirSync('bin', { recursive: true })

const result = spawnSync(
  'go',
  [
    '-C',
    'services',
    'build',
    '-trimpath',
    '-ldflags',
    `-s -w -X main.Version=${process.env.SCAFFOLD_VERSION || 'dev'}`,
    '-o',
    output,
    './cmd/api',
  ],
  { stdio: 'inherit', env: { ...process.env, CGO_ENABLED: '0' } },
)
process.exit(result.status ?? 1)
