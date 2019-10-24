// Create or migrate the strains database.
package main

import (
	"fmt"
	"github.com/swtch1/too_many_strains/cmd/database-migration/cli"
)

var (
	// dbVersion is the desired version of the database
	dbVersion = 1
	version   = "v1"
)

func main() {
	cli.Init("migrate", version)
	fmt.Println(New())
}

// New does lots of cool stuff.
func New() string {
	return ""
}
