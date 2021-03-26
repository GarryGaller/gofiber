package jwtlimiter

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "strconv"
    "strings"
    "time"

    "translator/gormutils"
    "translator/models"
    "translator/jwtutils"
    
    "github.com/dgraph-io/ristretto"
    "github.com/form3tech-oss/jwt-go"
    "github.com/gofiber/fiber/v2"
    jwtware "github.com/gofiber/jwt/v2"
    "gorm.io/gorm"
)

type Item struct {
    Token       string
    Chars       uint64
    CachedChars uint64
    CacheTime   int64
    Limit       uint64
    ExpiresAt   time.Time
}

type Config struct {
    Limit         uint64
    Next          func(c *fiber.Ctx) bool
    LimitReached  func(c *fiber.Ctx, limit, chars uint64) error
    Fail          func(c *fiber.Ctx, err error) error
    Cost          int64
    CacheSize     int64
    TTL           time.Duration
    JWT           jwtware.Config
    DB            *gormutils.DataBase
}

func NewCache(cfg *Config) *ristretto.Cache {
    cache, err := ristretto.NewCache(&ristretto.Config{
        NumCounters: cfg.CacheSize * 10, // number of keys to track frequency of.
        MaxCost:     cfg.CacheSize,      // maximum cost of cache 
        BufferItems: 64,            // number of keys per Get buffer.
        //OnExit:      OnExit,    //
        OnEvict: func(item *ristretto.Item) {
            // OnEvict is called for every eviction and passes the hashed key, value,
            // and cost to the function.
            //fmt.Printf("ON EVICT:%d %#v %#v\n",
            //    item.Key, item.Value, item.Expiration)

            token := item.Value.(Item).Token
            chars := item.Value.(Item).CachedChars

            if cfg.DB != nil {
                cfg.DB.GORM.
                    Model(models.New()).
                    Where("token = ?", token).
                    Update("chars", gorm.Expr("chars + ?", chars))
            }
        },

        OnReject: func(item *ristretto.Item) {
            // OnReject is called for every rejection done via the policy.
            //fmt.Printf("ON EVICT:%d %#v %#v\n",
            //    item.Key, item.Value, item.Expiration)

            token := item.Value.(Item).Token
            chars := item.Value.(Item).CachedChars

            if cfg.DB != nil {
                cfg.DB.GORM.
                    Model(models.New()).
                    Where("token = ?", token).
                    Update("chars", gorm.Expr("chars + ?", chars))
            }
        },
    })

    if err != nil {
        panic(err)
    }

    return cache
}

var ConfigDefault = Config{
    CacheSize:    10000,    // 0.1Mb
    Cost:         400,
    TTL:          0,      // бессрочно
    Next:         Next,
    LimitReached: LimitReached,
    Fail:         Fail,
    JWT: jwtware.Config{
        ContextKey:    "user",
        SigningMethod: "HS256",
    },
    DB: nil,
}

func GetIPs(c *fiber.Ctx) []string {
    ips := c.IPs()
    if len(ips) == 0 {
        ips = append(ips, c.IP())
    }
    return ips
}

func GetLimitFromToken(c *fiber.Ctx) uint64 {
    var limit uint64
    user := c.Locals("user")
    if user != nil {
        userToken := user.(*jwt.Token)
        claims := userToken.Claims.(jwt.MapClaims)
        iLimit, found := claims["Max"]
        if found {
            l, ok := iLimit.(uint64)
            if ok {
                limit = l
            }
        }
    } else {
        panic(errors.New("ContextKey <user> not found"))
    }

    return limit
}

func GetTokenFromHeader(c *fiber.Ctx) string {
    token := ""
    auth := strings.Split(c.Get("Authorization"), " ")
    if len(auth) > 1 {
        token = auth[1]
    }
    return token
}

func Next(c *fiber.Ctx) bool { return false }
//func Cost(value interface{}) int64 { return 400 } 


func LimitReached(c *fiber.Ctx, limit, chars uint64) error {
    return c.Status(402).JSON(fiber.Map{
        "status": fiber.StatusPaymentRequired,
        "message": fmt.Sprintf("%s:%d chars > %d maxchars",
            http.StatusText(402), chars, limit,
        ),
    })
}

func Fail(c *fiber.Ctx, err error) error {

    resp := &struct {
        Status  int
        Message string
    }{
        Status:  401,
        Message: err.Error(),
    }

    return c.Status(fiber.StatusUnauthorized).JSON(resp)
}

