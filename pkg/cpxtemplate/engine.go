package cpxtemplate

import (
	"log"
	"os"

	"github.com/Shopify/go-lua"
)

func Validate() {
	l := lua.NewState()
	lua.OpenLibraries(l)
	if err := lua.DoString(l, "n = 0"); err != nil {
		log.Fatalln(err)
	}
	l.Dump(os.Stdout)
}
