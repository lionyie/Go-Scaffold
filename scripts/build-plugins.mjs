#!/usr/bin/env node
import { spawnSync } from 'node:child_process'
import { existsSync, readdirSync, readFileSync } from 'node:fs'
import path from 'node:path'

const ROOT = process.cwd()
const PLUGINS = path.join(ROOT, 'runtime-plugins')
const pnpm = process.platform === 'win32' ? 'pnpm.cmd' : 'pnpm'

if (!existsSync(PLUGINS)) {
  console.log('build-plugins: no runtime-plugins directory')
  process.exit(0)
}

const packages = readdirSync(PLUGINS, { withFileTypes: true })
  .filter((entry) => entry.isDirectory())
  .map((entry) => path.join(PLUGINS, entry.name, 'package.json'))
  .filter((file) => existsSync(file))
  .map((file) => JSON.parse(readFileSync(file, 'utf8')).name)

if (packages.length === 0) {
  console.log('build-plugins: no installed runtime plugins')
  process.exit(0)
}

for (const name of packages) {
  console.log(`build plugin ${name}`)
  const result = spawnSync(pnpm, ['--filter', name, 'build'], { stdio: 'inherit', shell: process.platform === 'win32' })
  if (result.status !== 0) process.exit(result.status ?? 1)
}
