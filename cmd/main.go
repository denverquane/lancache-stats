package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	flag.Parse()
	port := flag.Int("port", 5000, "Port on which to run HTTP stats server")
	logPath := flag.String("log-path", "", "Log path")
	if logPath == nil {
		log.Fatal("Nil logpath ptr")
	}

	if _, err := os.Stat(*logPath); os.IsNotExist(err) {
		log.Fatal(err)
	}
	log.Println(port)
}
