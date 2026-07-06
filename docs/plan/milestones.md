# Milestones

| # | Milestone | Target Phase | Criteria | Status |
|:--|:--|:--|:--|:--|
| 1 | Main CI green | After Phase 1 | Go tests pass locally; GitHub CI no longer fails on invalid filter or coverage path | Complete |
| 2 | Scripted release gates | After Phase 2 | Metadata and notes scripts tested; release workflow uses scripts | Complete |
| 3 | Maintainer-ready release workflow | After Phase 3 | `RELEASE.md`, Makefile, Dependabot, changelog, and governance docs updated | Complete |
| 4 | Provenance-ready release workflow | After Phase 4 | Changelog structure and npm package contracts are tested; npm package includes README; release workflow supports Trusted Publishing/OIDC with `NPM_TOKEN` fallback | Complete |
| 5 | Public governance gate | After Phase 5 | Public README has no internal metadata; PR and issue templates cover verification and secret-handling; governance validator runs in CI and release-check | Complete |
| 6 | Security automation baseline | After Phase 6 | CodeQL and OpenSSF Scorecard run on `main`, publish security results, and are protected by governance validation | Complete |
| 7 | Pinned workflow dependencies | After Phase 7 | External workflow actions are pinned to full SHAs; CI and release-check validate that pinned-actions contract | Complete |
| 8 | Release archive provenance | After Phase 8 | GitHub Release archives get artifact attestations; governance validation requires the release attestation action and permissions | Active |
