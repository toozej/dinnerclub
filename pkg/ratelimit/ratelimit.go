// Package ratelimit provides rate limiting middleware for the dinnerclub web application
// using the gin-rate-limit package to prevent abuse and ensure fair usage.
package ratelimit

import (
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Config holds rate limiting configuration for both general and authentication endpoints.
type Config struct {
	Enabled      bool
	Requests     int
	Period       string
	AuthRequests int
	AuthPeriod   string
}

// NewConfig creates a new rate limiting configuration with the specified parameters.
func NewConfig(enabled bool, requests int, period string, authRequests int, authPeriod string) *Config {
	return &Config{
		Enabled:      enabled,
		Requests:     requests,
		Period:       period,
		AuthRequests: authRequests,
		AuthPeriod:   authPeriod,
	}
}

// parseDuration safely parses a duration string with fallback
func parseDuration(durationStr string, fallback time.Duration) time.Duration {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Warnf("Invalid duration '%s', using fallback %v", durationStr, fallback)
		return fallback
	}
	return duration
}

// GeneralRateLimitMiddleware creates a general rate limiting middleware for all endpoints.
func GeneralRateLimitMiddleware(config *Config) gin.HandlerFunc {
	if !config.Enabled {
		log.Info("Rate limiting is disabled")
		return gin.HandlerFunc(func(c *gin.Context) {
			c.Next()
		})
	}

	period := parseDuration(config.Period, time.Minute)

	// Validate requests is non-negative before converting to uint
	if config.Requests < 0 {
		log.Warnf("Invalid requests value %d, using default 100", config.Requests)
		config.Requests = 100
	}

	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  period,
		Limit: uint(config.Requests),
	})

	mw := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: func(c *gin.Context, info ratelimit.Info) {
			log.Warnf("Rate limit exceeded for IP %s: %d requests in %v", c.ClientIP(), info.RemainingHits, period)
			c.JSON(429, gin.H{
				"error":   "Too many requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
		},
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	})

	log.Infof("General rate limiting enabled: %d requests per %v", config.Requests, period)
	return mw
}

// AuthRateLimitMiddleware creates a stricter rate limiting middleware for authentication endpoints.
func AuthRateLimitMiddleware(config *Config) gin.HandlerFunc {
	if !config.Enabled {
		return gin.HandlerFunc(func(c *gin.Context) {
			c.Next()
		})
	}

	period := parseDuration(config.AuthPeriod, time.Minute)

	// Validate auth requests is non-negative before converting to uint
	if config.AuthRequests < 0 {
		log.Warnf("Invalid auth requests value %d, using default 10", config.AuthRequests)
		config.AuthRequests = 10
	}

	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  period,
		Limit: uint(config.AuthRequests),
	})

	mw := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: func(c *gin.Context, info ratelimit.Info) {
			log.Warnf("Auth rate limit exceeded for IP %s: %d requests in %v", c.ClientIP(), info.RemainingHits, period)
			c.JSON(429, gin.H{
				"error":   "Too many authentication attempts",
				"message": "Authentication rate limit exceeded. Please try again later.",
			})
			c.Abort()
		},
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	})

	log.Infof("Auth rate limiting enabled: %d requests per %v", config.AuthRequests, period)
	return mw
}
