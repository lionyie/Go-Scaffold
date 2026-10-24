#!/usr/bin/env node
import { spawnSync } from 'node:child_process'
import process from 'node:process'

const head = spawnSync('git', ['rev-parse', '--verify', 'HEAD'], { encoding: 'utf8' })
if (head.status !== 0) {
  console.log('verify-generated: repository has no commits yet; skipping diff check')
  process.exit(0)
}

const diff = spawnSync('git', ['diff', '--exit-code'], { stdio: 'inherit' })
const cached = spawnSync('git', ['diff', '--cached', '--exit-code'], { stdio: 'inherit' })
if (diff.status !== 0 || cached.status !== 0) {
  console.error('verify-generated: generated files are not committed')
  process.exit(1)
}
console.log('verify-generated: clean')
