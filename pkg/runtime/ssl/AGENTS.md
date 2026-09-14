|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/runtime/ssl
|OVERVIEW:TLS/TLCP config loading for standard TLS (crypto/tls) and GM TLCP (GB/T 38636, gotlcp); builds `*tls.Config`/`*tlcp.Config` from file paths; no serving/dialing logic
|WHERE TO LOOK:standard TLS client/server conf:ssl.go:{TLSConfig,NewClientTLSConf,NewServerTLSConf,loadCertificates,loadCa,Validate}
|WHERE TO LOOK:TLCP (GM) conf+dial resolution:tlcp.go:{GMEnabled,NewServerTLCPConf,NewClientTLCPConf,ResolveTLCPDialConf,loadTLCPCertificates,loadTLCPPrivateKey,loadTLCPCa}
|WHERE TO LOOK:consumers:pkg/rest/client/http_client.go:{setupTransportSecurity,tlcpDialTLSContext}|pkg/rest/server/server.go:{startWithTLS,startWithTLCP}|pkg/relayhandler/server.go:GSE server-api client
|CONVENTIONS:GMEnabled()=EncCertFile+EncKeyFile both set⇒TLCP (no separate protocol switch); standard TLS keeps CA+cert+key triple semantics
|CONVENTIONS:TLCP double certificates stay ordered [sign,enc]; CertFile/KeyFile=signing pair, EncCertFile/EncKeyFile=encryption pair
|CONVENTIONS:password semantics differ by path: standard=legacy PEM (RFC 1423, x509.DecryptPEMBlock kept for operator-provided keys); TLCP=encrypted PKCS#8 via pkcs8.ParsePKCS8PrivateKeySM2
|CONVENTIONS:any custom-DialTLSContext TLCP client must resolve config through ResolveTLCPDialConf (ServerName from dial addr, cloned per dial); empty ServerName+verify enabled makes gotlcp silently skip the hostname check
|CONVENTIONS:pair-completeness validation (enc implies sign pair) lives in pkg/config TLSConfig.Validate, not here
|CONVENTIONS:no logging in this package; errors wrapped with `fmt.Errorf("...: %w", err)`
|CONVENTIONS:gotlcp NetDialer.Timeout bounds TCP connect AND TLCP handshake; http.Transport.TLSHandshakeTimeout never applies to custom dialers
|ANTI-PATTERNS:do not construct tlcp.Config with empty ServerName while verification is enabled (hostname check silently skipped)
|ANTI-PATTERNS:do not add serving/listening/dialing logic here; config building only (startWithTLCP/tlcpDialTLSContext live in pkg/rest/*)
|ANTI-PATTERNS:do not use legacy PEM encryption for new TLCP keys; TLCP path supports encrypted PKCS#8 only
|ANTI-PATTERNS:do not read file paths from user input here; os.ReadFile targets are trusted service config (G304 nolint anchored)
|COMMANDS:go test ./pkg/runtime/ssl
