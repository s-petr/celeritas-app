package main

import (
	"log"
	"myapp/data"
	"myapp/handlers"
	"myapp/middleware"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/s-petr/celeritas"
)

type application struct {
	App        *celeritas.Celeritas
	Handlers   *handlers.Handlers
	Models     data.Models
	Middleware *middleware.Middleware
	wg         sync.WaitGroup
}

func main() {
	c := initApplication()
	go c.listenForShutDown()
	log.Fatal(c.App.ListenAndServe())
}

func (a *application) shutdown() {
	// put any cleanup tasks here
	a.wg.Wait()
	a.App.InfoLog.Println("Starting cleanup tasks...")
	time.Sleep(3 * time.Second)
	a.App.InfoLog.Println("Finished cleanup tasks. Exiting...")
}

func (a *application) listenForShutDown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit

	a.App.InfoLog.Println("Received signal", s.String())
	a.shutdown()

	os.Exit(0)
}
