#!/usr/bin/env node
import { spawnSync } from 'node:child_process'
import { readFileSync, readdirSync, rmSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import process from 'node:process'

const ROOT = process.cwd()
const CONFIG = path.join(ROOT, 'scaffold.json')
const SKIP_DIR_NAMES = new Set(['.git', 'node_modules', 'dist', '.turbo', '.scaffold-tmp', '.vscode', '.idea'])
const SKIP_ROOT_DIRS = new Set(['data', 'bin'])
const SKIP_FILES = new Set(['pnpm-lock.yaml', 'go.sum', 'extensions/installed.lock.json'])
const TEXT_EXTENSIONS = new Set([
  '.go', '.mod', '.proto', '.yaml', '.yml', '.json', '.md', '.ts', '.vue', '.mjs', '.js', '.css', '.html',
  '.toml', '.sh', '.env', '.txt', '.gitignore', '.dockerignore', '.editorconfig', '.nvmrc', '.golangci.yml',
  '.prettierrc.json',
])

function fail(message) {
  console.error(`rename: ${message}`)
  process.exit(1)
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, 'utf8'))
}

function parseArgs(argv) {
  const flags = { dryRun: true, force: false, skipInstall: false, skipGenerate: false }
  const values = {}
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i]
    if (arg === '--') continue
    if (arg === '--dry-run') flags.dryRun = true
    else if (arg === '--yes') flags.dryRun = false
    else if (arg === '--force') flags.force = true
    else if (arg === '--skip-install') flags.skipInstall = true
    else if (arg === '--skip-generate') flags.skipGenerate = true
    else if (arg === '--name') values.name = argv[++i]
    else if (arg === '--module') values.module = argv[++i]
    else if (arg === '--display') values.display = argv[++i]
    else if (arg.startsWith('--name=')) values.name = arg.slice(7)
    else if (arg.startsWith('--module=')) values.module = arg.slice(9)
    else if (arg.startsWith('--display=')) values.display = arg.slice(10)
    else fail(`unknown argument ${arg}`)
  }
  return { flags, values }
}

function walk(dir, relDir = '') {
  const out = []
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name)
    const rel = relDir ? `${relDir}/${entry.name}` : entry.name
    if (entry.isDirectory()) {
      if (SKIP_DIR_NAMES.has(entry.name)) continue
      if (relDir === '' && SKIP_ROOT_DIRS.has(entry.name)) continue
      out.push(...walk(full, rel))
    } else {
      out.push(full)
    }
  }
  return out
}

function isText(file) {
  const name = path.basename(file)
  if (name === 'Makefile' || name === 'Dockerfile') return true
  return TEXT_EXTENSIONS.has(path.extname(name))
}

