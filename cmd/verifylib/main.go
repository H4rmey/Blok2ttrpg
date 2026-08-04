// Command verifylib is a throwaway verification tool that loads the built-in
// content library and prints every package and ability it can parse. It is used
// to confirm the Blok2Simplified library files import cleanly.
package main

import (
	"fmt"
	"os"

	"github.com/harmey/blok2ttrpg-v5/internal/premade"
)

func main() {
	lib := premade.New("library")

	abs, err := lib.ListAbilities()
	if err != nil {
		fmt.Println("error listing abilities:", err)
		os.Exit(1)
	}
	fmt.Printf("Abilities (%d):\n", len(abs))
	for _, a := range abs {
		fmt.Printf("  - %s (id=%s, type=%s)\n", a.Name, a.ID, a.Type)
	}

	pkgs, err := lib.ListPackages()
	if err != nil {
		fmt.Println("error listing packages:", err)
		os.Exit(1)
	}
	fmt.Printf("\nPackages (%d):\n", len(pkgs))
	for _, p := range pkgs {
		fmt.Printf("  - [%s] %s (id=%s, shifts=%d, abilities=%d)\n",
			p.Category, p.Name, p.ID, len(p.Shifts), len(p.Abilities))
	}
}
