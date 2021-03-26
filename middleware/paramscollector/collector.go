package paramscollector

import (
	//"unicode/utf8"
	"strings"

	"translator/utils"

	//"github.com/fatih/structs"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
)

type Params struct {
	Do     string
	As     string
	Source string
	Lang   string
	Out    string
	Merge  bool
}

func New() fiber.Handler {
	// Return new handler
	return func(c *fiber.Ctx) error {

		params := Params{}
		body := Params{}

		switch c.Method() {
		case "GET":
			_ = c.QueryParser(&params)

		case "POST":
			_ = c.BodyParser(&body)
			_ = c.QueryParser(&params)

			copier.CopyWithOption(&params, &body,
				copier.Option{IgnoreEmpty: true},
			)
		}

		if !utils.ContainsString(
			[]string{"json", "file", "raw"}, params.As) {
			params.As = "json"
		}

		c.Locals("do", c.Params("do"))
		c.Locals("source", strings.TrimRight(params.Source, " \n\t\r\v"))
		c.Locals("lang", strings.TrimSpace(params.Lang))
		c.Locals("as", strings.TrimSpace(params.As))
		c.Locals("out", params.Out)
		c.Locals("merge", params.Merge)

		return c.Next()
	}
}
