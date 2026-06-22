# Branch Strategy

This document describes the branching strategy for the Atheon project.

## Main Branches

### `main`
The production branch. Contains the latest stable release. All changes eventually flow into this branch through pull requests.

### `stable/clean`
The integration branch for tested, stable changes. Feature branches are merged here before being promoted to `main`. This branch serves as a staging area for releases.

### `dev/full-feature`
The development branch for integrating major features. Contains work-in-progress features and experimental changes.

## Workflow

1. **Feature Development**: Create feature branches from `stable/clean`
2. **Integration**: Merge completed features to `stable/clean`
3. **Release**: Promote from `stable/clean` to `main` when ready

## Branch Protection

- `main`: Requires PRs, blocks force pushes
- `stable/clean`: Requires PRs, blocks force pushes
- `dev/full-feature`: Requires PRs for main integrations

## Configuration Profiles

Configuration profiles are stored in `config/profiles/` and manage environment-specific settings.
