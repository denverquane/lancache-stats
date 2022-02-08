package main

import (
	"flag"
	"fmt"
	"github.com/denverquane/lancache-stats/pkg/lancache"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

func main() {
	port := flag.Int("port", 5000, "Port on which to run HTTP stats server")
	logPath := flag.String("log-path", "", "Log path")
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
	log.Println(startServer(*port, *logPath))
}

func startServer(port int, path string) error {
	currentStats := lancache.NewLogStatistics()
	var logBytes int64

	r := gin.Default()
	r.GET("/stats", func(c *gin.Context) {
		size, err := lancache.LogFileSize(path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		} else if size > logBytes {
			log.Printf("Log file has more bytes; starting processing from offset %d\n", logBytes)
			logBytes = lancache.ParseFileFromOffset(path, &currentStats, logBytes)
		}
		c.JSON(http.StatusCreated, currentStats)
	})
	// TODO endpoint to associate canonical names with IP addresses
	// TODO load canonical name associations from file
	// TODO endpoint for canonical name associations also writes to file

	// TODO endpoint to reset all data (clear cached/offsets, also clear name association file)
	return r.Run(fmt.Sprintf(":%d", port))
}
