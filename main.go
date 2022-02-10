package main

import (
	"flag"
	"fmt"
	"github.com/denverquane/lancache-stats/pkg/lancache"
	"github.com/gin-gonic/gin"
	"github.com/hpcloud/tail"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
)

func main() {
	port := flag.Int("port", 5000, "Port on which to run HTTP stats server")
	logPath := flag.String("log-path", "/data/logs", "Log path")
	flag.Parse()
	if logPath == nil {
		log.Fatal("Nil logpath ptr")
	}
	if port == nil {
		log.Fatal("Nil port ptr")
	}

	if _, err := os.Stat(*logPath); os.IsNotExist(err) {
		log.Fatal(err)
	}

	log.Println(startServerAndTail(*port, *logPath))
}

func startServerAndTail(port int, logpath string) error {
	stats := lancache.NewLogStatistics()
	lock := sync.RWMutex{}

	log.Println("Tailing access.log file for new changes")
	t, err := tail.TailFile(path.Join(logpath, "access.log"), tail.Config{Follow: true, MustExist: true})
	if err != nil {
		log.Fatal(err)
	}

	go lancache.ProcessTailAccessFile(t, &stats, &lock)

	r := gin.Default()
	r.GET("/stats", func(c *gin.Context) {
		lock.RLock()
		c.JSON(http.StatusCreated, stats)
		lock.RUnlock()
	})
	defer func() {
		t.Stop()
		t.Cleanup()
	}()
	// TODO endpoint to associate canonical names with IP addresses
	// TODO load canonical name associations from file
	// TODO endpoint for canonical name associations also writes to file

	// TODO endpoint to reset all data (clear cached/offsets, also clear name association file)
	return r.Run(fmt.Sprintf(":%d", port))
}
