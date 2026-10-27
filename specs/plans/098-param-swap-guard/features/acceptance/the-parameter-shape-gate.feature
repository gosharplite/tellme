# Acceptance — round 098 `098-param-swap-guard`

> Plan-side acceptance Gherkin (PM layer). The subject is tellme's **own build**, so the "user" is the
> maintainer and the "command" is the gate. These Examples are carried by `tools/params`' own tests and by
> the round's mutation witnesses (W1–W10) — a Make target is not a godog surface (the round-094/096
> precedent: the E2E contract drives the built binary, and this round changes no binary behaviour).

Feature: The parameter-shape gate refuses compile-blind argument swaps

  tellme ships a mechanical scan that reds when two adjacent parameters of an identical type can be
  exchanged at a call site without a compile error. The scan is red-first: it reds on the tree it was
  written against, and ships green only because that tree was fixed.

  Rule: A production signature with a swappable pair is refused.
    The gate names the file, the line, the function, and which arm fired, so the first red run is the
    specification of the work.

    Example: An arity-four signature with two adjacent strings
      Given a production function taking four parameters, two adjacent ones of type string
      When the parameter-shape gate runs
      Then it fails, naming that function and arm a

    Example: A three-parameter signature with three adjacent integers
      Given a production function taking three parameters, all of type int
      When the parameter-shape gate runs
      Then it fails, naming that function and arm b
      # arm (b) is arity-independent, so a post-fix accident at n = 3 cannot pass vacuously

    Example: An injected run of two closures
      Given a production function taking two adjacent func-typed parameters of different signatures
      When the parameter-shape gate runs
      Then it fails, naming that function and arm c
      # arm (c) is the DI idiom ("a closure list belongs in a spec struct"), not a swap rule:
      # swapping two closures never compiles, so no swap claim is made for it

  Rule: A contract is scanned, not just an implementation.
    An interface method is the shape every implementation must adopt, so the gate reads top-level
    interface declarations too.

    Example: A swappable pair on a domain port method
      Given a top-level interface method taking five parameters, two adjacent ones of type int
      When the parameter-shape gate runs
      Then it fails, naming that interface method and arm a

  Rule: The hazard is also at the call site.
    A struct literal with unkeyed fields can be swapped just as silently, and the standard vet analyser
    only sees it across packages.

    Example: A same-package literal with unkeyed fields
      Given a same-package composite literal of a named struct with two unkeyed elements
      When the parameter-shape gate runs
      Then it fails, naming that literal and arm d

    Example: A keyed literal is accepted
      Given the same literal with its fields keyed by name
      When the parameter-shape gate runs
      Then it passes

  Rule: Exempt parameters do not create a hazard.
    A run of contexts, a run of empty interfaces, and a lone trailing variadic are collectors, not
    swappable pairs.

    Example: Adjacent context parameters
      Given a production function whose adjacent parameters are all of type context.Context
      When the parameter-shape gate runs
      Then it passes

    Example: A trailing variadic
      Given a production function ending in a single variadic parameter
      When the parameter-shape gate runs
      Then it passes

  Rule: An exemption must still be true.
    Clearing a violation is done by fixing it or by a reviewed allow-list entry that carries a reason —
    and an entry that no longer matches anything is itself a failure.

    Example: An empty allow-list on a clean tree, with a measured universe
      Given a tree with no violation and an allow-list with no entry
      When the parameter-shape gate runs
      Then it passes and reports every baseline figure with its label and universe

    Example: A stale exemption
      Given an allow-list entry for a site that no longer violates
      When the parameter-shape gate runs
      Then it fails, naming the stale entry

    Example: Regeneration cannot launder a new violation
      Given a new violation and the allow-list regeneration path
      When regeneration runs
      Then it refuses to grow the list

  Rule: The verdict reads the tree, so the test cache cannot answer it.
    A cached green would let a newly added violation pass vacuously, so the gate is always run uncached.

    Example: A violation against a warm cache
      Given a green gate run that left the test cache warm
      When a violation is added and the gate runs again
      Then it fails
