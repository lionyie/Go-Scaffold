#!/usr/bin/env node
import { spawn } from 'node:child_process'

const child = spawn('go', ['-C', 'services', 'run', './cmd/api', '-conf', '../configs/config.local.yaml'], {
  stdio: 'inherit',
  env: process.env,
})
child.on('exit', (code) => process.exit(code ?? 1))
