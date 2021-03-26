package iplimiter

import (
    "time"
    "translator/utils"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/limiter"
)

type Config struct {
    IPs []string
    Max int
    Expiration time.Duration
    KeyGenerator func(c *fiber.Ctx) string
    Next func(c *fiber.Ctx) bool
    LimitReached func(c *fiber.Ctx) error 
    Storage fiber.Storage
}

var Next = func(c *fiber.Ctx) bool { return false }

func GetIPs(c *fiber.Ctx) []string {
    ips := c.IPs()
    if len(ips) == 0 {
        ips = append(ips, c.IP())
    }
    return ips
}

func configDefault(config ...Config) (limiter.Config, []string) {
    // Return default config if nothing provided
    if len(config) < 1 {
        return limiter.ConfigDefault, make([]string, 0)
    }
    ips := config[0].IPs
    // Override default config
    cfg := limiter.Config{}

    if config[0].Max != 0 {
        cfg.Max = config[0].Max
    }

    if config[0].Expiration != 0 {
        cfg.Expiration = config[0].Expiration
    }

    if config[0].KeyGenerator != nil {
        cfg.KeyGenerator = config[0].KeyGenerator
    }

    if config[0].LimitReached != nil {
        cfg.LimitReached = config[0].LimitReached
    }

    if config[0].Next != nil {
        cfg.Next = config[0].Next
    }

    if config[0].Storage != nil {
        cfg.Storage = config[0].Storage
    }

    return cfg, ips
}

// New creates a new middleware handler
func New(config ...Config) fiber.Handler {

    cfg, ips := configDefault(config...)
    
    // Return new handler
    handler := limiter.New(cfg)
    return func(c *fiber.Ctx) error {
        if cfg.Next != nil && cfg.Next(c) {
            return c.Next()
        }
        if len(ips) != 0 && len(GetIPs(c)) != 0 {
            ip := GetIPs(c)[0]
            if utils.ContainsString(ips, ip) {
                return handler(c)
            }
        }
        return c.Next()
    }
}
