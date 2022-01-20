package proxychecker

import (
    "fmt"
    //"net/http"
    "strings"
    "translator/logger"
    
    "github.com/gofiber/fiber/v2"
)

const MESSAGE = "Proxy usage detected: your real IP address does not match the IP address of the request."

var PROXY_HEADERS = []string{
    "HTTP-X-FORWARDED-FOR",
    "HTTP-FORWARDED-FOR",
    "HTTP-X-FORWARDED",
    "HTTP-FORWARDED",
    "HTTP-CLIENT-IP",
    "HTTP-FORWARDED-FOR-IP",
    "HTTP-VIA",
    "VIA",
    "X-FORWARDED-FOR",
    "FORWARDED-FOR",
    "X-FORWARDED",
    "FORWARDED",
    "CLIENT-IP",
    "FORWARDED-FOR-IP",
    "HTTP-PROXY-CONNECTION",
    "X-Real-IP",
    "X-ProxyUser-Ip", // google services
}

type Config struct {
    Headers        []string
    Next           func(c *fiber.Ctx) bool
    IsProxyHeaders func(c *fiber.Ctx, headers []string) bool
    IsProxyIP      func(c *fiber.Ctx) bool
    Response       func(c *fiber.Ctx) error
    Local          bool
    logger         logger.LeveledLogger
}

func (cfg *Config) SetLogger(logger logger.LeveledLogger) {
    cfg.logger = logger
}

func (cfg *Config) Log() logger.LeveledLogger {
    return cfg.logger
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

func IsProxyHeaders(c *fiber.Ctx, proxyHeaders []string) bool {
    //ip := GetIPs(c)[0]
    findHeaders := make([]string, 0)
    //log := c.Context().Logger().(*logrus.Logger)
    log, ok := c.Locals("logger").(logger.LeveledLogger)
    if ok {
        log.Infof(
            "[ADDR     ] Local   :[%s] <-> Remote:[%s]\n",
            c.Context().LocalAddr(),
            c.Context().RemoteAddr(),
        )
        log.Infof("[ADDR     ] RemoteIP:[%s-%s] <-> XForwardedFor:[%s]\n",
            c.IP(),
            c.Context().RemoteIP().String(),
            strings.Join(c.IPs(), ", "),
        )
    }
    if len(proxyHeaders) != 0 {
        for _, header := range proxyHeaders {
            value := c.Get(header)
            if value != "" {
                findHeaders = append(findHeaders, fmt.Sprintf("%s:%s", header, value))
                //return true
            }
        }
    }
    if len(findHeaders) > 0 {
        if ok {
            log.Infof("[PROXY HEADERS] %s\n", strings.Join(findHeaders, ", "))
        }
        return true
    }

    return false
}

func IsProxyIP(c *fiber.Ctx) bool {
    var ip1, ip2, ips string
    ip1 = c.Get("X-Real-IP")
    ips = c.Get("Forwarded-For-IP")
    ipList := strings.Split(ips, ",")

    if len(ipList) > 0 {
        ip2 = strings.TrimSpace(ipList[0])
    }
    return ip1 != "" && ip2 != "" && ip1 != ip2
}

var Response = func(c *fiber.Ctx) error {
    return c.Status(403).JSON(fiber.Map{
        "status":  fiber.StatusForbidden,
        "message": MESSAGE,
    })
}

var ConfigDefault = Config{
    Headers:        PROXY_HEADERS,
    IsProxyHeaders: IsProxyHeaders,
    IsProxyIP:      IsProxyIP,
    Next:           Next,
    Response:       Response,
}

func configDefault(config ...Config) Config {
    // Return default config if nothing provided
    if len(config) < 1 {
        return ConfigDefault
    }

    // Override default config
    cfg := config[0]

    // Set default values
    if cfg.logger == nil {
        cfg.logger = ConfigDefault.logger
    }

    if cfg.Headers == nil {
        cfg.Headers = ConfigDefault.Headers
    }

    if cfg.Next == nil {
        cfg.Next = ConfigDefault.Next
    }

    if cfg.Local {
        cfg.Next = NextIfLocal
    }

    if cfg.Response == nil {
        cfg.Response = ConfigDefault.Response
    }

    if cfg.IsProxyHeaders == nil {
        cfg.IsProxyHeaders = ConfigDefault.IsProxyHeaders
    }

    if cfg.IsProxyIP == nil {
        cfg.IsProxyIP = ConfigDefault.IsProxyIP
    }

    return cfg
}

// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
    cfg := configDefault(config...)
    // Return new handler
    return func(c *fiber.Ctx) error {
        // Don't execute middleware if Next returns true
        c.Locals("logger", cfg.Log())
        if cfg.Next != nil && cfg.Next(c) {
            return c.Next()
        }

        //clientHeaders := c.Context().Request.Header.RawHeaders()
        if cfg.IsProxyHeaders(c, cfg.Headers) && cfg.IsProxyIP(c) {
            return cfg.Response(c)
        }

        return c.Next()
    }
}
