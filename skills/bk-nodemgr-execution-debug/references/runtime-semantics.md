# Execution State and Retry Semantics

Read this when explaining retries, skipped or pending steps, apparent state contradictions, or a timeout. These are runtime interpretation rules, not additional API parameters. Confirm the deployed version when using implementation details to explain an incident; source alone cannot establish the incident's cause.

## Two Different Kinds of Retry

| Mechanism              | What changes                                                                  | How to interpret it                                                                                                                                                             |
| ---------------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Retry inside an action | The action implementation is invoked again within the same operation instance | Messages such as “retried action, for the No.N attempt” do not establish that another `oper_inst_id` was created. Read the final action state as well as intermediate failures. |
| Retry an operation     | A new operation instance is created and associated with the same operation    | Compare instance IDs and lifecycle evidence. Keep the original failure separate from the new attempt's outcome.                                                                 |

For a partial operation retry with a failed action, the runtime searches backward from that action to an allowed retry start point. If the source instance failed or timed out at operation level without a failed action, the current runtime instead falls back to starting from the beginning. It can therefore rerun steps **before** the failed step, not only the failed step itself. Steps before the selected start point are initialized as `skipped`; their action implementation is not rerun in that attempt. This is not a failure and does not imply missing execution logs. Consult the source instance when the user's question concerns what those earlier steps did.

Do not claim that every `skipped` state proves partial retry, or that skipped steps were freshly verified as successful. Establish the retry relationship from available evidence. Full operation retry creates fresh pending actions instead. Do not infer retry mode solely from a log line containing the word “retry”.

Source entrypoints: `pkg/workflow/worker.go` (`callActionDefWithRetry`, `executeAndWatchAction`); `pkg/workflow/controller.go` (`generateRetryOperationInstance`, `findPartialRetryStartIndex`, `createPartialRetryInstance`).

## Workflow Aggregate State Is Not a Live Recalculation

Node and plugin workflows are updated by background monitoring of operations' latest-instance summaries; a workflow-list request reads the stored result rather than forcing recomputation. Compare observation times and instance evidence when summaries disagree. A bounded follow-up can distinguish an in-progress update, but persistent disagreement needs investigation, not indefinite polling.

In the current aggregation logic, all latest instances successful yields `success`, all failed or timed out yields `failed`, and other terminal combinations yield `partial_failed`. Consequently, `partial_failed` does **not** prove that some targets succeeded and others failed: even all `terminated` can produce it. Report the actual per-target states rather than translating the label into an unsupported success/failure split.

Source entrypoints: `internal/backend/storage/node/storage.go` and `internal/backend/storage/plugin/storage.go` (`calWorkflowStatusAndTime`, workflow monitoring); `internal/backend/router/api-v3/node/workflow/list.go` and `internal/backend/router/api-v3/plugin/workflow/list.go` (stored workflow queries).

## Why the Instance Can Fail Without a Failed Action

The runtime can execute instance-level extra processing before an action and after it finishes, according to the instance lifecycle. If this processing fails, it can mark the operation instance `failed` even when the ordinary action states are successful, or before any action has run.

When instance state and action states appear inconsistent:

1. Confirm that the operation summary, instance data, and logs describe the **same instance** and comparable observation times.
2. Inspect `extra_execution_logs` as well as action logs. A pre/post-processing error is a valid investigation path; do not invent a failed action to explain the instance state.
3. Use the precise error and corresponding runtime evidence to establish causality. If extra processing logged no useful detail, follow the instance into service/APM evidence rather than assuming the action list is complete causal evidence.

Source entrypoints: `pkg/workflow/worker.go` (`doOperExtraExecution`, the calls before and after `doAction`).

## State and Time Are Object-specific

| Observation                                     | What it does and does not establish                                                                                                                                                                   |
| ----------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Operation instance `init`                       | An instance has been initialized; not proof that its action was dispatched or a remote command ran.                                                                                                   |
| Operation instance `launched` with `start_time` | The runtime records the timestamp while launching the action chain, before successful chain submission is known. It is not the remote command's start time or proof that a worker started the action. |
| Action `pending`                                | Awaiting execution or rescheduling; not automatically a failure. The runtime can return a retryable action to pending during worker shutdown.                                                         |
| Action `running` with its own start time        | The runtime entered action execution; this still does not by itself prove a remote system accepted or completed a command.                                                                            |
| Action `skipped`                                | Its implementation is not rerun on this execution path. Determine the reason from retry or other runtime evidence.                                                                                    |
| `terminated`                                    | The execution was terminated; do not attribute it to a user cancellation without evidence. Worker shutdown or context cancellation can also produce termination.                                      |
| `timeout`                                       | Check whether the evidence says action timeout or operation-instance timeout; these are distinct deadlines. Do not automatically blame the remote host.                                               |

Actions are chained. A failed or timed-out step can leave later steps pending because they were not reached; those later steps are not automatically additional failures or the cause of the first failure. A pending action can also retain an earlier start time when rescheduled, so do not infer current progress from one timestamp alone.

The runtime attempts to persist action logs periodically and at execution completion; writes can fail, so this does not guarantee complete logs. A single query can observe state and logs at different stages of being written. If this distinction matters, make a bounded follow-up observation or use service evidence; do not promise instant consistency or explain every persistent discrepancy as delay.

Source entrypoints: `pkg/workflow/controller.go` (`LaunchOperationInstance`); `pkg/workflow/operation/instance.go` (`Lifecycle.Launch`); `pkg/workflow/action/instance.go` (`Lifecycle.Pending`); `pkg/workflow/worker.go` (`executeAndWatchAction`, `handleLongTailAction`, `autoRefreshActionDataMsg`, `autoRefreshExtraExecutionMsg`).

## Check the ID Value, Not Just Its Log Label

Some runtime timeout, termination, and rescheduling messages label a value `oper-inst-id` while formatting the operation's `OperationID`. An observed `oper:…` value does not become an instance ID because of that label.

Preserve the value and verify its relationship. Query the operation's instance history with the actual `operation_id`, select the matching attempt using incident time or explicit association, then use the returned `oper_inst_id` for the log query. Do not replace `oper:` with `oper-inst:`, or use a truncated value ending in an ellipsis. If the relationship cannot be verified, report the ambiguity.

Source entrypoints: `pkg/workflow/worker.go` (`executeAndWatchAction`, `handleLongTailAction`); `pkg/identifier/identifier.go`.
