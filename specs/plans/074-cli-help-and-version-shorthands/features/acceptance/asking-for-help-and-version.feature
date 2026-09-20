# Acceptance — the operator can ask for help and the version with short flags (round 074)

**Plan Package**: `specs/plans/074-cli-help-and-version-shorthands`
**Feature under acceptance**: tellme's command-line flag surface (`-h`/`--help`, `-v`/`--version`)

> PM-readable acceptance journeys in business language (no API/tool/technical detail).
> `axb-dsl-refine` later splits these into executable interface truth under
> `specs/truth/features/cli/**`; coverage requires every Rule here to be carried by
> at least one interface Rule.

## Feature: The operator can discover the flags and the running build

### Rule: Asking for help prints the flag list and succeeds

#### Example: The operator asks for help with the short flag

- **Given** the operator has a runnable tellme installation
- **When** the operator asks tellme for help with "-h"
- **Then** tellme prints its flag list
- **And** the help is reported as a success (not an error)
- **And** no request is sent to any model

#### Example: The operator asks for help with the long flag

- **Given** the operator has a runnable tellme installation
- **When** the operator asks tellme for help with "--help"
- **Then** tellme prints its flag list
- **And** the help is reported as a success (not an error)

### Rule: Asking for the build version prints it and succeeds

#### Example: The operator asks for the version with the short flag

- **Given** the operator has a runnable tellme installation
- **When** the operator asks tellme for the version with "-v"
- **Then** tellme prints its build version
- **And** the run shows no turn chrome
- **And** tellme exits successfully

#### Example: The operator asks for the version with the long flag

- **Given** the operator has a runnable tellme installation
- **When** the operator asks tellme for the version with "--version"
- **Then** tellme prints its build version
- **And** the run shows no turn chrome
- **And** tellme exits successfully

### Rule: An unrecognized flag is still refused

#### Example: An unknown flag is still a usage error

- **Given** the operator has a runnable tellme installation
- **When** the operator starts tellme with the unrecognized flag "-z"
- **Then** tellme reports the usage is invalid
- **And** tellme exits with the usage error code
