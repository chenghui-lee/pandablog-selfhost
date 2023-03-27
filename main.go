package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/chenghui-lee/pandablog-selfhost/app"
	"github.com/chenghui-lee/pandablog-selfhost/app/lib/timezone"
)

func init() {
	// Verbose logging with file name and line number.
	log.SetFlags(log.Lshortfile)
	// Set the time zone.
	timezone.Set()
}

func main() {
	handler, err := app.Boot()
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Start the web server.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Web server running on port:", port)
	log.Fatalln(http.ListenAndServe(":"+port, handler))
}
