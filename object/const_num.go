package object

import "time"

const (
	NUMGRPCClientTimeout      = 40 * time.Second
	NUMGRPCServerTimeout      = 40 * time.Second
	NUMHTTPClientTimeout      = 40 * time.Second
	NUMHTTPServerTimeout      = 40 * time.Second
	NUMJWTExpiration          = 40 * time.Hour
	NUMJWTNotBefore           = 40 * time.Second
	NUMRSAGenerateKeyBits     = 4096
	NUMSystemGracefulShutdown = 40 * time.Second
)
