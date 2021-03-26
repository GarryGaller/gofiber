package countrieslimiter

import (
    //"fmt"
    "net/http" 
    
    "translator/utils"
    "github.com/ip2location/ip2location-go"
    "github.com/gofiber/fiber/v2"
)

var path = "IP2LOCATION-LITE-DB1.BIN"

type Config struct {
    DB *ip2location.DB
    Path string
    Countries []string
    Next func(c *fiber.Ctx) bool
    Response  func(c *fiber.Ctx) error     
}

func GetIPs(c *fiber.Ctx) []string {
    ips := c.IPs()
    if len(ips) == 0 {
        ips = append(ips, c.IP())
    }
    return ips
}

func Next(c *fiber.Ctx) bool {
    return c.IP() == "127.0.0.1" 
}

var Response = func(c *fiber.Ctx) error {
    return c.Status(403).JSON(fiber.Map{
        "status":  fiber.StatusForbidden,
        "message": http.StatusText(403),
    })
}


func OpenDB(path string) (*ip2location.DB, error) {
    return  ip2location.OpenDB(path)
}

var ConfigDefault = Config{
    Countries: make([]string,0),
    Next: Next,
    Response: Response,
}

func configDefault(config ...Config) Config {
    // Return default config if nothing provided
    if len(config) < 1 {
        return ConfigDefault
    }

    // Override default config
    cfg := config[0]

    if cfg.Path == "" {
        cfg.Path = ConfigDefault.Path
    }
    db, err := OpenDB(cfg.Path)
    if err != nil {
        panic(err)
    }
    cfg.DB = db

    if cfg.Next == nil {
        cfg.Next = ConfigDefault.Next
    }
    
    if cfg.Countries == nil {
        cfg.Countries = ConfigDefault.Countries
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
        if len(cfg.Countries) != 0 && len(GetIPs(c)) != 0 {
            ip :=  GetIPs(c)[0] 
            results, err := cfg.DB.Get_all(ip)
            if err == nil {
                if !utils.ContainsString(cfg.Countries, results.Country_short) {
                    c.Locals("ip", ip)
                    c.Locals("country",results.Country_short)
                    return cfg.Response(c)
                }  
            }
        }
        return  c.Next()
    }
}
