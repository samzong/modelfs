#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

ensure_envtest_assets() {
  local version="${ENVTEST_K8S_VERSION:-1.28.0}"
  local assets_dir_rel="${ENVTEST_ASSETS_DIR:-.cache/envtest}"
  mkdir -p "${assets_dir_rel}"
  local assets_dir
  assets_dir="$(cd "${assets_dir_rel}" && pwd)"

  if [ -x "${assets_dir}/etcd" ] && [ -x "${assets_dir}/kube-apiserver" ] && [ -x "${assets_dir}/kubectl" ]; then
    export KUBEBUILDER_ASSETS="${assets_dir}"
    return
  fi

  echo "==> Installing envtest assets (version ${version})"
  mkdir -p "${assets_dir}"
  local setup_bin
  if command -v setup-envtest >/dev/null 2>&1; then
    setup_bin="$(command -v setup-envtest)"
  else
    echo "==> setup-envtest binary not found; installing"
    local gobin
    gobin="$(go env GOPATH)/bin"
    GOBIN="${gobin}" go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
    setup_bin="${gobin}/setup-envtest"
  fi
  "${setup_bin}" use "${version}" --bin-dir "${assets_dir}"
  export KUBEBUILDER_ASSETS="${assets_dir}"
}

run_go_pipeline() {
  echo "==> Running gofmt via make tidy"
  make tidy

  ensure_envtest_assets

  echo "==> Running Go unit tests"
  go test ./...

  echo "==> Running Helm lint"
  make helm-lint
}

run_ui_pipeline() {
  if [ -d "${ROOT_DIR}/ui" ]; then
    if ! command -v pnpm >/dev/null 2>&1; then
      echo "==> WARNING: ui/ detected but pnpm is not installed; skipping frontend lint/test"
      return
    fi
    echo "==> Running frontend checks"
    pushd ui >/dev/null
    if [ -f "pnpm-lock.yaml" ]; then
      pnpm install --frozen-lockfile
    else
      pnpm install
    fi
    pnpm lint
    pnpm test
    popd >/dev/null
  else
    echo "==> Skipping frontend checks (ui/ directory not present)"
  fi
}

SCOPE="${TEST_SCOPE:-all}"
case "${SCOPE}" in
  go)
    run_go_pipeline
    ;;
  ui)
    run_ui_pipeline
    ;;
  all)
    run_go_pipeline
    run_ui_pipeline
    ;;
  *)
    echo "Unknown TEST_SCOPE=${SCOPE}. Expected one of: go, ui, all" >&2
    exit 1
    ;;
esac

echo "==> scripts/test-all.sh completed (scope=${SCOPE})"

