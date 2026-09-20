#!/usr/bin/env bash
# modelith-drift — ADVISORY: a modeled entity whose concept no longer appears in code.
#
# The domain model (docs/domain-model/**) is descriptive docs, subordinate to
# truth — but it is MAINTAINED (load-bearing): a modeled entity that has vanished
# from the code is a stale entry. This check reports, per modeled
# entity/enum/glossary term, whether ANY of its "code anchors" — its own name, a
# backticked identifier in its definition, or one of its enum values — still
# appears in the production Go sources. Zero anchors => advisory warning.
#
# It NEVER fails the build and is NOT a `make verify` member. Scope is the CODE
# model only (docs/domain-model/tellme.modelith.yaml); the quality and
# environment models are not code-backed.
#
# Anti-noise note: a NAME-DIFF check ("a new exported Go type has no model
# entry") was measured at ~79% false positives on this repo (ports and value
# types vastly outnumber modeled concepts) and is deliberately NOT implemented —
# shipping it would recreate the retired "noisy advisory surface becomes a
# permanent muse" failure (the RF-063-10/RF-068-1 class). See
# docs/domain-model/README.md -> Drift guard.
set -euo pipefail

MODEL="${1:-docs/domain-model/tellme.modelith.yaml}"
[ -f "$MODEL" ] || { echo "modelith-drift: model file not found: $MODEL"; exit 0; }

# Deliberate exceptions: modeled terms that describe an ABSENCE — their anchors
# are absent from code by design (a settled exclusion, not a missing concept).
EXCEPTIONS=" NoSecurityLayer "

code_tokens="$(mktemp)"; blocks="$(mktemp)"
trap 'rm -f "$code_tokens" "$blocks"' EXIT

# 1. Production code identifier set (comments included — deliberately lenient).
grep -rhoE '[A-Za-z_][A-Za-z0-9_]*' internal/ cmd/ --include='*.go' --exclude='*_test.go' 2>/dev/null | sort -u > "$code_tokens" || true

# 2. Per-entity anchors: name <TAB> token token token ...
awk '
  /^(entities|enums|glossary):[[:space:]]*$/ { if (cur != "") print cur; cur=""; sec=1; next }
  sec && /^[a-z]/ { if (cur != "") print cur; cur=""; sec=0; next }
  sec && /^  [A-Za-z][A-Za-z0-9]*:/ {
    if (cur != "") print cur;
    n=$0; sub(/^  /,"",n); sub(/:.*$/,"",n);
    cur=n"\t"n" "; next
  }
  sec {
    line=$0;
    while (match(line, /`[^`]+`/)) {
      t=substr(line, RSTART+1, RLENGTH-2);
      gsub(/^[^A-Za-z]+/, "", t);
      sub(/[^A-Za-z0-9_].*$/, "", t);
      if (t ~ /^[A-Za-z][A-Za-z0-9_]*$/) cur=cur t " ";
      line=substr(line, RSTART+RLENGTH);
    }
    if (match($0, /- name:[[:space:]]*/)) {
      v=substr($0, RSTART+RLENGTH); sub(/[[:space:]].*$/, "", v);
      if (v ~ /^[A-Za-z][A-Za-z0-9_-]*$/) cur=cur v " ";
    }
  }
  END { if (cur != "") print cur }
' "$MODEL" | sort -u > "$blocks"

# 3. Report entities with zero code anchors.
checked=0; stale=0
while IFS=$'\t' read -r name tokens; do
  [ -n "$name" ] || continue
  case "$EXCEPTIONS" in *" $name "*) continue ;; esac
  checked=$((checked + 1))
  found=0
  for t in $tokens; do
    if grep -qxF "$t" "$code_tokens"; then found=1; break; fi
  done
  if [ "$found" -eq 0 ]; then
    echo "  ⚠ $name — no code anchor (name/type/enum value) found in production Go"
    stale=$((stale + 1))
  fi
done < "$blocks"

if [ "$stale" -gt 0 ]; then
  echo "  → $stale of $checked modeled term(s) look STALE — review $MODEL against the code"
else
  echo "  ✓ every modeled term still has a code anchor ($checked checked)"
fi
exit 0  # advisory only — never fails
