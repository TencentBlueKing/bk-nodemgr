---
name: bk-nodemgr-pkg-design
description: Use when designing, implementing, extending, or refactoring an independent Go pkg capability in bk-nodemgr, especially IHandler entrypoints, hidden implementation details, interface-based dependencies, same-contract composition, assembly, or changes spanning a contract and its consumers. Not for ordinary local edits, generic Go syntax, or third-party API protocol details alone.
---

# Independent Package Design

## Purpose

Build capabilities that callers can use without reconstructing their implementation. Use abstraction to connect layers, encapsulation to isolate changes, and consistent conventions to make both local and end-to-end changes understandable.

Success is not a tidy directory or the fewest changed files. Implementation changes should stay local where possible; necessary contract changes should have a discoverable, complete impact path.

## Start From the Consumer and Local Evidence

1. Identify the real consumer, required outcome, inputs, outputs, and current behavior. State the capability without naming its first host or underlying technology.
2. Inspect applicable project rules, actual callers and assembly, and 2-3 comparable implementations. Prefer mature third-party adapter conventions for capability boundaries, unified entrypoints, and hidden clients; do not endorse an entire legacy package or copy its defects.
3. Distinguish mandatory local rules from examples. Do not silently rename existing contracts, rewrite constructors, or adopt a new architecture for uniformity. Disclose conflicts and obtain approval for the affected departure.
4. Apply the conventions below to the actual task, not to an imagined future platform. These bundled references are self-contained; they do not require access to an originating repository example.

## Workflow

### 1. Name the capability and its boundary

Write one sentence describing what the package provides and one describing what remains with its consumers. A package may own a domain; independence does not require business-free utilities. Service-specific policy belongs in its owning service, even when it implements a shared interface.

Separate capability from delivery: reading content does not imply HTTP routing or rendering. Implementation-specific configuration must not make a neutral capability depend on its first application host.

### 2. Define the contract before its implementation

Use `package.IHandler` as the default primary entrypoint for a new capability: the package name already supplies the domain. Define only required operations. If needed, compose cohesive subinterfaces and let consumers accept the smallest capability they need.

Specify result meaning, absence, errors, context lifetime, resource ownership, mutation, and side effects where relevant. Returning a client, filesystem, or protocol response and making callers perform the real operation is not encapsulation.

Read [Package Conventions](references/package-conventions.md) for naming, public surface, contract semantics, and file responsibilities.

### 3. Hide details and connect dependencies explicitly

Implement the contract with `Handler` or mechanism-qualified handler types. Keep state and helpers private unless consumers need them. Use explicit compile-time conformance checks to associate implementations with interfaces.

Consumers hold abstract capabilities. Constructors and assembly may know concrete types; concrete-returning constructors are valid. Centralize implementation selection and composition in the appropriate assembly location, without requiring a registry, DI framework, global singleton, or mode switch.

For external protocols, isolate transport requests and responses behind the capability boundary and reuse the established client. A package without an external protocol does not need a synthetic `cli` layer.

### 4. Choose the smallest justified extension

- **Implementation change:** preserve the capability and keep the change inside its implementation.
- **Parallel implementation:** implement the existing contract for a genuinely required alternative mechanism.
- **Composition:** depend on an existing capability, add one coherent responsibility, and implement the same contract outward where semantics remain compatible.
- **Contract change:** when approved requirements change the capability itself, update the affected chain rather than hiding required behavior behind concrete assertions or optional methods.

Read [Dependencies and Composition](references/dependencies-and-composition.md) when adding an implementation, injecting dependencies, or layering policy. No speculative implementation is required to justify a real capability boundary, but an interface with no consumer need or isolation benefit is not useful by itself.

### 5. Check locality and traceability

Explain one relevant implementation change and, when the requirement affects the contract, its end-to-end route. Identify the actual consumers, contract, implementations, compositions, and assembly locations; do not stop at “find references.”

Read [Change Traceability](references/change-traceability.md) for a worked contract migration and impact checklist. Trace indirect dependencies and verify behavioral compatibility as well as signatures.

### 6. Implement and verify through the owned surface

Preserve existing names, behavior, and unrelated signatures unless their migration is authorized. Use project-owned error, context, logging, and configuration conventions where applicable. Compile the affected dependency graph and use meaningful existing checks or requirement-grounded behavior scenarios, respecting the project's test-delivery rules.

Do not claim that interface conformance proves cancellation, isolation, stream lifetime, or completeness. Do not create a new validation framework merely to demonstrate this skill.

## Delivery

Keep the explanation proportional to the change. Report:

- **Capability and boundary:** what is provided and what stays outside.
- **Contract and relationships:** entrypoint, relevant semantics, implementations, dependency types, and assembly.
- **Change path:** what stays local, what must change together, and why.
- **Conventions and departures:** local evidence, any approved deviation, and unresolved decisions.
- **Verification:** checks actually performed and remaining limits.

Do not generate a permanent architecture document or repeat this entire checklist in a small implementation report.

## Avoid

| Temptation                                                     | Better decision                                                          |
| -------------------------------------------------------------- | ------------------------------------------------------------------------ |
| Put host-specific routing/configuration in a shared capability | Keep delivery in the host and expose the required capability             |
| Add an interface but store concrete dependencies in consumers  | Connect consumers and implementations through the actual contract        |
| Force every constructor to return an interface                 | Inspect consumer dependencies; preserve valid concrete returns           |
| Put all capabilities in one giant `IHandler`                   | Compose cohesive subinterfaces; keep unrelated domains separate          |
| Add switches or type assertions at every call site             | Select and compose centrally; make required capabilities explicit        |
| Avoid all cross-layer changes to protect the abstraction       | Migrate an approved contract along its explicit dependency chain         |
| Copy a mature provider's entire layout                         | Reuse its conventions, not unnecessary protocol layers or legacy defects |

## Companion Boundaries

This skill guides construction and evolution, not a second architecture approval process. Use `bk-nodemgr-architecture-judgment` for unresolved placement or architectural departures, and `bk-nodemgr-thirdparty` for external API contracts. Use the relevant project context, error, logging, and config skills rather than duplicating their rules here.

## Evaluation

`evals/evals.json` contains self-contained capability, composition, contract-migration, naming-transfer, and unnecessary-abstraction scenarios. They need no original source code. Store execution outputs and review artifacts outside the repository.
