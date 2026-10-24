#!/usr/bin/env node
import { existsSync, readdirSync, rmSync, writeFileSync } from 'node:fs'
import path from 'node:path'

const ROOT = process.cwd()

function remove(target) {
  rmSync(path.join(ROOT, target), { recursive: true, force: true })
}

function cleanKeeping(directory, keep) {
  const full = path.join(ROOT, directory)
  if (!existsSync(full)) return
  for (const entry of readdirSync(full, { withFileTypes: true })) {
    if (entry.name === keep) continue
    rmSync(path.join(full, entry.name), { recursive: true, force: true })
  }
}

remove('bin')
remove('.scaffold-tmp')
remove('.turbo')
remove('apps/web/dist')
remove('apps/admin/dist')
remove('services/internal/webui/dist')
for (const app of ['web', 'admin']) {
  const dir = path.join('services', 'internal', 'webui', 'dist', app)
  remove(dir)
}
if (!existsSync(path.join(ROOT, 'services', 'internal', 'webui', 'dist'))) {
  const dir = path.join(ROOT, 'services', 'internal', 'webui', 'dist')
  rmSync(dir, { recursive: true, force: true })
  // go:embed requires the directory to exist even when empty.
  const { mkdirSync } = await import('node:fs')
  mkdirSync(dir, { recursive: true })
  writeFileSync(path.join(dir, '.gitkeep'), '')
}

try {
  for (const plugin of readdirSync(path.join(ROOT, 'runtime-plugins'), { withFileTypes: true })) {
    if (!plugin.isDirectory()) continue
    cleanKeeping(path.join('runtime-plugins', plugin.name, 'dist'), '.gitkeep')
  }
} catch {
  // no runtime-plugins directory
}
