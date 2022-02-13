package cpxtemplate

import (
	"log"

	"github.com/Shopify/go-lua"
)

func Validate() {
	l := lua.NewState()
	lua.OpenLibraries(l)
	if err := lua.DoString(l, "n = 0"); err != nil {
		log.Fatalln(err)
	}
	log.Println(lua.Stack(l, 1))
}
