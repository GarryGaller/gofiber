package blacklist

import (
    //"fmt"
    //"net/http"
    "time"

    "translator/utils"

    "github.com/gofiber/fiber/v2"
)

type Config struct {
    IPs      []string
    Sleep    int
    Next     func(c *fiber.Ctx) bool
    Response func(c *fiber.Ctx) error
    Local    bool
}

func GetIPs(c *fiber.Ctx) []string {
    ips := c.IPs()
    if len(ips) == 0 {
        ips = append(ips, c.IP())
    }
    return ips
}

func NextIfLocal(c *fiber.Ctx) bool {
    return GetIPs(c)[0] == "127.0.0.1"
}

func Next(c *fiber.Ctx) bool {
    return false
}


var ResponseAtOnceConnClose = func(c *fiber.Ctx) error {
    return c.Context().Conn().Close()
}

var ResponseWithSleepConnClose = func(c *fiber.Ctx) error {
    sleep, ok := c.Locals("sleep").(time.Duration)
    if ok {
        c.Request().Reset()
        c.Request().ResetBody()
        time.Sleep(sleep * time.Second)
    }
    return nil
}

var ConfigDefault = Config{
    IPs:      make([]string, 0),
    Sleep:    0,
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
    
    if cfg.Local {
        cfg.Next = NextIfLocal
    } 

    if cfg.Sleep == 0 {
        cfg.Sleep = ConfigDefault.Sleep
    }

    if cfg.Sleep > 0 {
        cfg.Response = ResponseWithSleepConnClose
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

        if cfg.Sleep > 0 {
            c.Locals("sleep", cfg.Sleep)
        }

        if len(cfg.IPs) != 0 && len(GetIPs(c)) != 0 {
            ip := GetIPs(c)[0]
            if utils.ContainsString(cfg.IPs, ip) {
                return cfg.Response(c)
            }
        }
        return c.Next()
    }
}
