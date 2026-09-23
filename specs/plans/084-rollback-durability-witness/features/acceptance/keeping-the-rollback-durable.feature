Feature: Keeping the rollback's durability guarantee falsifiable

  # Plan-side acceptance journey — round 084, anchor issue
  # https://github.com/gosharplite/tellme/issues/169. Round 081 shipped a
  # "durable and atomic" rollback (temp file + fsync + atomic rename) and
  # asserted that guarantee on six live surfaces, but no input could falsify the
  # fsync half. This round keeps the guarantee and gives it a witness that can
  # fail.
  #
  # The observable rollback contract is UNCHANGED, so this acceptance layer only
  # restates it — a reader of `-b` output cannot tell this round shipped. The
  # fsync clause itself is NOT observable through the CLI; per aixbdd-tmg
  # ADR 0006 it is carried by a unit-tier witness (a mechanism-seam pin that
  # reddens when the file fsync is removed) and its best-effort directory fsync
  # is recorded as an accepted-unwitnessed limit — neither is a Gherkin journey.
  # This layer is PM-readable business language; the executable interface truth
  # lives under specs/truth/features/cli/**.

  Rule: A rollback keeps its observable guarantees

    Example: Rolling back one turn removes only the last turn
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history holds two completed turns
      When the operator runs tellme with the back flag "1"
      Then tellme reports that one turn was rolled back
      And the session history holds one completed turn
      And the archive still holds every archived line
      And tellme exits successfully

    Example: A rollback that cannot safely rewrite the history changes nothing
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history holds two completed turns
      When a rollback cannot safely rewrite the active history
      Then the prior history is left intact
      And no temporary file is left behind
      And the rollback fails loudly
