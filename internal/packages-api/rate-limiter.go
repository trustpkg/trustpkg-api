package packagesApi

import (
	"net"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	DefaultRequestsPerWindow = 600
	DefaultRequestWindow     = time.Minute
)

type RateLimitConfig struct {
	Key               string
	RequestsPerWindow int
	RequestWindow     time.Duration
}

type RateLimiter interface {
	Allow(ctx fiber.Ctx) (bool, error)
}

type IPRateLimiter struct {
	key               string
	requestsPerWindow int
	requestWindow     time.Duration
	trustedProxies    []*net.IPNet
	development       bool
}

func NewIPRateLimiter(config RateLimitConfig) *IPRateLimiter {
	if config.Key == "" {
		config.Key = "default"
	}
	if config.RequestsPerWindow <= 0 {
		config.RequestsPerWindow = DefaultRequestsPerWindow
	}
	if config.RequestWindow <= 0 {
		config.RequestWindow = DefaultRequestWindow
	}

	return &IPRateLimiter{
		key:               config.Key,
		requestsPerWindow: config.RequestsPerWindow,
		requestWindow:     config.RequestWindow,
		trustedProxies:    loadTrustedProxies(),
		development:       isDevelopmentEnvironment(),
	}
}

func (limiter *IPRateLimiter) Allow(ctx fiber.Ctx) (bool, error) {
	clientIP := limiter.clientIP(ctx)
	window := time.Now().UnixNano() / int64(limiter.requestWindow)
	return redisAllowRequest(ctx.Context(), limiter.key, clientIP, window, limiter.requestsPerWindow, limiter.requestWindow)
}

func (limiter *IPRateLimiter) clientIP(ctx fiber.Ctx) string {
	remoteIP := net.ParseIP(ctx.IP())
	if !limiter.development && remoteIP != nil && limiter.isTrustedProxy(remoteIP) {
		if cloudflareIP := net.ParseIP(strings.TrimSpace(ctx.Get("CF-Connecting-IP"))); cloudflareIP != nil {
			return cloudflareIP.String()
		}
	}

	if remoteIP != nil {
		return remoteIP.String()
	}
	return "unknown"
}

func (limiter *IPRateLimiter) isTrustedProxy(ip net.IP) bool {
	for _, proxy := range limiter.trustedProxies {
		if proxy.Contains(ip) {
			return true
		}
	}
	return false
}

func loadTrustedProxies() []*net.IPNet {
	var trusted []*net.IPNet
	for value := range strings.SplitSeq(os.Getenv("CLOUDFLARE_TRUSTED_PROXIES"), ",") {
		_, network, err := net.ParseCIDR(strings.TrimSpace(value))
		if err == nil {
			trusted = append(trusted, network)
		}
	}
	return trusted
}
