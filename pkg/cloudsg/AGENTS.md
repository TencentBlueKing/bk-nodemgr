|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/cloudsg
|Overview:Cloud security group domain package|owns rule/inspection/pool-summary types and the rule coverage semantics that decide whether a permission is already granted|leaf package, imports stdlib only
|Structure:pkg/cloudsg:{cloud_security_group.go,cloud_security_group_coverage.go,cloud_security_group_coverage_test.go,doc.go}
|Where to look:domain types:pkg/cloudsg/cloud_security_group.go:{CloudSecurityGroupVendor,CloudSecurityGroupRuleAction,CloudSecurityGroupRule,EffectiveAction,CloudSecurityGroupInspection,CloudSecurityGroupPoolSummary}
|Where to look:coverage semantics:pkg/cloudsg/cloud_security_group_coverage.go:{CloudSecurityGroupRuleCovers,CloudSecurityGroupRuleCoveredBy}|helpers:{cloudSGProtocolCovers,cloudSGCIDRCovers,cloudSGPortsCover,cloudSGPortRangesUnionCovered,cloudSGParsePortRanges}
|Where to look:behavior spec:pkg/cloudsg/cloud_security_group_coverage_test.go:table tests for cidr/port/protocol/union/action cases
|Where to look:transport conversions:pkg/proto/{application,backend}/api/v3/cloud_security_group.go:pool summary↔proto
|Where to look:access point binding field:pkg/types/topo.go:AccessPoint.CloudSecurityGroupPoolID|stored by pkg/dao/mongo/accesspoint
|Conventions:CloudSecurityGroupVendor is only a typed string used by pool config to select the provider implementation|concrete vendor constants belong to the provider package, not here
|Conventions:Action empty→ACCEPT via EffectiveAction (uppercased)|coverage only considers ACCEPT rules on both sides|DROP never covers and is never covered
|Conventions:Covers=one rule fully covers target (action, protocol, cidr, ports)|CoveredBy=union of existing ACCEPT rules whose protocol and cidr each cover the target as a whole contribute ports
|Conventions:port expression=ALL|single port|list `a,b`|range `a-b`|ALL target is covered only by an ALL rule|adjacent/overlapping ranges merge before comparison
|Conventions:protocol ALL covers any protocol|empty protocol is TCP|cidr accepts bare IP or prefix|cidr is never unioned across rules
|Conventions:fail-closed direction=unparseable cidr/port or any doubt reports not covered|the only cost is a redundant authorization, never a skipped required one
|Conventions:CloudSecurityGroupInspection.Rules holds every live rule of the security group|the whole pool is managed, user rules and platform rules are not distinguished
|Dependencies:stdlib only (net/netip,sort,strconv,strings)|keep it free of pkg/proto, pkg/types, internal/**, and cloud SDK imports so any layer can import it without cycles
|Anti-patterns:do not call cloud APIs, storage, or config from here|provider calls, pool registry, and persistence stay in their own layers
|Anti-patterns:do not re-implement cidr/port/protocol comparison elsewhere|extend this package and its table tests instead
|Anti-patterns:do not loosen coverage to treat partial or unparseable matches as covered
|Verification:go test ./pkg/cloudsg|if converters touched:go test ./pkg/proto/...
