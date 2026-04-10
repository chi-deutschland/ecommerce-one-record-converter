# Contributing to eCommerce ONE Record Converter

Thank you for your interest in contributing! This project is part of the
[DTAC](https://www.digitales-testfeld-air-cargo.de/) research initiative and is maintained with limited resources, so we
appreciate every contribution.

## How to Report Bugs

Open a [GitHub Issue](../../issues/new?template=bug_report.md) with:

- A clear description of the problem
- Steps to reproduce it
- What you expected to happen vs. what actually happened
- Your environment (Go version, OS, NE:ONE Server version)

## How to Suggest Features

Open a [GitHub Issue](../../issues/new) with the label `enhancement` describing your idea and
the use case it would address.

## How to Submit Changes

1. **Fork** the repository
2. Create a **feature branch** from `main` (e.g. `feature/my-improvement`)
3. Make your changes
4. Run the tests and linter before committing:
   ```bash
   # Backend
   go test ./...

   # Frontend
   cd cmd/frontend
   npm run lint
   ```
5. Open a **Pull Request** against `main` with a clear description of your changes

## Code Style

- **Go**: Follow standard Go conventions. The project uses
  [zerolog](https://github.com/rs/zerolog) for logging and
  [failsafe-go](https://github.com/failsafe-go/failsafe-go) for resilience policies.
- **Frontend**: Follow the existing Next.js / Tailwind patterns in `cmd/frontend`.

## Questions?

If you have questions about the project or how to contribute, feel free to open an issue — we're
happy to help.

