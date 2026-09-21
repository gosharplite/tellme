# Acceptance — a skill described with a YAML block scalar is listed by its text (round 075)

**Plan Package**: `specs/plans/075-skill-frontmatter-block-scalars`
**Feature under acceptance**: the skills catalog the operator's runtime home holds (`list_skills`)

> PM-readable acceptance journeys in business language (no tool/technical detail).
> `axb-dsl-refine` later splits these into executable interface truth under
> `specs/truth/features/cli/**`; coverage requires every Rule here to be carried by
> at least one interface Rule.

## Feature: A skill's description is listed as its real text, whatever its frontmatter style

### Rule: A description written as a folded block scalar is listed by its folded text

#### Example: A description folded across two lines

- **Given** the operator has a runnable tellme installation
- **And** the runtime home holds a skill whose description is written as a folded block scalar across two lines
- **And** a configured provider whose endpoint asks tellme to list its skills
- **When** the operator starts tellme with the prompt "What skills do you have?"
- **Then** the listing describes that skill with the folded text (the two lines joined by a single space)
- **And** the listing does not show a bare block-scalar indicator in place of the description

### Rule: Well-formed inline frontmatter keeps working

#### Example: A description written inline is listed unchanged

- **Given** the operator has a runnable tellme installation
- **And** the runtime home holds a skill whose description is written inline
- **And** a configured provider whose endpoint asks tellme to list its skills
- **When** the operator starts tellme with the prompt "What skills do you have?"
- **Then** the listing describes that skill with exactly the inline text
