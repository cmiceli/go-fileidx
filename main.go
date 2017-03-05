package main

import (
	"flag"
	"fmt"
	"os"
)

import (
	"bitbucket.org/gaymish/file_indexer/db"
	"bitbucket.org/gaymish/file_indexer/walker"
	"github.com/golang/glog"
)

type modeConst int

const (
	SETUP modeConst = iota
	ADD
	CHECK
)

func main() {
	//parse cli
	modeFlag := flag.String("mode", "none", "modes: {setup, add, check}")
	src := flag.String("dir", ".", "Source directory")
	dbFilename := flag.String("db", "photos.db", "sqlite db")
	flag.Parse()

	var mode modeConst

	switch *modeFlag {
	case "setup":
		mode = SETUP
	case "add":
		mode = ADD
	case "check":
		mode = CHECK
	default:
		flag.Usage()
		os.Exit(1)
	}

	glog.Info("Setting up the database")
	db, err := db.NewSQLiteDB(*dbFilename)
	if err != nil {
		glog.Fatal(err)
	}

	if mode == SETUP {
		glog.Info("Running Setup")
		err := db.Setup()
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	files := make(chan walker.FileInfo)
	errchan := make(chan error)
	errCounter := 0
	successCounter := 0

	glog.Info("creating walker")
	w := walker.NewWalker(files, errchan)
	glog.Info("Walking ", *src)
	go w.Walk(*src)

FileLooper:
	for {
		select {
		case d, ok := <-files:
			if !ok {
				break FileLooper
			}
			glog.Info("Found ", d.Fileobj.Name())
			switch mode {
			case ADD:
				glog.Info("Adding ", d.Fileobj.Name())
				b, err := db.CheckExistence(d.Fileobj.Name(), d.Hash)
				if err != nil {
					glog.Errorf("Error checking file: %v", err)
					errCounter++
					continue
				}
				if b {
					glog.Info("Already exists: ", d.Fileobj.Name())
					continue
				}
				err = db.Add(d.Fileobj.Name(), d.Hash, d.Filename)
				if err != nil {
					glog.Errorf("Error adding file %v", err)
					errCounter++
					continue
				}
				successCounter++
			case CHECK:
				glog.Info("Checking ", d.Fileobj.Name())
				b, err := db.CheckExistence(d.Fileobj.Name(), d.Hash)
				if err != nil {
					glog.Errorf("Error checking file %v", err)
					errCounter++
					continue
				}
				if !b {
					glog.Error("Not Found ", d.Fileobj.Name())
				}
				successCounter++
			}
		case err, ok := <-errchan:
			if !ok {
				break FileLooper
			}
			if err != nil {
				errCounter++
				continue
			}
		}
	}

	fmt.Println("Number of successes: ", successCounter)
	fmt.Println("Number of errors: ", errCounter)

}
