Feature: Discovering the skills available in the runtime home

  # Acceptance only (business journeys): when the operator's runtime home holds
  # skills, tellme lets the operator discover them by asking — each reported by
  # name and purpose — and the list reflects exactly the skills present (a
  # skill's supporting material is not a skill of its own; a runtime home with
  # none is reported honestly). A skill the operator points tellme to can then
  # be opened and followed on demand.

  Rule: tellme reports the skills present in the runtime home

    Example: An operator asks what skills are available
      Given the runtime home holds a skill "golang-patterns" and a skill "golang-testing"
      When the operator asks tellme to list its skills
      Then tellme reports a skill named "golang-patterns"
      And tellme reports a skill named "golang-testing"
      And each reported skill carries a short description of what it is for
      And tellme exits successfully

  Rule: A skill's supporting material is not reported as a skill of its own

    Example: A skill that also ships reference material
      Given the runtime home holds a skill "golang-patterns" that also ships its own reference material
      When the operator asks tellme to list its skills
      Then tellme reports the skill "golang-patterns" once
      And tellme does not report that skill's reference material as a separate skill
      And tellme exits successfully

  Rule: A runtime home with no skills is reported honestly

    Example: An operator who has no skills installed
      Given the runtime home holds no skills
      When the operator asks tellme to list its skills
      Then tellme reports that no skills are available
      And tellme exits successfully

  Rule: A listed skill can be opened and followed on demand

    Example: The assistant follows a skill the operator points it to
      Given the runtime home holds a skill "golang-testing" that explains how to write tests
      And the operator has listed the available skills
      When the operator asks tellme to follow the "golang-testing" skill
      Then tellme opens that skill's guidance
      And tellme's answer reflects the guidance in that skill
      And tellme exits successfully
