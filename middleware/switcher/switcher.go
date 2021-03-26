package switcher

import (
    //"fmt"
    //"net/http"
    
    "github.com/gofiber/fiber/v2"
)

type Config struct {
    Skip     bool
    Next     func(c *fiber.Ctx) bool
    Response func(c *fiber.Ctx) error
}

var ResponseAtOnceConnClose = func(c *fiber.Ctx) error {
    return c.Context().Conn().Close()
}


var Next = func(c *fiber.Ctx) bool {return c.IP() == "127.0.0.1"}

var ConfigDefault = Config{
    Next:     Next,
    Response: ResponseAtOnceConnClose,
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
        if cfg.Skip {return c.Next()}
        if cfg.Next != nil && cfg.Next(c) {
            return c.Next()
        }
        return cfg.Response(c)
    }
}
