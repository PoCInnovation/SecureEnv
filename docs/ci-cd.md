# CI/CD

## Branches

Feature branches open pull requests into `dev`. `dev` is promoted to `main` through a pull request, and releases are cut from `main`. Dependabot also targets `dev`.

## Pull request checks (`ci.yml`)

| Job | Blocks on |
|---|---|
| Lint | `golangci-lint` findings, untidy `go.mod`/`go.sum` |
| Test | failing tests (race detector on); coverage is shown in the job summary |
| Integration (Vault) | store contract failing against a real Vault 2.1 server |
| Vulnerabilities | vulnerabilities **reachable** from the code (`govulncheck`) |
| Secrets | any secret in the git history (`gitleaks`, config in `.gitleaks.toml`) |
| Build binaries | GoReleaser snapshot for every platform; archives are downloadable from the run |
| Docker image | `linux/amd64` + `linux/arm64` build, and fixable HIGH/CRITICAL CVEs in the image (Trivy) |
| **CI result** | any job above not succeeding |

Require only **CI result** (and **Analyze (Go)** from `codeql.yml`) in the branch protection of `dev` and `main`. New jobs are then enforced without touching the settings.

`codeql.yml` runs CodeQL `security-extended` queries on pull requests and every Monday.

Pull requests test the merge result, so pushes are only built on `main`. Concurrent runs for the same branch are cancelled.

### Supply chain

- Actions are pinned to commit SHAs with the version in a comment, and base images are pinned by digest. Dependabot updates both weekly.
- The gitleaks binary is downloaded with a checksum check, since `gitleaks-action` requires a paid license for organisations.
- Jobs run with read-only permissions by default and checkouts do not persist credentials.

### Known historical secrets

`.gitleaksignore` lists Vault tokens committed in 2023, before the rewrite. They stay in the public history, so they **must be revoked** on the Vault server they belonged to. Only add an entry after revoking the secret.

## Releasing

1. `release-drafter.yml` keeps a draft release up to date from merged pull requests. The version bump comes from the `major`, `minor` or `patch` label (patch by default).
2. Review the draft and publish it. This creates the `vX.Y.Z` tag.
3. `release.yml` runs on the tag:
   - reruns the tests and `govulncheck`;
   - GoReleaser attaches the archives and `checksums.txt` to the release:
     - `secureenv`: linux, darwin, windows × amd64, arm64
     - `secureenv-api`: linux, darwin × amd64, arm64
   - pushes `ghcr.io/pocinnovation/secureenv-api` for `linux/amd64` and `linux/arm64`, tagged `X.Y.Z`, `X.Y` and `sha-<commit>`, with an SBOM;
   - adds GitHub build provenance attestations to the archives and the image.

Try the release build locally with `make release-snapshot`.
