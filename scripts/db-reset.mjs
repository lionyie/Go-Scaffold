#!/usr/bin/env node
import { existsSync, readdirSync, rmSync } from 'node:fs'
import path from 'node:path'
import { createInterface } from 'node:readline/promises'
import process from 'node:process'

const ROOT = process.cwd()
const DATA = path.join(ROOT, 'data')
const yes = process.argv.includes('--yes')

if (!yes) {
  if (!process.stdin.isTTY) {
    console.error('db:reset: refusing to delete data without --yes in a non-interactive shell')
    process.exit(1)
  }
  const rl = createInterface({ input: process.stdin, output: process.stdout })
  const answer = await rl.question('Delete all local database files in data/? [y/N] ')
  rl.close()
  if (answer.trim().toLowerCase() !== 'y') process.exit(0)
}

if (!existsSync(DATA)) {
  console.log('db:reset: no data directory')
  process.exit(0)
}

let removed = 0
for (const entry of readdirSync(DATA, { withFileTypes: true })) {
  if (!entry.isFile()) continue
  if (!/\.(db|db-shm|db-wal|sqlite)$/.test(entry.name)) continue
  rmSync(path.join(DATA, entry.name), { force: true })
  removed += 1
}
console.log(`db:reset: removed ${removed} file(s)`)
