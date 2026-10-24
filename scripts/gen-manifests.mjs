import { spawnSync } from 'node:child_process'
import { existsSync } from 'node:fs'
import os from 'node:os'
import path from 'node:path'

const binName = process.platform === 'win32' ? 'controller-gen.exe' : 'controller-gen'
const gobin = process.env.GOBIN || path.join(os.homedir(), 'go', 'bin')
const localBinary = path.join(gobin, binName)

// Prefer the installed binary; otherwise fall back to `go run` so no PATH setup
// is required on a fresh machine.
const [command, prefix] = existsSync(localBinary)
  ? [localBinary, []]
  : ['go', ['run', 'sigs.k8s.io/controller-tools/cmd/controller-gen@v0.23.0']]

function run(args) {
  const result = spawnSync(command, [...prefix, ...args], { stdio: 'inherit' })
  if (result.status !== 0) {
    process.exit(result.status ?? 1)
  }
}

run(['object', 'paths=./api/kube/vela/v1'])
run(['crd', 'paths=./api/kube/vela/v1', 'output:crd:dir=deploy/k8s/crds'])
