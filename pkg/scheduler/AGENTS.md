|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/scheduler
|OVERVIEW:shared lightweight task scheduler for periodic background tasks; owns task registration/lifecycle, cron interval parsing, timeout context creation, panic recovery, duplicate-run prevention
|BOUNDARY:generic infrastructure only|business task definitions/policies belong in internal service packages or caller packages|retry/backoff/business compensation are caller-owned
|STRUCTURE:pkg/scheduler:{scheduler.go,context.go,logger_adaptor.go,scheduler_test.go,README.md,example/example.go}
|WHERE TO LOOK:public contracts:scheduler.go:{Scheduler,Task,NewTask,NewScheduler,NextActiveTime,CronParser}
|WHERE TO LOOK:interval presets:context.go:{Yearly,Monthly,Weekly,Daily,Hourly,Every constants,Every}
|WHERE TO LOOK:cron logging bridge:logger_adaptor.go:{LoggerAdapter,formatString,formatTimes}
|WHERE TO LOOK:behavior examples:example/example.go|README.md|scheduler_test.go
|CONVENTIONS:use contextx.IContext in task callbacks; create cancellation/deadline via contextx helpers to preserve project context semantics
|CONVENTIONS:keep task IDs stable and unique; RegisterTask returns duplicate-task error instead of overwriting existing scheduled tasks
|CONVENTIONS:task timeout is enforced in executeTask with contextx.WithTimeout; task functions must observe ctx cancellation and return promptly
|CONVENTIONS:robfig/cron is configured with seconds support plus SkipIfStillRunning and Recover; preserve no-overlap and panic-isolation behavior
|CONVENTIONS:scheduler package logs infrastructure failures only; callers own business logs, retry strategy, idempotency, and side-effect rollback
|CONVENTIONS:ListTask returns a copied map view; do not expose internal task map or mutate scheduler state without mutex protection
|CONVENTIONS:interval strings must be valid cron expressions accepted by CronParser; duration inputs to NewTask are converted to Every+duration
|ANTI-PATTERNS:do not place service-specific periodic task logic, workflow policy, or DB cleanup rules inside pkg/scheduler
|ANTI-PATTERNS:do not replace contextx.IContext with raw context.Context on task APIs
|ANTI-PATTERNS:do not remove SkipIfStillRunning or add parallel execution for the same task ID without explicit contract change
|ANTI-PATTERNS:do not make scheduler silently retry failed tasks; failure is logged and next scheduled run handles re-entry
|ANTI-PATTERNS:do not swallow panics/errors without logger.G.Sys signal and task-id context
|ANTI-PATTERNS:do not make cron/parser behavior depend on wall-clock globals beyond explicit time inputs like NextActiveTime(from)
|DEPENDENCIES:contextx for context lifecycle|pkg/logger for infrastructure logs|github.com/robfig/cron/v3 for scheduling
|COMMANDS:go test ./pkg/scheduler
