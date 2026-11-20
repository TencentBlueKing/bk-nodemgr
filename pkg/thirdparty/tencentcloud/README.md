# Tencent Cloud Security Group Handler

Provides the Tencent Cloud VPC security group scenarios used by backend.

- `Authorize` describes live ingress rules and creates only missing `ACCEPT` rules.
- `Inspect` is read-only and returns managed rules plus total ingress counts.
- Managed rules carry `bk-nodemgr-managed` in `PolicyDescription`.
- Rules are keyed by CIDR, port and case-insensitive protocol. Repeated authorization is idempotent.
- Quota failures return `QuotaExceededError`; no delete or merge operation is exposed.
