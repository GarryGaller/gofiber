package metrics

import (
    "github.com/ansrivas/fiberprometheus/v2"
    "github.com/gofiber/fiber/v2"
)

type Config struct {
    ServiceName string
    NameSpace   string
    SubSystem   string
    MetricsPath string
}

var ConfigDefault = Config{
    ServiceName: "translator",
    MetricsPath: "/metrics",
}

func configDefault(config ...Config) Config {
    // Return default config if nothing provided
    if len(config) < 1 {
        return ConfigDefault
    }

    // Override default config
    cfg := config[0]

    // Set default values
    if cfg.ServiceName == "" {
        cfg.ServiceName = ConfigDefault.ServiceName
    }

    if cfg.NameSpace == "" {
        cfg.NameSpace = ConfigDefault.NameSpace
    }

    if cfg.SubSystem == "" {
        cfg.SubSystem = ConfigDefault.SubSystem
    } 
    
    if cfg.MetricsPath == "" {
        cfg.MetricsPath = ConfigDefault.MetricsPath
    }

    return cfg
}

func New(app *fiber.App, config ...Config) fiber.Handler {
    cfg := configDefault(config...)

    prometheus := fiberprometheus.NewWith(
        cfg.ServiceName,
        cfg.NameSpace,
        cfg.SubSystem,
    )
    
    prometheus.RegisterAt(app, cfg.MetricsPath)
    
    return prometheus.Middleware
}
