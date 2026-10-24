#!/usr/bin/env node
import { cpSync, existsSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import path from 'node:path'

const ROOT = process.cwd()
const TARGET = path.join(ROOT, 'services', 'internal', 'webui', 'dist')

rmSync(TARGET, { recursive: true, force: true })
mkdirSync(TARGET, { recursive: true })
writeFileSync(path.join(TARGET, '.gitkeep'), '')

for (const app of ['web', 'admin']) {
  const appDir = path.join(ROOT, 'apps', app)
  const manifestPath = path.join(appDir, 'app.json')
  if (!existsSync(manifestPath)) continue
  const manifest = JSON.parse(readFileSync(manifestPath, 'utf8'))
  if (!manifest.embed) continue
  const source = path.join(appDir, 'dist')
  if (!existsSync(source)) {
    console.error(`copy-web-dist: ${app} has no dist; run build:${app} first`)
    process.exit(1)
  }
  const destination = path.join(TARGET, app)
  cpSync(source, destination, { recursive: true })
  cpSync(manifestPath, path.join(destination, 'app.json'))
  console.log(`copied ${path.relative(ROOT, source)} -> ${path.relative(ROOT, destination)}`)
}
