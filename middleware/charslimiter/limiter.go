package charslimiter

import (
    //"fmt"
    "fmt"
    "net/http"
    "encoding/binary"
    
    "github.com/coocood/freecache"
    "github.com/gofiber/fiber/v2"
    
)


type Config struct {
    Cache         *freecache.Cache
    Limit         uint64
    Expiration    int
    Next          func(c *fiber.Ctx) bool
    LimitReached  func(c *fiber.Ctx, limit, value uint64) bool
    Response      func(c *fiber.Ctx, limit, chars uint64) error
    KeyGenerator  func(c *fiber.Ctx) string
}

func KeyGenFromIP(c *fiber.Ctx) string {
    if len(c.IPs()) > 0 {
        return c.IPs()[0]
    }
    return c.IP()
}

var Response = func(c *fiber.Ctx, limit, chars uint64) error {
    return c.Status(402).JSON(fiber.Map{
        "status": fiber.StatusPaymentRequired,
        "message": fmt.Sprintf("%s:%d chars more %d maxchars per hour",
            http.StatusText(402), chars, limit,
        ),
    })
}

func GetIPs(c *fiber.Ctx) []string {
    ips := c.IPs()
    if len(ips) == 0 {
        ips = append(ips, c.IP())
    }
    return ips
}

var Next = func(c *fiber.Ctx) bool { return false } //return c.IP() == "127.0.0.1" }

func LimitReached(c *fiber.Ctx, limit, value uint64) bool {
    return limit > 0 && value > limit
}

var ConfigDefault = Config{
    Cache:        freecache.NewCache(5 * 1024 * 1024), // 5 mb
    Limit:        0,
    Expiration:   3600,  // 1 hour
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

    return cfg
}
 
func CacheIncrUint(cache *freecache.Cache, key []byte, incr uint64) error {
    var err error
    buffer := make([]byte, 8)
    got, _, _ := cache.GetWithExpiration(key)
    ttl, _ := cache.TTL(key)
    value := binary.LittleEndian.Uint64(got)
    value += incr
    binary.LittleEndian.PutUint64(buffer, value)
    err = cache.Set(key, buffer, int(ttl))
    return err
}

func BytesFromUint(value uint64) ([]byte) {
    buffer := make([]byte, 8)
    binary.LittleEndian.PutUint64(buffer, value)
    return buffer
}


// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
    cfg := configDefault(config...)
    // Return new handler
    return func(c *fiber.Ctx) error {
        
        var value uint64
        var err error
        var exists bool
        // Don't cache response if Next returns true
        if cfg.Next != nil && cfg.Next(c) {
            return c.Next()
        }

        //-------------------------------------
        ip := []byte(cfg.KeyGenerator(c))
        bValue, _, err := cfg.Cache.GetWithExpiration(ip)
        exists = err == nil
        if exists {
            value = binary.LittleEndian.Uint64(bValue)
            if cfg.LimitReached(c, cfg.Limit, value) {
                return cfg.Response(c, cfg.Limit, value)
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
            buffer := BytesFromUint(uint64(chars))
            cfg.Cache.Set(ip, buffer, cfg.Expiration)
        } else {
            // key found
            CacheIncrUint(cfg.Cache, ip, uint64(chars))
        }
        
        // finish response
        return nil
    }
}
