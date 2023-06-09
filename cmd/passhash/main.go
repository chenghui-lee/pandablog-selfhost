package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/chenghui-lee/pandablog-selfhost/app/lib/passhash"
	"github.com/chenghui-lee/pandablog-selfhost/app/lib/timezone"
)

func init() {
	// Verbose logging with file name and line number.
	log.SetFlags(log.Lshortfile)
	// Set the time zone.
	timezone.Set()
}

func main() {
	var pass string
	if len(os.Args) >= 2 {
		pass = os.Args[1]
	} else {
		fmt.Print("Please input your desired password: ")
		p, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			log.Fatalf("Unable to read password: %v", err)
		}
		pass = string(p)
	}

	// Generate a new private key.
	s, err := passhash.HashString(pass)
	if err != nil {
		log.Fatalln(err.Error())
	}

	sss := base64.StdEncoding.EncodeToString([]byte(s))
	fmt.Printf("PBB_PASSWORD_HASH=%v\n", sss)
}
