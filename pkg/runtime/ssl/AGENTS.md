|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/runtime/ssl
|Overview:Pure runtime conversion of CA/Cert/Key content into standard-library *tls.Config for clients and servers; no file I/O, configuration resolution, or business logic
|Structure:pkg/runtime/ssl:{ssl.go,README.md}
|Where to look:value holder and enablement:ssl.go:{TLSConfig,IsEnabled}
|Where to look:TLS builders:ssl.go:{NewClientTLSConf,NewServerTLSConf}|material parsing:ssl.go:{loadCa,loadCertificates}
|Where to look:upstream source resolution and validation:pkg/config/config.go:TLSConfig.{ToSSLConfig,Validate,IsEnabled,GetCA,GetCert,GetKey}
|Conventions:TLSConfig is a value holder:{InsecureSkipVerify,VerifyClient bool; CA,Cert,Key []byte; Password string}|material fields contain bytes, never file paths
|Conventions:runtime IsEnabled=CA+Cert+Key all non-empty; no certificate parsing|config IsEnabled=all three source categories present|callers gate TLS setup; builders do not gate themselves
|Conventions:NewClientTLSConf sets {RootCAs,Certificates,InsecureSkipVerify}; ignores VerifyClient; leaves MinVersion unset|InsecureSkipVerify disables server chain/hostname verification; testing only
|Conventions:NewServerTLSConf sets {ClientCAs,Certificates,MinVersion=TLS1.2}|VerifyClient=true→ClientAuth=RequireAndVerifyClientCert
|Conventions:ToSSLConfig validates enabled configurations before resolving material; inline content takes precedence over files|disabled configuration→flags-only runtime config, no source reads
|Conventions:source pairing and password-requires-key validation belong to pkg/config.TLSConfig.Validate; do not duplicate here|certificate parsing belongs to {loadCa,loadCertificates}
|Conventions:loadCa→x509.CertPool.AppendCertsFromPEM; invalid CA PEM returns error|loadCertificates→tls.X509KeyPair
|Conventions:Password supports legacy PEM-encrypted private keys via deprecated x509.DecryptPEMBlock; retain for existing deployments, do not extend to new scenarios
|Conventions:No logging or global state; return errors to callers
|Where to look:enablement consumers:pkg/rest/{server/server.go,client/http_client.go}; pkg/discover/etcddiscover/etcd.go; internal/{backend,file}/service/service.go:Redis/Mongo wiring; pkg/config/file.go:export URL scheme
|Anti-patterns:Do not change IsEnabled semantics without checking gated consumers, including config-layer gates and HTTP/HTTPS export URLs
|Anti-patterns:Do not read files or resolve inline/base64 sources here; source resolution belongs upstream in pkg/config
|Anti-patterns:Do not add service-specific TLS cipher/policy tweaks; consuming services own their policy
