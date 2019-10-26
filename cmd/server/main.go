// A Cannabis repository, useful for storing and searching various Cannabis strains.
package main

import (
	"fmt"
	"github.com/swtch1/too_many_strains/cmd/server/cli"
	"github.com/swtch1/too_many_strains/pkg"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	// appName should be populated at build time by build ldflags
	appName = "unset: please file an issue"
	// buildVersion is the application version and should be populated at build time by build ldflags
	// this default message should be overwritten
	buildVersion = "unset: please file an issue"
)

func main() {
	cli.Init(appName, buildVersion)
	tms.InitLogger(os.Stderr, cli.LogLevel, cli.LogFormat, cli.PrettyPrintJsonLogs)

	db := tms.DBServer{
		Username: "root",
		Password: "password",
		Name:     "so_many_strains",
	}
	if err := db.Open(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//srv := tms.Server{
	//	Port: cli.Port,
	//}

	//go HandleInterrupt()
	//log.Infof("starting server on port %d", cli.Port)
	//log.Fatal(srv.ListenAndServe())
}

// HandleInterrupt will immediately terminate the server if it detects an interrupt signal.
func HandleInterrupt() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("interrupt: stopping server...")
		os.Exit(1)
	}()
}
