package charslimiter

import (
    
    "bytes"
    "encoding/gob"
    "encoding/hex"
    "sync"
    "time"

    "translator/logger"
    "translator/models"
    "translator/utils/gormutil"

    "github.com/patrickmn/go-cache"
)

type CacheMap = map[string]cache.Item


type Cache struct {
    *cache.Cache
    config CacheConfig
    done   chan struct{}
    m      sync.RWMutex
}

var DefaultCacheConfig = CacheConfig{
    Intervals: map[string]time.Duration{
        "expiration": 60 * time.Minute,
        "cleanup":    10 * time.Minute,
        "sync":       60 * time.Minute,
    },
}

type CacheConfig struct {
    DB        *gormutil.DataBase
    Log       logger.LeveledLogger
    Intervals map[string]time.Duration
}

func (c *Cache) SetLogger(log logger.LeveledLogger) {
    c.config.Log = log
}

func (c *Cache) Log() logger.LeveledLogger {
    return c.config.Log
}

func NewCache(cfg CacheConfig) (c *Cache) {

    expiration := cfg.Intervals["expiration"]
    cleanupInterval := cfg.Intervals["cleanup"]
    syncInterval := cfg.Intervals["sync"]

    c = &Cache{
        Cache:  cache.New(expiration, cleanupInterval),
        config: cfg,
        done:   make(chan struct{}),
    }

    if c.config.Log == nil {
        c.config.Log = logger.NewLogrus()
    }

    if c.config.DB != nil {
        go c.autosync(syncInterval)
    }
    return
}

func GetCache(cfg CacheConfig) (c *Cache) {

    c = &Cache{
        config: cfg,
        done:   make(chan struct{}),
    }

    if c.config.Log == nil {
        c.config.Log = logger.NewLogrus()
    }

    items, err := c.Load()
    if err != nil {
        c.Log().Errorf("[CHLIMITER] cache:load:err:%v\n", err)
    }

    expiration := cfg.Intervals["expiration"]
    cleanupInterval := cfg.Intervals["cleanup"]
    syncInterval := cfg.Intervals["sync"]

    if len(items) != 0 {
        loadedCache := cache.NewFrom(expiration, cleanupInterval, items)
        c.Cache = loadedCache
        c.Log().Debugf("[CHLIMITER] cache:loaded:count:%d| %#v\n",
            c.ItemCount(), c.Cache)

    } else {
        c.Cache = cache.New(expiration, cleanupInterval)
        c.Log().Debugf("[CHLIMITER] cache:new:err:%v| %#v\n", err, c.Cache)
    }

    if c.config.DB != nil {
        go c.autosync(syncInterval)
    }

    return
}

func (c *Cache) Remove() (affected int64, err error) {

    if c.config.DB == nil {
        err = ErrDatabaseNotInitialized
        return
    }

    model := &models.UnauthorizedUser{}
    c.m.Lock()
    defer c.m.Unlock()
    affected, err = c.config.DB.Delete(model, "id = ?", 1)

    return
}

func (c *Cache) Save() (affected int64, err error) {

    if c.config.DB == nil {
        err = ErrDatabaseNotInitialized
        return
    }

    model := &models.UnauthorizedUser{ID: 1}
    items := c.serialize(c.Items())
    cnt := len(items)
    model.Cache = make([]byte, cnt+2, cnt+2)
    copy(model.Cache, `\x`)
    copy(model.Cache[2:], items)

    //model.Cache = []byte(`\x`)
    //model.Cache = append(model.Cache, c.serialize(c.Items())...)
    c.m.Lock()
    defer c.m.Unlock()
    affected, err = c.config.DB.InsertOrUpdate(&model, "id", []string{"cache", "updated_at"})

    return
}

func (c *Cache) Load() (out map[string]cache.Item, err error) {

    if c.config.DB == nil {
        err = ErrDatabaseNotInitialized
        return
    }

    model := &models.UnauthorizedUser{}

    affected, err := c.config.DB.Select(model, "id = ?", 1)
    if err != nil {
        return
    }

    if affected != 0 {
        items, err := c.deserialize(model.Cache)
        if err == nil {
            // отфильтровать полученный набор, удалив все expired данные
            out = make(map[string]cache.Item)
            now := time.Now().UnixNano()
            for k, v := range items {
                if v.Expiration > 0 {
                    if now > v.Expiration {
                        continue
                    }
                }
                out[k] = v
            }
            items = nil
        }
    }

    return
}

func (c *Cache) Sync() (err error) {

    if c.ItemCount() > 0 {
        affected, err := c.Save()

        if err != nil {
            c.Log().Errorf("[CHLIMITER] cache:sync:affected:%d:err:%v\n", affected, err)
        } else {
            c.Log().Debugf("[CHLIMITER] cache:sync:count:%d", c.ItemCount())
            //for k, v := range c.Items() {
            //  fmt.Printf("%s %d %d\n", k, v.Object.(uint64),
            //      v.Expiration,
            //  )
            //}
        }
    } else {
        affected, err := c.Remove()
        if err != nil {
            c.Log().Errorf("[CHLIMITER] cache:delete:affected:%d:err:%v\n", affected, err)
        } else {
            c.Log().Debugf("[CHLIMITER] cache:delete:affected:%d", affected)
        }
    }

    return
}

func (c *Cache) serialize(items map[string]cache.Item) (out []byte) {

    var src bytes.Buffer

    enc := gob.NewEncoder(&src)
    enc.Encode(items)
    encodedSize := hex.EncodedLen(src.Len())
    out = make([]byte, encodedSize)
    hex.Encode(out, src.Bytes())
    c.Log().Debugf("[CHLIMITER] cache:serialize:size:[src:%d] [enc:%d]\n",
        src.Len(),
        encodedSize,
    )
    return
}

func (c *Cache) deserialize(data []byte) (out map[string]cache.Item, err error) {

    var src bytes.Buffer

    out = make(map[string]cache.Item)
    decodedSize := hex.DecodedLen(len(data) - 2)
    dst := make([]byte, decodedSize)
    n, err := hex.Decode(dst, data[2:])
    if err != nil {
        c.Log().Errorf("[CHLIMITER] cache:deserialize:err: %v\n", err)
        return
    }

    src.Write(dst[:n])
    dec := gob.NewDecoder(&src)
    dec.Decode(&out)

    c.Log().Debugf("[CHLIMITER] cache:deserialize:size: [src:%d] [dec:%d] \n",
        len(data),
        decodedSize,
    )

    return
}

func (c *Cache) Close() {
    c.done <- struct{}{}
}

func (c *Cache) autosync(syncInterval time.Duration) {
    ticker := time.NewTicker(syncInterval)
    defer ticker.Stop()

    for {
        select {
        case <-c.done:
            return
        case <-ticker.C:
            c.Sync()
        }
    }
}