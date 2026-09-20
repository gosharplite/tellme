Feature: Searching file contents

  # Interface truth (CLI end, `chat` module) — the `search_files` tool (round 071; ADR 0043):
  # a bounded, deterministic in-file content search over a directory subtree — the missing half of
  # the reader trio (`list_files`/`read_files`/`get_tree` find and read a file but cannot locate
  # content). The query is a literal string by default; `is_regex: true` opts into a pattern. Each
  # match is reported as `path:line: <trimmed line>`, sorted deterministically (path asc, then line
  # asc), bounded by the tool resource contract with a hard degenerate 100-match cap. There is NO
  # path/safety boundary (a settled no-security-layer exclusion) and NO `WorkspacePolicy` directory
  # ignore list (the caller scopes with `path`) — a recorded divergence from the reference.
  # Acceptance journey: features/acceptance/searching-file-contents.feature.

  Rule: A query is found across the files under a directory

    # The fixture is chosen so that a depth-first walk order DIFFERS from the
    # required path order: the sub-folder "a" is visited before the file "a.go",
    # yet "a.go" sorts before "a/x.txt" ('.' < '/'). So this Example REDs unless
    # the tool sorts deterministically (TD-071-3).
    Example: A query that appears in several files returns the matching lines in path order
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "a.go" whose text is "hello from go"
      And the working directory contains a sub-folder "a"
      And the working directory contains a file "a/x.txt" whose text is "hello from txt"
      And a configured provider "test-model" whose endpoint searches the working directory for "hello" and then answers with "done"
      When the operator starts tellme with the prompt "Where is 'hello' used?"
      Then tellme searched the working directory using its search_files tool
      And the search result lists a match in "a.go"
      And the search result lists a match in "a/x.txt"
      And the search result lists the matches in path order
      And tellme exits successfully

  Rule: A search that finds nothing is a result, not an error

    Example: A query with no match reports zero matches
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "hello world"
      And a configured provider "test-model" whose endpoint searches the working directory for "absent" and then answers with "done"
      When the operator starts tellme with the prompt "Find anything called absent."
      Then the search result reported no matches
      And tellme exits successfully

  Rule: A search larger than the result bound is trimmed

    Example: A search over the result bound ends with a truncation marker
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" holding 200 lines containing "needle"
      And a configured provider "test-model" whose endpoint searches the working directory with a small result budget and then answers with "done"
      When the operator starts tellme with the prompt "Find needle."
      Then the search result was trimmed to what the run can hold
      And tellme exits successfully

  Rule: An invalid pattern is a stable tool error

    Example: A malformed regex is reported as a tool error
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "hello"
      And a configured provider "test-model" whose endpoint searches the working directory with an invalid pattern and then answers with "done"
      When the operator starts tellme with the prompt "Search with a pattern."
      Then the search result reported an invalid pattern
      And tellme exits successfully