function titleCase(name) {
  return name
    .split(/[-_.]/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}

function ensureClean(flags) {
  if (flags.force) return
  const head = spawnSync('git', ['rev-parse', '--verify', 'HEAD'], { encoding: 'utf8' })
  if (head.status !== 0) return
  const status = spawnSync('git', ['status', '--porcelain'], { encoding: 'utf8' })
  if (status.stdout.trim() !== '') fail('working tree is not clean; commit or stash first (or pass --force)')
  const lock = readJSON(path.join(ROOT, 'extensions', 'installed.lock.json'))
  if (Object.keys(lock.extensions ?? {}).length > 0) {
    fail('extensions are installed; remove them before renaming (or pass --force)')
  }
}

function applyRule(source, rule) {
  const pattern = rule.pattern ?? new RegExp(rule.from.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g')
  const matches = source.match(pattern)
  if (!matches) return { source, count: 0 }
  return { source: source.replace(pattern, rule.to), count: matches.length }
}

function main() {
  const { flags, values } = parseArgs(process.argv.slice(2))
  if (!values.name || !/^[a-z][a-z0-9-]*$/.test(values.name)) {
    fail('usage: rename.mjs --name <lowercase-name> [--module <go-module>] [--display <name>] [--yes] [--dry-run] [--force]')
  }
  const config = readJSON(CONFIG)
  const name = values.name
  const module =
    values.module ??
    (config.module.endsWith('/services')
      ? config.module.replace(/\/[^/]+\/services$/, `/${name}/services`)
      : `github.com/example/${name}`)
  const displayName = values.display ?? titleCase(name)
  const envPrefix = name.toUpperCase().replace(/[^A-Z0-9]/g, '_')
  const npmScope = `@${name}`
  const namespace = `${name}-system`

  ensureClean(flags)

  const files = walk(ROOT).filter((file) => {
    const rel = path.relative(ROOT, file).split(path.sep).join('/')
    return !SKIP_FILES.has(rel) && isText(file)
  })

  const rules = [
    {
      type: 'go-module',
      from: config.module,
      to: module,
      include: (rel) =>
        (rel.startsWith('services/') && (rel.endsWith('.go') || rel.endsWith('go.mod'))) ||
        (rel.startsWith('contracts/') && rel.endsWith('.proto')) ||
        rel.startsWith('buf.gen.') ||
        (rel.startsWith('extensions/templates/') && (rel.endsWith('.go') || rel.endsWith('.proto'))),
    },
    {
      type: 'npm-scope',
      from: `${config.npmScope}/`,
      to: `${npmScope}/`,
      include: (rel) =>
        rel.endsWith('package.json') ||
        rel.startsWith('apps/') ||
        rel.startsWith('packages/') ||
        rel.startsWith('scripts/') ||
        (rel.startsWith('extensions/templates/') && rel.endsWith('.json')),
    },
    {
      type: 'env-prefix',
      from: `${config.envPrefix}_`,
      to: `${envPrefix}_`,
      include: () => true,
    },
    {
      type: 'env-source',
      pattern: new RegExp(`"${config.envPrefix}"`, 'g'),
      to: `"${envPrefix}"`,
      include: (rel) => rel.endsWith('.go'),
    },
    {
      type: 'binary',
      pattern: new RegExp(`(?<!@)(?<![\\w-])${config.binary}(?![\\w.\\-])`, 'g'),
      to: name,
      include: (rel) =>
        rel === 'package.json' ||
        rel === 'Makefile' ||
        rel.startsWith('scripts/') ||
        rel.startsWith('services/cmd/') ||
        rel.startsWith('deploy/') ||
        rel.startsWith('configs/') ||
        rel.startsWith('docs/') ||
        rel.startsWith('README'),
    },
    {
      type: 'namespace',
      from: config.namespace,
      to: namespace,
      include: (rel) => rel.startsWith('deploy/') || rel.startsWith('configs/') || rel.startsWith('docs/') || rel.startsWith('README'),
    },
    {
      type: 'db-name',
      from: config.dbName,
      to: name,
      include: (rel) => rel.startsWith('configs/') || rel.startsWith('scripts/') || rel.startsWith('deploy/') || rel.startsWith('docs/'),
    },
    {
      type: 'docker-image',
      from: config.dockerImage,
      to: name,
      include: (rel) => rel.startsWith('deploy/') || rel.startsWith('.github/') || rel === 'Makefile' || rel.startsWith('README'),
    },
    {
      type: 'display-name',
      from: config.displayName,
      to: displayName,
      include: (rel) => rel.startsWith('README') || rel.startsWith('docs/') || rel === 'package.json',
    },
  ]

  const changes = []
  const updates = new Map()
  const protectedToken = '\u0000SCAFFOLD_JSON\u0000'
  for (const file of files) {
    const rel = path.relative(ROOT, file).split(path.sep).join('/')
    if (rel === 'scaffold.json') continue
    let source = readFileSync(file, 'utf8')
    // The config file keeps its name even after the project is renamed.
    source = source.split('scaffold.json').join(protectedToken)
    let original = source
    for (const rule of rules) {
      if (!rule.include(rel)) continue
      const result = applyRule(source, rule)
      if (result.count > 0) {
        source = result.source
        changes.push({ file: rel, type: rule.type, count: result.count })
      }
    }
    source = source.split(protectedToken).join('scaffold.json')
    if (source !== original) updates.set(file, source)
  }

  for (const change of changes) {
    console.log(`${change.file}: ${change.type} x${change.count}`)
  }
  console.log(`\n${updates.size} file(s) would change`)
  if (flags.dryRun) {
    console.log('dry run; pass --yes to apply')
    return
  }

  for (const [file, source] of updates) writeFileSync(file, source)

  const nextConfig = {
    ...config,
    name,
    module,
    envPrefix,
    npmScope,
    binary: name,
    dbName: name,
    dockerImage: name,
    namespace,
    displayName,
    description: `${displayName} — built with the scaffold.`,
  }
  writeFileSync(CONFIG, `${JSON.stringify(nextConfig, null, 2)}\n`)

  if (!flags.skipInstall) {
    rmSync(path.join(ROOT, 'pnpm-lock.yaml'), { force: true })
    const pnpm = process.platform === 'win32' ? 'pnpm.cmd' : 'pnpm'
    const install = spawnSync(pnpm, ['install'], { cwd: ROOT, stdio: 'inherit', shell: process.platform === 'win32' })
    if (install.status !== 0) fail('pnpm install failed')
  }
  if (!flags.skipGenerate) {
    const generate = spawnSync(process.execPath, ['scripts/generate.mjs'], { cwd: ROOT, stdio: 'inherit' })
    if (generate.status !== 0) fail('code generation failed')
    const tidy = spawnSync('go', ['-C', 'services', 'mod', 'tidy'], { cwd: ROOT, stdio: 'inherit' })
    if (tidy.status !== 0) fail('go mod tidy failed')
  }
  console.log(`renamed to ${name} (${module})`)
}

main()
