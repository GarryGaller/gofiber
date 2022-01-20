package charslimiter

import (
    "errors"
    "fmt"
    "net/http"
    "time"
    
    "github.com/gofiber/fiber/v2"
)

var ErrDatabaseNotInitialized = errors.New("database is not initialized")


type Config struct {
    Cache        *Cache
    Limit        uint64
    Expiration   int
    Next         func(c *fiber.Ctx) bool
    LimitReached func(c *fiber.Ctx, limit, value uint64) bool
    Response     func(c *fiber.Ctx, limit, chars uint64, period time.Duration) error
    KeyGenerator func(c *fiber.Ctx) string
    Local        bool
}

func GetIPs(c *fiber.Ctx) []string {
    ips := c.IPs()
    if len(ips) == 0 {
        ips = append(ips, c.IP())
    }
    return ips
}

func NextIfLocal(c *fiber.Ctx) bool {
    return c.IsFromLocal()
}

func Next(c *fiber.Ctx) bool {
    return false
}

func KeyGenFromIP(c *fiber.Ctx) string {
    if len(c.IPs()) > 0 {
        return c.IPs()[0]
    }
    return c.IP()
}

var Response = func(c *fiber.Ctx, limit, chars uint64, period time.Duration) error {
    return c.Status(402).JSON(fiber.Map{
        "status": fiber.StatusPaymentRequired,
        "message": fmt.Sprintf("%s:%d chars more %d maxchars for the period in %s",
            http.StatusText(402), chars, limit, period,
        ),
    })
}

func LimitReached(c *fiber.Ctx, limit, value uint64) bool {
    return limit > 0 && value > limit
}

var ConfigDefault = Config{
    Limit:        0,    // chars limit
    Expiration:   3600, // 1 hour
    Next:         Next,
    LimitReached: LimitReached,
    Response:     Response,
    KeyGenerator: KeyGenFromIP,
}

func configDefault(config ...Config) Config {
    // Return default config if nothing provided
    if len(config) < 1 {
        return ConfigDefault
    }

    // Override default config
    cfg := config[0]

    // Set default values
    if cfg.Cache == nil {
        cfg.Cache = ConfigDefault.Cache
    }

    if cfg.Next == nil {
        cfg.Next = ConfigDefault.Next
    }

    if cfg.Local {
        cfg.Next = NextIfLocal
    }

    if cfg.LimitReached == nil {
        cfg.LimitReached = ConfigDefault.LimitReached
    }

    if cfg.Limit == 0 {
        cfg.Limit = ConfigDefault.Limit
    }

    if cfg.Expiration == 0 {
        cfg.Expiration = ConfigDefault.Expiration
    }

    if cfg.Response == nil {
        cfg.Response = ConfigDefault.Response
    }

    if cfg.Cache == nil {
        cfg.Cache = NewCache(DefaultCacheConfig)
    }

    return cfg
}

// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
    cfg := configDefault(config...)

    // Return new handler
    return func(c *fiber.Ctx) error {

        var value interface{}
        var err error
        var exists bool
        // Don't limit response if Next returns true
        if cfg.Next != nil && cfg.Next(c) {
            return c.Next()
        }
        //-------------------------------------
        ip := cfg.KeyGenerator(c)
        value, exists = cfg.Cache.Get(ip)
        if exists {
            if cfg.LimitReached(c, cfg.Limit, value.(uint64)) {
                return cfg.Response(c, cfg.Limit, value.(uint64),
                    time.Duration(cfg.Expiration)*time.Second,
                )
            }
        }
        // Continue stack, return err to Fiber if exist
        if err := c.Next(); err != nil {
            return err
        }
        //-------------------------------------
        chars := c.Locals("chars").(int)
        if !exists {
            // key not found
            expires := time.Duration(cfg.Expiration) * time.Second
            cfg.Cache.Set(ip, uint64(chars), expires)
        } else {
            // if key found
            _, err = cfg.Cache.IncrementUint64(ip, uint64(chars))
            if err != nil {
                return err
            }
        }

        // finish response
        return nil
    }
}
