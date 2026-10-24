#!/usr/bin/env node
import { spawnSync } from 'node:child_process'
import { existsSync, readdirSync, rmSync, statSync } from 'node:fs'
import path from 'node:path'
import process from 'node:process'

import { codegenDirs } from './generated-dirs.mjs'

const ROOT = process.cwd()
const isWindows = process.platform === 'win32'

function run(command, args, options = {}) {
  console.log(`> ${command} ${args.join(' ')}`)
  const result = spawnSync(command, args, {
    cwd: ROOT,
    stdio: 'inherit',
    shell: isWindows && command.endsWith('.cmd'),
    ...options,
  })
  if (result.status !== 0) process.exit(result.status ?? 1)
}

function keepSet(entry) {
  return new Set((entry.keep ?? []).map((item) => path.resolve(ROOT, item)))
}

function removeTree(target, keep) {
  const full = path.resolve(ROOT, target)
  if (!existsSync(full)) return
  const info = statSync(full)
  if (!info.isDirectory()) {
    if (!keep.has(full)) rmSync(full, { force: true })
    return
  }
  for (const entry of readdirSync(full, { withFileTypes: true })) {
    const child = path.join(full, entry.name)
    if (keep.has(child)) continue
    if (entry.isDirectory()) removeTree(child, keep)
    else rmSync(child, { force: true })
  }
}

for (const entry of codegenDirs) {
  removeTree(entry.path, keepSet(entry))
}

run('go', ['-C', 'services', 'run', './cmd/genpermissions'])
run('buf', ['generate', '--template', 'buf.gen.backend.yaml'])
run('buf', ['generate', '--template', 'buf.gen.conf.yaml'])
run('buf', ['generate', '--template', 'buf.gen.frontend.yaml'])
run(isWindows ? 'pnpm.cmd' : 'pnpm', ['--filter', '@scaffold/api-client', 'gen'])
run('go', ['-C', 'services', 'run', 'github.com/google/wire/cmd/wire@v0.7.0', './cmd/api'])
run('go', ['-C', 'services', 'mod', 'tidy'])
run(process.execPath, ['scripts/verify-authz.mjs'])
