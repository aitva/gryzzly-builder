package main

import "log"
import "os"

func main() {
	l := log.New(os.Stdout, "MAIN  ", log.LstdFlags|log.Lshortfile)
	l.Println("Hello World!")
}
