package geoip

import (
	//"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/ip2location/ip2location-go"
	"github.com/sirupsen/logrus"
	"translator/utils"
)

const IP2LOCATION_PATH = "IP2LOCATION-LITE-DB1.BIN"

type Config struct {
	DB        *ip2location.DB
	Path      string
	Countries []string
	Next      func(c *fiber.Ctx) bool
	Response  func(c *fiber.Ctx) error
	Local     bool
	logger    *logrus.Logger
}

func (cfg *Config) SetLogger(logger *logrus.Logger) {
	cfg.logger = logger
}

func (cfg *Config) Log() *logrus.Logger {
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
	return GetIPs(c)[0] == "127.0.0.1"
}

func Next(c *fiber.Ctx) bool {
	return false
}

var Response = func(c *fiber.Ctx) error {
	return c.Status(403).JSON(fiber.Map{
		"status":  fiber.StatusForbidden,
		"message": http.StatusText(403),
	})
}

func OpenDB(path string) (*ip2location.DB, error) {
	return ip2location.OpenDB(path)
}

var ConfigDefault = Config{
	Countries: make([]string, 0),
	Path:      IP2LOCATION_PATH,
	Next:      Next,
	Response:  Response,
	logger:    logrus.New(),
}

func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		return ConfigDefault
	}

	// Override default config
	cfg := config[0]

	if cfg.logger == nil {
		cfg.logger = ConfigDefault.logger
	}

	if cfg.Path == "" {
		cfg.Path = ConfigDefault.Path
	}

	if !utils.FileExists(cfg.Path) {
		cfg.Next = func(c *fiber.Ctx) bool { return true }
		cfg.Log().Errorf("[GEOIP    ] File not found\n: %s", cfg.Path)
	} else {
		db, _ := OpenDB(cfg.Path)
		//if err != nil {
		//    panic(err)
		//}
		cfg.DB = db

		if cfg.Next == nil {
			cfg.Next = ConfigDefault.Next
        }
       
        if cfg.Local {
		    cfg.Next = NextIfLocal
		} 
    
		if cfg.Countries == nil {
			cfg.Countries = ConfigDefault.Countries
		}

		if cfg.Response == nil {
			cfg.Response = ConfigDefault.Response
		}
	}
	return cfg
}

// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
	cfg := configDefault(config...)
	//fmt.Printf("Countries\n: %s", cfg.Countries)
	// Return new handler
	return func(c *fiber.Ctx) error {
		c.Locals("logger", cfg.Log())
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}
		if len(cfg.Countries) != 0 && len(GetIPs(c)) != 0 {
			ip := GetIPs(c)[0]
			results, err := cfg.DB.Get_all(ip)
			if err == nil {
				if !utils.ContainsString(cfg.Countries, results.Country_short) {
					c.Locals("ip", ip)
					c.Locals("country", results.Country_short)
					return cfg.Response(c)
				}
			}
		}
		return c.Next()
	}
}