func OnExit(val interface{}) {
    /* На удаление записи из кэша делаем запись в БД*/
    //fmt.Printf("ON EXIT:%d\n", val)

}

func configDefault(config ...Config) Config {
    // Return default config if nothing provided

    if len(config) < 1 {
        return ConfigDefault
    }

    // Override default config
    cfg := config[0]

    if cfg.CacheSize == 0 {
        cfg.CacheSize = ConfigDefault.CacheSize
    }

    if cfg.TTL == 0 {
        cfg.TTL = ConfigDefault.TTL
    }

    if cfg.Limit == 0 {
        cfg.Limit = ConfigDefault.Limit
    }

    if cfg.Cost == 0 {
        cfg.Cost = ConfigDefault.Cost
    }

    if cfg.Next == nil {
        cfg.Next = ConfigDefault.Next
    }

    if cfg.LimitReached == nil {
        cfg.LimitReached = ConfigDefault.LimitReached
    }

    if cfg.Fail == nil {
        cfg.Fail = ConfigDefault.Fail
    }

    if cfg.DB == nil {
        cfg.DB = ConfigDefault.DB
    }

    if cfg.JWT.SigningKey == nil {
        cfg.JWT.SigningKey = ConfigDefault.JWT.SigningKey
    }

    if cfg.JWT.SigningMethod == "" {
        cfg.JWT.SigningMethod = ConfigDefault.JWT.SigningMethod
    }

    if cfg.JWT.ContextKey == "" {
        cfg.JWT.ContextKey = ConfigDefault.JWT.ContextKey
    }

    if cfg.JWT.TokenLookup == "" {
        cfg.JWT.TokenLookup = ConfigDefault.JWT.TokenLookup
    }

    if cfg.JWT.AuthScheme == "" {
        cfg.JWT.AuthScheme = ConfigDefault.JWT.AuthScheme
    }

    if cfg.JWT.ErrorHandler == nil {
        cfg.JWT.ErrorHandler = ConfigDefault.JWT.ErrorHandler
    }

    if cfg.JWT.Filter == nil {
        cfg.JWT.Filter = ConfigDefault.JWT.Filter
    }

    if cfg.JWT.SuccessHandler == nil {
        cfg.JWT.SuccessHandler = ConfigDefault.JWT.SuccessHandler
    }

    return cfg
}

// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
    cfg := configDefault(config...)
    cache := NewCache(&cfg)

    // do not validate
    cfg.JWT.Filter = func(c *fiber.Ctx) bool {
        // проверить наличие токена  в кэше
        token := GetTokenFromHeader(c)
        _, ok := cache.Get(token)
        //fmt.Printf("FILTER: %s|%#v\n", token, ok)
        return ok
    }
    // if validation failed
    cfg.JWT.ErrorHandler = func(c *fiber.Ctx, err error) error {
        
        token := GetTokenFromHeader(c)
        
        if err.Error() == "Token is expired" {
            if cfg.DB != nil {
                user := models.New()
                user.Status = "expired"
                result := cfg.DB.GORM.
                    Where("token = ?", token).
                    Updates(user)
                    //Delete(user)
                if result.RowsAffected == 0 {
                    fmt.Printf("The entry was not updated: %s|%#v\n",
                        token, result.Error)
                }
            }
        }

        resp := &struct {
            Status  int
            Message string
        }{
            Status:  401,
            Message: err.Error(),
        }
        //fmt.Printf("ErrorHandler:%#v\n", err)
        return c.Status(fiber.StatusUnauthorized).JSON(resp)
    }

    // на успешную авторизацию добавляем токен в кэш
    cfg.JWT.SuccessHandler = func(c *fiber.Ctx) error {
        
        token := GetTokenFromHeader(c)
        expires := c.Locals("expires").(time.Time)
        limit := c.Locals("max").(uint64)
        translated := c.Locals("translated").(uint64)
        // если лимит достигнут, отдаем ошибку
        if limit > 0 && translated > limit {
            return cfg.LimitReached(c, limit, translated)
        }

        // Continue stack, return err to Fiber if exist
        if err := c.Next(); err != nil {
            return err
        }

        // устанавливаем в кэше значение
        current := uint64(c.Locals("chars").(int))
        
        item := Item{
            Token:       token,
            Chars:       translated,
            CachedChars: current,
            CacheTime:   time.Now().Add(cfg.TTL).Unix(),
            Limit:       limit,
            ExpiresAt:   expires}

        //fmt.Printf("ttl:%s\n", cfg.TTL)
        _ = cache.SetWithTTL(token, item, cfg.Cost, cfg.TTL)
        //fmt.Printf("ADD CACHE: %s|%d\n", token, translated)
        translated += current
        if limit > 0 {
            c.Set("X-Charslimit-Limit", fmt.Sprint(limit))
            c.Set("X-Charslimit-Remaining", fmt.Sprint(limit-translated))
        }
        c.Set("X-ExpiresAt-Date", fmt.Sprint(expires))
        // finish response
        return nil
    }

    // Return new handler
    handler := jwtware.New(cfg.JWT)

    return func(c *fiber.Ctx) error {
        if cfg.Next != nil && cfg.Next(c) {
            return c.Next()
        }

        token := GetTokenFromHeader(c)
        // извлекаем лимит из токена
        //limit := GetLimitFromToken(c)
        val, ok := cache.Get(token)
        // если в кэше
        if ok {
            item := val.(Item)
            translated := item.Chars  // число символов из БД
            limit := item.Limit
            expires := item.ExpiresAt
            cacheTime := item.CacheTime
            // проверяем достигнут ли лимит
            translated += item.CachedChars
            
            if limit > 0 && translated > limit {
                return cfg.LimitReached(c, limit, translated)
            }

            // Continue stack, return err to Fiber if exist
            if err := c.Next(); err != nil {
                return err
            }

            // обновляем в кэше значение
            current := uint64(c.Locals("chars").(int))
            item.CachedChars += current
            remainTime := cacheTime - time.Now().Unix() 
            
            if remainTime <= 0 {remainTime = 1}
            ttl := time.Duration(remainTime) * time.Second
            //fmt.Printf("remainTime:%s\n", ttl)
            _ = cache.SetWithTTL(token, item, cfg.Cost, ttl)
            //fmt.Printf("EXIST: %s|%d\n", token, item.Chars)
            translated += current
            if limit > 0 {
                c.Set("X-Charslimit-Limit", fmt.Sprint(limit))
                c.Set("X-Charslimit-Remaining", fmt.Sprint(limit-translated))
            }
            c.Set("X-ExpiresAt-Date", fmt.Sprint(expires))

        } else {
            // если токен не в кэше, смотрим есть ли он в БД
            var translated uint64
            var limit uint64
            var expires time.Time

            if cfg.DB != nil {
                user := models.New()
                // читаем данные по токену из БД
                result := cfg.DB.GORM.First(user, "token = ?", token)
                if result.Error != nil {
                    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
                        //  if token not found
                        return cfg.Fail(c,
                            errors.New("The token was not found in the database"))
                    } else {
                        return cfg.Fail(c, result.Error)
                    }
                }
                // если токен аннулирован, то отдаем ошибку
                if user.Status == "revoked" {
                    return cfg.Fail(c,
                        errors.New("The token was revoked"))
                }
                // текущее значение числа переведенных символов из БД
                translated = uint64(user.Chars)
                limit = uint64(user.Max)
                expires = user.ExpiresAt
            } else {

                claims, err := jwtutils.ExtractClaims(
                    token, cfg.JWT.SigningKey.([]byte))
                
                if err != nil {
                    return cfg.Fail(c, err)
                }

                //expires = time.Unix(int64(claims["exp"].(float64)), 0)
                var exp int64
                switch v := claims["exp"].(type) {
                case float64:
                    exp = int64(v)
                case json.Number:
                    exp, _ = v.Int64()
                }
                expires = time.Unix(exp, 0)

                if  max, ok := claims["Max"]; ok {
                    switch v := max.(type) {
                    case float64:
                        limit = uint64(v)
                    case json.Number:
                        lm, _ := v.Int64()
                        limit = uint64(lm)
                    case string:
                        limit, _ = strconv.ParseUint(v, 0, 64)
                    default:
                        limit = 0    
                    }
                }
            }

            c.Locals("expires", expires)
            c.Locals("max", limit)
            c.Locals("translated", translated)
            // и после успешной проверки - в кэш
            return handler(c)
        }
        // finish response
        return nil
    }
}
