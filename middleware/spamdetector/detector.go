package spamdetector

import (
    //"fmt"
    "regexp"
    "sort"
    "strings"
    "unicode/utf8"

    "github.com/gofiber/fiber/v2"
    "github.com/xtgo/set"
)

var NONLETTERS = regexp.MustCompile(`[[:digit:][:punct:][:cntrl:][:space:]"#$%&'()*+,\-./:;<=>?@[\\\]^_{|}~]`)
var LATIN_OR_CYRILLIC = regexp.MustCompile(`\p{Latin}|\p{Cyrillic}`) // `\p{Latin}|\p{Cyrillic}`
var LATIN = regexp.MustCompile(`\p{Latin}`)
var CYRILLIC = regexp.MustCompile(`\p{Cyrillic}`)

func IsLatin(value string) bool    { return LATIN.MatchString(value) }
func IsCyrillic(value string) bool { return CYRILLIC.MatchString(value) }

type Config struct {
    LD       float64 //lexical diversity
    MaxLen   int
    Targets  []string
    Next     func(c *fiber.Ctx) bool
    Response func(c *fiber.Ctx) error
}

func Next(c *fiber.Ctx) bool { return false }

func Response(c *fiber.Ctx) error {
    message := c.Locals("message").(string)

    return c.Status(400).JSON(fiber.Map{
        "status":  fiber.StatusBadRequest,
        "message": message,
    })
}

var ConfigDefault = Config{
    LD:       0.5, // 50 %
    Targets:  make([]string, 0),
    MaxLen:   1000,
    Next:     Next,
    Response: Response,
}

func configDefault(config ...Config) Config {
    // Return default config if nothing provided
    if len(config) < 1 {
        return ConfigDefault
    }

    // Override default config
    cfg := config[0]

    // Set default values
    if cfg.MaxLen == 0 {
        cfg.MaxLen = ConfigDefault.MaxLen
    }
    
    if cfg.Next == nil {
        cfg.Next = ConfigDefault.Next
    }

    if cfg.Response == nil {
        cfg.Response = ConfigDefault.Response
    }

    if cfg.LD == 0 {
        cfg.LD = ConfigDefault.LD
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

        for _, target := range cfg.Targets {
            value := c.Locals(target).(string)
            runesCount := utf8.RuneCountInString(value)
            if cfg.MaxLen !=-1 && runesCount > cfg.MaxLen{
                value = string([]rune(value)[:cfg.MaxLen])
            }
            
            nonLettersCount := len(NONLETTERS.FindAllString(value, -1))
            // the string consists only of these chars
            if (runesCount == nonLettersCount) {
                message := "Looks too much like spam: lots of non-letter characters"
                c.Locals("message", message)
                return cfg.Response(c)
            }
            
            chars := sort.StringSlice(LATIN_OR_CYRILLIC.FindAllString(value, -1))
            chars.Sort()
            countUniq := set.Uniq(chars)
            
            //latin := LATIN.FindAllString(value, -1)
            //chars := sort.StringSlice(latin)
            //chars.Sort()
            //latinCount := set.Uniq(chars)
            
            //cyrillic := CYRILLIC.FindAllString(value, -1)
            //chars = sort.StringSlice(cyrillic)
            //chars.Sort()
            //cyrillicCount := set.Uniq(chars)
            
            //if (runesCount > 1 && (latinCount == 1 || cyrillicCount == 1)){
            if (runesCount > 1 && countUniq == 1) {
                message := "Looks too much like spam: lots of the same ones characters"
                c.Locals("message", message)
                return cfg.Response(c)
            } 
            
            //if (level := LetterVariety(value); level < cfg.LD) {
            //    message = fmt.Sprintf("Looks too much like spam %.1f", level)
            //    c.Locals("message", message)
            //    return cfg.Response(c)
            //}
        }
        return c.Next()
    }
}

func LetterVariety(value string) float64 {
    // if it is Latin or Cyrillic, check the text for a variety of letters
    var chars sort.Interface
    var alphaLen int
    var level float64 = 100.0

    latin := IsLatin(value)
    cyrillic := IsCyrillic(value)

    if latin || cyrillic {
        value = strings.ToLower(value)

        if latin {
            alphaLen = 26
            chars = sort.StringSlice(LATIN.FindAllString(value, -1))
        }
        if cyrillic {
            alphaLen = 33
            chars = sort.StringSlice(CYRILLIC.FindAllString(value, -1))
        }

        sort.Sort(chars)
        uniqRunesCount := set.Uniq(chars) // Uniq returns the size of the set
        level = float64(uniqRunesCount) / float64(alphaLen)
        //fmt.Printf("%d / %d = %f| %#v %#v\n", uniqRunesCount, runesCount, coef, cfg.LD, coef < cfg.LD )
    }
    return level
}
