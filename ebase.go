//
// Ebase frame for daemon program
// Author Jonsen Yang
// Date 2013-07-05
// Copyright (c) 2013 ForEase Times Technology Co., Ltd. All rights reserved.
//
package ebase

import (
	"flag"
	"fmt"
	"github.com/forease/config"
	"os"
	"os/signal"
	"path"
	"reflect"
	"runtime"
	"syscall"
)

var (
	// 日志
	Log        *BaseLog
	Config     *config.Config
	SigHandler = make(map[string]interface{}) //
	G          = make(map[string]interface{}) //

	// 定义命令行参数
	verbose = flag.Bool("v", false, "Verbose output")
	help    = flag.Bool("h", false, "Show this help")
	chroot  = flag.Bool("w", false, "Setup enable chroot")
	cfgfile = flag.String("c", "", "Config file")
	workdir = flag.String("d", "", "Setup work dir")
	pidfile = flag.String("p", "", "Pid file")
)

func init() {
	flag.Parse()
	if *help {
		Help()
		return
	}

	if *verbose {
		fmt.Println("version:")
		os.Exit(0)
	}

	if *workdir != "" {
		fmt.Println("workdir: ", *workdir, os.Args)
		if err := syscall.Chdir(*workdir); err != nil {
			fmt.Printf("Can't change to work dir [%s]: %s\n", *workdir, err)
			os.Exit(1)
		}

		if *chroot {
			pwd, _ := os.Getwd()
			if err := syscall.Chroot(pwd); err != nil {
				fmt.Printf("Can't Chroot to %s: %s\n", *workdir, err)
				os.Exit(1)
			}
			fmt.Printf("I'll Chroot to %s !\n", pwd)
		}
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	Config = LoadConfig("")
	CreatePid()
	Log = defaultLog()

	SigHandler["sighup"] = func() {
		Log.Debug("reload config")
		Config = LoadConfig("")
	}

	//
	if ok, _ := Config.Bool("sys.signal", false); ok {
		go SignalHandle(SigHandler)
	}
}

func LoadConfig(configFile string) (cfg *config.Config) {
	var err error
	if configFile == "" {
		if *cfgfile != "" {
			configFile = *cfgfile
		} else {
			configFile = "/opt/etc/" + path.Base(os.Args[0]) + ".conf"
		}
	}
	if configFile == "" {
		return
	}

	cfg, err = config.NewConfig(configFile, 16)
	if err != nil {
		fmt.Println("read config file error: ", err)
		os.Exit(1)
	}

	return cfg
}

func defaultLog() (l *BaseLog) {

	logType, _ := Config.String("log.type", "consloe")
	logFile, _ := Config.String("log.file", "")
	logLevel, _ := Config.Int("log.level", 5)
	logFlag, _ := Config.Int("log.flag", 9)
	//logEnable, _ = Config.Bool("log.enable", true)
	opt := &LogOptions{Type: logType, File: logFile, Level: logLevel, Flag: logFlag}
	return NewLog(opt)
}

//
func SignalHandle(funcs map[string]interface{}) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGHUP)

	for {
		select {
		case s := <-ch:
			switch s {
			default:
			case syscall.SIGHUP:
				if f, ok := funcs["sighup"]; ok {
					if ff := reflect.ValueOf(f); ff.Kind() == reflect.Func {
						ff.Call(nil)
					}
				}
				break
			case syscall.SIGINT:
				if f, ok := funcs["sigint"]; ok {
					if ff := reflect.ValueOf(f); ff.Kind() == reflect.Func {
						ff.Call(nil)
					}
				}
				os.Exit(1)
			case syscall.SIGUSR1:
				if f, ok := funcs["sigusr1"]; ok {
					if ff := reflect.ValueOf(f); ff.Kind() == reflect.Func {
						ff.Call(nil)
					}
				}
			case syscall.SIGUSR2:
				if f, ok := funcs["sigusr2"]; ok {
					if ff := reflect.ValueOf(f); ff.Kind() == reflect.Func {
						ff.Call(nil)
					}
				}
			}
		}
	}
}

// create pid file
func CreatePid() {
	pid := os.Getpid()

	if pid < 1 {
		fmt.Println("Get pid err")
		os.Exit(1)
	}

	var pidf string
	if *pidfile != "" {
		pidf = *pidfile
	} else {
		pidf, _ = Config.String("sys.pid", "")
		if pidf == "" {
			pidf = "/var/run/" + path.Base(os.Args[0]) + ".pid"
		}
	}

	if pidf == "" {
		fmt.Println("pid file not setup")
		os.Exit(1)
	}

	f, err := os.OpenFile(pidf, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Open pid file err ", err)
		os.Exit(1)
	}

	f.WriteString(GetIntStr(pid))
	f.Close()
}

func Daemon(nochdir, noclose int) int {
	var ret, ret2 uintptr
	var err syscall.Errno

	darwin := runtime.GOOS == "darwin"

	// already a daemon
	if syscall.Getppid() == 1 {
		return 0
	}

	// fork off the parent process
	ret, ret2, err = syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
	if err != 0 {
		return -1
	}

	// failure
	if ret2 < 0 {
		os.Exit(-1)
	}

	// handle exception for darwin
	if darwin && ret2 == 1 {
		ret = 0
	}

	// if we got a good PID, then we call exit the parent process.
	if ret > 0 {
		os.Exit(0)
	}

	/* Change the file mode mask */
	_ = syscall.Umask(0)

	// create a new SID for the child process
	s_ret, s_errno := syscall.Setsid()
	if s_errno != nil {
		Log.Errorf("Error: syscall.Setsid errno: %d", s_errno)
	}
	if s_ret < 0 {
		return -1
	}

	if nochdir == 0 {
		os.Chdir("/")
	}

	if noclose == 0 {
		f, e := os.OpenFile("/dev/null", os.O_RDWR, 0)
		if e == nil {
			fd := f.Fd()
			syscall.Dup2(int(fd), int(os.Stdin.Fd()))
			syscall.Dup2(int(fd), int(os.Stdout.Fd()))
			syscall.Dup2(int(fd), int(os.Stderr.Fd()))
		}
	}

	return 0
}

func Help() {
	fmt.Printf(
		"\nUseage: %s [ Options ]\n\n"+
			"Options:\n"+
			"  -c Server config file [Default: etc/serverd.conf]\n"+
			"  -d Work dir [Default: publish]\n"+
			"  -h Display this mssage\n"+
			"  -p Pid file [Default: var/serverd.pid]\n"+
			"  -w Enable chroot to work dir [Required: -d ]\n\n"+
			"------------------------------------------------------\n\n"+
			"  Author:  16hot (im16hot@gmail.com) \n"+
			"  Company: Beijing ForEase Times Technology Co., Ltd.\n"+
			"  Website: http://www.forease.net\n"+
			"  MyBlog:  http://16hot.com\n"+
			"  Version: 1.0 Beta1\n\n"+
			"------------------------------------------------------\n\n",
		os.Args[0])

	os.Exit(0)
}
