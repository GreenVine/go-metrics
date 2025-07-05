package serving

import (
	"context"
	"log"
	"sync"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// KeyExtractorFunc is a function that extracts a rate limit key from a request
type KeyExtractorFunc func(fullMethodName string, req any) string

// MethodNameKeyExtractor extracts the method name as the rate limit key
func MethodNameKeyExtractor(fullMethodName string, _ any) string {
	return fullMethodName
}

// RateLimiter is an interface for request rate limiting.
type RateLimiter interface {
	// Allow checks if a request should be accepted.
	// Returns false to reject the request.
	Allow(ctx context.Context, fullMethodName string, req any) (bool, string)
}

// RateLimitConfig provides configuration for each method.
type RateLimitConfig struct {
	// Method is the gRPC method name
	// Example: greenvine.gometrics.device.v1.UpsertConfigRequest
	Method string

	// QPSLimit is the rate limit in queries per second
	QPSLimit int64

	// KeyExtractor adds a custom key for more granular rate limiting
	// If nil, only the method name will be used as the rate limit key
	KeyExtractor KeyExtractorFunc
}

// NewRateLimiterConfig creates a new rate limiter configuration for a method without any QPS limits.
func NewRateLimiterConfig(method string) RateLimitConfig {
	return RateLimitConfig{
		Method:       method,
		QPSLimit:     0, // no QPS limit
		KeyExtractor: MethodNameKeyExtractor,
	}
}

// WithKeyExtractor adds a key extractor to the rate limiter configuration.
func (c RateLimitConfig) WithKeyExtractor(keyExtractor KeyExtractorFunc) RateLimitConfig {
	c.KeyExtractor = keyExtractor
	return c
}

// WithQPSLimit sets the QPS limit for the rate limiter configuration.
func (c RateLimitConfig) WithQPSLimit(qpsLimit int64) RateLimitConfig {
	c.QPSLimit = qpsLimit
	return c
}

// TokenBucketRateLimiter implements RateLimiter using the token bucket algorithm
type TokenBucketRateLimiter struct {
	sync.RWMutex

	// configs holds all rate limit configurations by method name
	configs map[string]RateLimitConfig

	// limiters maps rate limit keys to their token buckets
	limiters map[string]*rate.Limiter
}

// NewTokenBucketRateLimiter creates a new TokenBucketRateLimiter.
func NewTokenBucketRateLimiter(configs []RateLimitConfig) *TokenBucketRateLimiter {
	perMethodConfig := make(map[string]RateLimitConfig)
	for _, config := range configs {
		perMethodConfig[config.Method] = config
	}

	return &TokenBucketRateLimiter{
		configs:  perMethodConfig,
		limiters: make(map[string]*rate.Limiter),
	}
}

// Allow determines whether a request should be accepted or rate-limited.
func (t *TokenBucketRateLimiter) Allow(_ context.Context, fullMethodName string, request any) (bool, string) {
	t.RLock()
	config, ok := t.configs[fullMethodName]
	t.RUnlock()

	// Allow the request if no method-specific config is found.
	if !ok {
		return true, ""
	}

	rateLimitKey := config.KeyExtractor(fullMethodName, request)
	if rateLimitKey == "" {
		// Reject the request if the rate limit key cannot be derived (fail-close).
		return false, ""
	}

	// Get or create a rate limiter for this rateLimitKey
	t.Lock()
	limiter, ok := t.limiters[rateLimitKey]
	if !ok {
		limiter = rate.NewLimiter(rate.Limit(config.QPSLimit), 1 /** burst */)
		t.limiters[rateLimitKey] = limiter
	}
	t.Unlock()

	return limiter.Allow(), rateLimitKey
}

// RateLimitInterceptor returns a new unary server interceptor for rate limiting
func RateLimitInterceptor(limiter RateLimiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		methodName := extractRequestMethod(req)
		if methodName == "" {
			return nil, status.Errorf(codes.Internal, "failed to determine method name for request: %v", req)
		}

		if allow, rateLimitKey := limiter.Allow(ctx, methodName, req); !allow {
			log.Printf("Request %q with rate-limit key %q was rejected", info.FullMethod, rateLimitKey)
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

func extractRequestMethod(maybeProto any) string {
	if p, ok := maybeProto.(proto.Message); ok {
		log.Println(string(p.ProtoReflect().Descriptor().FullName()))
		return string(p.ProtoReflect().Descriptor().FullName())
	}

	return ""
}
