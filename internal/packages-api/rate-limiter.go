package packagesApi

import (
	"net"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	requestsPerWindow = 600
	requestWindow     = time.Minute
)

type ipRateLimiter struct {
	trustedProxies []*net.IPNet
	development    bool
}

func newIPRateLimiter() *ipRateLimiter {
	return &ipRateLimiter{
		trustedProxies: loadTrustedProxies(),
		development:    strings.EqualFold(strings.TrimSpace(os.Getenv("ENVIRONMENT")), "development"),
	}
}

func (limiter *ipRateLimiter) allow(ctx fiber.Ctx) (bool, error) {
	clientIP := limiter.clientIP(ctx)
	window := time.Now().Unix() / int64(requestWindow.Seconds())
	return redisAllowSearchRequest(ctx.Context(), clientIP, window, requestsPerWindow, requestWindow)
}

func (limiter *ipRateLimiter) clientIP(ctx fiber.Ctx) string {
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

func (limiter *ipRateLimiter) isTrustedProxy(ip net.IP) bool {
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
