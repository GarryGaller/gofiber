package bodylimiter

import (
    "fmt"
    "net/http"

    "github.com/gofiber/fiber/v2"
)

type Config struct {
    Limit    uint64
    Next     func(c *fiber.Ctx) bool
    LimitReached func(c *fiber.Ctx, limit, value uint64) bool
    Response func(c *fiber.Ctx, limit uint64, value uint64) error
}

var Response = func(c *fiber.Ctx, limit, size uint64) error {
    return c.Status(413).JSON(fiber.Map{
        "status": fiber.StatusRequestEntityTooLarge,
        "message": fmt.Sprintf("%s:%d byte > %d byte",
            http.StatusText(413), size, limit),
    })
}

var Next = func(c *fiber.Ctx) bool { return false } // return c.IP() == "127.0.0.1" }

func LimitReached(c *fiber.Ctx, limit, value uint64) bool {
    return limit > 0 && value > limit
}


var ConfigDefault = Config{
    Limit:    0,
    Next:     Next,
    LimitReached: LimitReached,
    Response: Response,
}

func configDefault(config ...Config) Config {
    // Return default config if nothing provided
    if len(config) < 1 {
        return ConfigDefault
    }

    // Override default config
    cfg := config[0]

    // Set default values
    if cfg.Next == nil {
        cfg.Next = ConfigDefault.Next
    }
    
    if cfg.LimitReached == nil {
        cfg.LimitReached = ConfigDefault.LimitReached
    }

    if cfg.Limit == 0 {
        cfg.Limit = ConfigDefault.Limit
    }

    if cfg.Response == nil {
        cfg.Response = ConfigDefault.Response
    }

    return cfg
}

// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
    cfg := configDefault(config...)
    // Return new handler
    return func(c *fiber.Ctx) error {
        // Don't execute middleware if Next returns true
        if cfg.Next != nil && cfg.Next(c) {
            return c.Next()
        }
        size := uint64(len(c.Body()))
        if cfg.LimitReached(c, cfg.Limit, size) {
            return cfg.Response(c, cfg.Limit, size)
        }

        return c.Next()
    }
}
