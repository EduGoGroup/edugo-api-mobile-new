# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0] - 2026-02-24


## [0.13.0] - 2026-03-16

### Tipo de Release: patch

- fix: use RowsAffected instead of uuid.Nil for not-found detection in GetByIDAndSchool
- Use raw SQL for guardian queries
- update dependencia
- feat: enable Node 24 for GitHub Actions in Azure deployment workflow

---

## [0.12.0] - 2026-03-11

### Tipo de Release: patch

- Update dependencies in go.sum to latest versions
- feat: UI review improvements — Makefile updates and documentation
- update (#24)
- feat: assessment question start, answer, and submit endpoints (#23)
- chore: upgrade shared common/v0.54.0 + repository/v0.4.0, regenerate swagger
- fix: resolve Copilot review comments on PR #22
- feat: assessment CRUD - N:N materials, NUMERIC fields, full management
- fix: guardian route permissions (#20)
- feat(mobile-api): AuditMiddleware + explicit audit logging (#18)

---

## [0.11.0] - 2026-03-07

### Tipo de Release: patch

- feat: assessment CRUD with N:N materials and SDUI field system (#22)

---

## [0.10.0] - 2026-03-06

### Tipo de Release: patch

- update (#21)
- feat(mobile-api): AuditMiddleware + explicit audit logging (#18) (#19)
- skill and swagger

---

## [0.9.0] - 2026-03-04

### Tipo de Release: patch

- chore(deps): bump shared/common to v0.52.0
- fix(pagination): address code review comments from PR #16
- feat(pagination): implement real pagination with COUNT for mobile endpoints
- perf(docker): eliminate Go compilation from Docker, reduce image time ~80%

---

## [0.8.0] - 2026-03-03

### Tipo de Release: patch

- chore(deps): bump edugo-shared/auth to v0.52.0 and repository to v0.3.2
- chore(deps): bump edugo-infrastructure/postgres to v0.58.0

---

## [0.7.0] - 2026-03-02

### Tipo de Release: patch

- fix: add missing 401 Unauthorized responses to assessment endpoints
- Add assessment management and question endpoints to API docs

---

## [0.6.0] - 2026-03-02

### Tipo de Release: patch

- chore: bump edugo-infrastructure/postgres to v0.57.0
- fix: address 8 Copilot review comments
- feat: Phase 3 - Assessment CRUD + Questions API for teachers

---

## [0.5.0] - 2026-03-02

### Tipo de Release: patch

- fix: address Copilot review feedback
- chore: bump edugo-infrastructure/postgres v0.56.0 and edugo-shared/common v0.51.0
- feat: add guardian auto-registration endpoints

---

## [0.4.0] - 2026-02-26

### Tipo de Release: patch

- chore: upgrade edugo-shared/auth to v0.51.0
- feat: Sprint 8 — add version and updatedAt to screen config response

---

## [0.3.0] - 2026-02-26

### Tipo de Release: patch

- Add search and search_fields filters to materials and attempts APIs

---

## [0.2.0] - 2026-02-25

### Tipo de Release: patch

- Update dependencies and remove unused Actions field from screen DTOs
- fix: use GITHUB_TOKEN instead of GHCR_TOKEN for registry auth

---

## [0.1.1] - 2026-02-24

### Tipo de Release: patch

- fix: correct target port 9091 to 8080 in deploy pipeline

---

## [0.1.0] - 2026-02-24

### Tipo de Release: patch

- fix
- chore: release v0.1.0
- Update Azure deploy workflow to use secrets for env vars
- Update screen service with Redis caching and IAM client integration
- chore: Elimina archivos de configuración y actualiza las referencias de DTO y la estructura de `SavePreferencesRequest` en la documentación de Swagger.
- Add Swagger docs and migrate to GORM for Postgres
- feat: initial commit - clean API Mobile rebuild

---
### Tipo de Release: patch

- Update Azure deploy workflow to use secrets for env vars
- Update screen service with Redis caching and IAM client integration
- chore: Elimina archivos de configuración y actualiza las referencias de DTO y la estructura de `SavePreferencesRequest` en la documentación de Swagger.
- Add Swagger docs and migrate to GORM for Postgres
- feat: initial commit - clean API Mobile rebuild

---
