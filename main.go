package main

import (
	"github.com/denverquane/lancache-stats/pkg/lancache"
	"github.com/gin-gonic/gin"
	"github.com/hpcloud/tail"
	"log"
	"net/http"
	"os"
	"path"
)

const (
	DefaultPort    = "5000"
	DefaultLogPath = "/data/logs"
)

func main() {
	port := os.Getenv("LANCACHE_STATS_PORT")
	logPath := os.Getenv("LANCACHE_STATS_LOG_PATH")

	if logPath == "" {
		logPath = DefaultLogPath
	}
	if port == "" {
		port = DefaultPort
	}

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		log.Fatal(err)
	}

	log.Println(startServerAndTail(port, logPath))
}

func startServerAndTail(port, logpath string) error {
	coll := lancache.NewLogCollection()

	log.Println("Tailing access.log file for new changes")
	t, err := tail.TailFile(path.Join(logpath, "access.log"), tail.Config{Follow: true, MustExist: true})
	if err != nil {
		log.Fatal(err)
	}

	go lancache.ProcessTailAccessFile(t, &coll)

	r := gin.Default()
	r.GET("/summary", func(c *gin.Context) {
		c.JSON(http.StatusOK, coll.Summarize())
	})

	defer func() {
		t.Stop()
		t.Cleanup()
	}()
	// TODO endpoint to associate canonical names with IP addresses
	// TODO load canonical name associations from file
	// TODO endpoint for canonical name associations also writes to file

	// TODO endpoint to reset all data (clear cached/offsets, also clear name association file)
	return r.Run(":" + port)
}

type HttpError struct {
	Error string `json:"error"`
}
