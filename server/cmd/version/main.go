// Command version prints the application version for build and release tooling.
package main

import (
	"fmt"

	"jww/internal/buildinfo"
)

func main() {
	fmt.Println(buildinfo.Version)
}
