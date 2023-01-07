# golang-starter
Golang starter template

## features of this starter template
- follows common Golang best practices in terms of repo/project layout, and includes explanations of what goes where in README files
- Cobra library for CLI handling, and Viper library for reading config files already plugged in and ready to expand upon
- Goreleaser to build Docker images and most standard package types across Linux, MacOS and Windows
    - also includes auto-generated manpages and shell autocompletions
- Makefile for easy building, deploying, testing, updating, etc. both Dockerized and using locally installed Golang toolchain
- docker-compose project for easily hosting built Dockerized Golang project, with optional support for Golang web services
- scripts to make using the starter template easy, and to update the Golang version when a new one comes out
- built-in security scans, vulnerability warnings and auto-updates via Dependabot and GitHub Actions
- auto-generated documentation
- pre-commit hooks for ensuring formatting, linting, security checks, etc.

## changes required to use this as a starter template
- ensure new repo has the following GitHub Actions Secrets and local shell exports/environment variables stored in `./.env`
    - GH_TOKEN
    - GH_GHCR_TOKEN
    - DOCKERHUB_USERNAME
    - DOCKERHUB_TOKEN
    - QUAY_USERNAME
    - QUAY_TOKEN
    - SNYK_TOKEN
    - COSIGN_PRIVATE_KEY
    - COSIGN_PASSWORD
- find/replace golang-starter to new repo name
    - run `bash -c ./scripts/use_starter.sh $NEW_PROJECT_NAME_GOES_HERE`
    - to rename with a different GitHub username `bash -c ./scripts/use_starter.sh $NEW_PROJECT_NAME_GOES_HERE $GITHUB_USERNAME_GOES_HERE`
- set built packages visibility in GitHub packages to public
    - navigate to https://github.com/users/$USERNAME/packages/container/$REPO/settings
    - scroll down to "Danger Zone"
    - change visibility to public

## changes required to update golang version
- run `bash -c ./scripts/update_golang_version.sh $NEW_VERSION_GOES_HERE`
