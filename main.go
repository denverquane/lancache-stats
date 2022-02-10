package main

import (
	"errors"
	"github.com/denverquane/lancache-stats/pkg/lancache"
	"github.com/gin-gonic/gin"
	"github.com/hpcloud/tail"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
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
	r.GET("/logs", func(c *gin.Context) {
		fs, err := getFiltersFromQuery(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, HttpError{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, coll.Filter(fs))
	})

	r.GET("/summary", func(c *gin.Context) {
		fs, err := getFiltersFromQuery(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, HttpError{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, coll.Filter(fs).Summarize())
	})

	defer func() {
		if err := t.Stop(); err != nil {
			log.Println(err)
		}
		t.Cleanup()
	}()
	// TODO endpoint to associate canonical names with IP addresses
	// TODO load canonical name associations from file
	// TODO endpoint for canonical name associations also writes to file

	// TODO endpoint to reset all data (clear cached/offsets, also clear name association file)
	return r.Run(":" + port)
}

func getFiltersFromQuery(c *gin.Context) ([]lancache.LogPredicateFn, error) {
	fs := make([]lancache.LogPredicateFn, 0)
	startTime, endTime := time.Time{}, time.Now()
	minSize, maxSize := uint64(0), uint64(math.MaxUint64)
	if c.Query("starttime") != "" {
		s, err := parseTime(c.Query("starttime"))
		if err != nil {
			return fs, errors.New("Issue parsing starttime as Unix timestamp: " + err.Error())
		}
		startTime = s
	}
	if c.Query("endtime") != "" {
		e, err := parseTime(c.Query("endtime"))
		if err != nil {
			return fs, errors.New("Issue parsing endtime as Unix timestamp: " + err.Error())
		}
		endTime = e
	}
	if c.Query("minsize") != "" {
		min, err := strconv.ParseUint(c.Query("minsize"), 10, 64)
		if err != nil {
			return fs, errors.New("Issue parsing minsize: " + err.Error())
		}
		minSize = min
	}
	if c.Query("maxsize") != "" {
		max, err := strconv.ParseUint(c.Query("maxsize"), 10, 64)
		if err != nil {
			return fs, errors.New("Issue parsing maxsize: " + err.Error())
		}
		maxSize = max
	}

	return []lancache.LogPredicateFn{
		lancache.ClientPredFn(c.Query("client")),
		lancache.SrcPredFn(c.Query("src")),
		lancache.DestPredFn(c.Query("dest")),
		lancache.TimeRangePredFn(startTime, endTime),
		lancache.SizeRangePredFn(minSize, maxSize),
	}, nil
}

func parseTime(t string) (time.Time, error) {
	i, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(i, 0), nil
}

type HttpError struct {
	Error string `json:"error"`
}
