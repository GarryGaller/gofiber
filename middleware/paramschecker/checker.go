package paramschecker

import (
    "fmt"

    "github.com/gofiber/fiber/v2"
)

type Config struct {
    Targets   map[string]func(p string) bool
    Next      func(c *fiber.Ctx) bool
    Response  func(c *fiber.Ctx) error
    ReadLimit int
    Local     bool
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

var Response = func(c *fiber.Ctx) error {
    return c.Status(400).JSON(fiber.Map{
        "status": fiber.StatusBadRequest,
        "message": fmt.Sprintf(
            "Invalid the <%s:%s> parameter in the <%s> request",
            c.Locals("target"),
            c.Locals(c.Locals("target").(string)),
            c.Method()),
    })
}

func NotEmpty(p string) bool { return p != "" }

func MoreOne(p string) bool { return len(p) > 1 }

var ConfigDefault = Config{
    Targets:   make(map[string]func(p string) bool),
    Next:      Next,
    Response:  Response,
    ReadLimit: 1024,
}

func configDefault(config ...Config) Config {
    // Return default config if nothing provided
    if len(config) < 1 {
        return ConfigDefault
    }

    // Override default config
    cfg := config[0]

    // Set default values
    if cfg.ReadLimit == 0 {
        cfg.ReadLimit = ConfigDefault.ReadLimit
    }

    if cfg.Next == nil {
        cfg.Next = ConfigDefault.Next
    }    
    
    if cfg.Local {
        cfg.Next = NextIfLocal
    }

    if cfg.Targets == nil {
        cfg.Targets = ConfigDefault.Targets
    }

    if cfg.Response == nil {
        cfg.Response = ConfigDefault.Response
    }

    return cfg
}

func New(config ...Config) fiber.Handler {
    cfg := configDefault(config...)

    // Return new handler
    return func(c *fiber.Ctx) error {
        // Don't execute middleware if Next returns true
        if cfg.Next != nil && cfg.Next(c) {
            return c.Next()
        }
        for target, checker := range cfg.Targets {
            if checker == nil {
                checker = NotEmpty
            }

            data := c.Locals(target).(string)
            if len(data) > cfg.ReadLimit {
                data = data[:cfg.ReadLimit]
            }

            if !checker(data) {
                c.Locals("target", target)
                return cfg.Response(c)
            }
        }
        return c.Next()
    }
}
