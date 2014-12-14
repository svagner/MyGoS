package main

import (
	"./config"
	"./interface/web"
	"flag"
	"github.com/yookoala/realpath/realpath"
	"log"
	"os"
	"runtime"
	"syscall"
)

func daemonize(nochdir, noclose bool) (*os.Process, error) {
	daemonizeState := os.Getenv("_GOLANG_DAEMONIZE_FLAG")
	switch daemonizeState {
	case "":
		syscall.Umask(0)
		os.Setenv("_GOLANG_DAEMONIZE_FLAG", "1")
	case "1":
		syscall.Setsid()
		os.Setenv("_GOLANG_DAEMONIZE_FLAG", "2")
	case "2":
		os.Setenv("_GOLANG_DAEMONIZE_FLAG", "")
		return nil, nil
	}

	var attrs os.ProcAttr

	if !nochdir {
		attrs.Dir = "/"
	}

	if noclose {
		attrs.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	} else {
		f, err := os.Open("/dev/null")
		if err != nil {
			return nil, err
		}
		attrs.Files = []*os.File{f, f, f}
	}

	exe, err := realpath.Realpath(os.Args[0])
	if err != nil {
		return nil, err
	}

	p, err := os.StartProcess(exe, os.Args, &attrs)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func main() {
	var cfgFile = flag.String("config", "config.ini", "Configuration file (ini)")
	flag.Parse()
	var config config.Config
	if err := config.ParseConfig(*cfgFile); err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("Init config file ", *cfgFile)
	if runtime.GOOS == "darvin" {
		config.Global.Type = "standalone"
		log.Println("Demon will run as standalone daemon type...")
	}
	if config.Global.Type == "daemon" {
		_, err := daemonize(false, false)
		if err != nil {
			log.Println(err.Error())
		}
	}
	web.Start(config.Http)
}
