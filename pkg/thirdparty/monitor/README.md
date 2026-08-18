# monitor

`monitor` encapsulates Monitor API Gateway calls used by backend workflows.

The package exposes `IHandler` as the only caller-facing boundary. It currently supports a single scenario method: get or create the Agent event `data_id` for a business, returned as `taskProcEventDataID` input for the upper resolver.

Disabled mode is explicit: `NewNoOpHandler` never sends outbound requests and returns no value, so callers can apply their own defaults without writing default values back to storage.
