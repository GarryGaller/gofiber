package htmlchecker

import (
    "fmt"
    "strings"
    "github.com/gofiber/fiber/v2"
    "github.com/gabriel-vasile/mimetype"
)

type Config struct {
    Targets  map[string]func(p string) bool
    Next     func(c *fiber.Ctx) bool
    Response func(c *fiber.Ctx) error
    ReadLimit int
}

var Response = func(c *fiber.Ctx) error {
    return c.Status(400).JSON(fiber.Map{
        "status": fiber.StatusBadRequest,
        "message": fmt.Sprintf("%s is not html", c.Locals("target")),
        })
}

func Next(c *fiber.Ctx) bool { return false }

func IsHTML2(p string) bool { 
    mime := mimetype.Detect([]byte(p))
    t := strings.TrimSpace(strings.Split(mime.String(), ";")[0])
    return  t == "text/html"
} 

func IsHTML(p string) bool { 
    mime := mimetype.Detect([]byte(p))
    return mime.Is("text/html")
} 


var ConfigDefault = Config{
    Targets:  make(map[string]func(p string) bool),
    Next:     Next,
    Response: Response,
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
