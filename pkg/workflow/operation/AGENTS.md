|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/workflow/operation
|Overview:Operation definition contracts, runtime orchestration model, instance lifecycle management, and retry semantics for workflow operations
|Structure:pkg/workflow/operation:{definition.go,operation.go,instance.go}
|Where to look:operation definition contracts:definition.go:{Definition,DefinitionSnapshot,ExtraExecution}
|Where to look:operation runtime model:operation.go:{Operation,RetryMode,RetryFlag,Param}:trigger/instance tracking, retry flags, param passing
|Where to look:instance lifecycle:instance.go:{InstanceData,InstanceBriefData,InstanceMetadata,Lifecycle,State}:state machine, metadata, bilingual logging
|Where to look:state transitions:instance.go:{Lifecycle methods,State constants}:Launch→Start→End, state validation
|Where to look:retry semantics:operation.go:{RetryMode,RetryFlag}:ALL vs PARTIAL retry, source/retry instance tracking
|Conventions:operation has 1+ instances tracked in InstanceIDs[]|max 100 instances per operation (maxInstanceNum)
|Conventions:state transitions must be deterministic and idempotent under retries|use Lifecycle.End() or EndByActionState() for state changes
|Conventions:bilingual messages via InstanceMetadata.SetMessage(zhMsg, enMsg, level) or Log() builder pattern
|Conventions:extra execution runs at specific lifecycle states (StateLaunched|StateSuccess|StateFailed|StateTimeout|StateTerminated)
|Conventions:retry flags track source→retry instance relationships via RetryFlag{SourceInstanceID,RetryInstanceID}
|Conventions:state flow:StateInit→StateLaunched→StateRunning→{StateSuccess|StateFailed|StateTimeout|StateTerminated}
|Conventions:action state maps to operation state:action.StateSuccess/StateSkipped→StateSuccess|action.StateFailed→StateFailed|action.StateTimeout→StateTimeout|action.StateTerminated→StateTerminated
|Conventions:use CheckStateFinished(state) to test terminal states|IsRunning()/IsTerminated() for lifecycle checks
|Conventions:InstanceBriefData for lightweight tracking (metadata+lifecycle+latest action)|InstanceData for full runtime (includes ActionInstanceDataMap)
|Anti-patterns:do not mutate Lifecycle.State directly|use Launch()/Start()/End() methods to ensure timestamp consistency
|Anti-patterns:do not bypass State.Validate() when accepting external state values
|Anti-patterns:do not exceed maxInstanceNum (100) instances per operation|CheckEnforceability() validates this
|Anti-patterns:do not assume operation has instances|use len(InstanceIDs) checks before accessing GetLastInstanceID()/GetNonLastInstanceIDs()
|Anti-patterns:do not break retry flag chain|each retry must reference SourceInstanceID correctly
|Anti-patterns:do not add business logic to operation primitives|keep this layer generic runtime infrastructure
|Dependencies:pkg/workflow/action for action instance contracts|pkg/workflow/common for Message and error types|pkg/contextx for context propagation
|Parent AGENTS:../AGENTS.md (workflow runtime overview)
