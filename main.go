package main

import (
	"os"
	"fmt"
	"time"
	"flag"
	"syscall"
	"io/ioutil"
)

var (
	_unshare = flag.Bool("unshare", false, "mount CLONE_NEWNS")
	_fork    = flag.Bool("fork",    false, "fork unshare")
)

func test(n int, info string) {
	fmt.Printf("        test %d [%s] -> ls /tmp/vinicio/\n", n, info)

	files, _ := ioutil.ReadDir("/tmp/vinicio")
	for _, file := range files {
		fmt.Printf("        test %d [%s] -> %s\n", n, info, file.Name())
	}
	fmt.Println()
}

func initLoader() {
	fmt.Println("INIT LOADER", os.Getpid())

	// detach parent
	// create session
	// syscall.Setsid()

	var args []string

	args = append(args, "init")

	var sys syscall.SysProcAttr
	if *_unshare {
		fmt.Println("     SET CLONE_NEWNS")
		sys.Unshareflags = syscall.CLONE_NEWNS

		args = append(args, "--unshare")
	}


	if *_fork {

		args = append(args, "--fork")
	}

	args = append(args, "loader")

	pid, err := syscall.ForkExec(
		"/proc/self/exe", args,
		&syscall.ProcAttr{
			Env:   []string{
						fmt.Sprintf("_=%d", os.Getpid()),
			},
			Files: []uintptr{
						0,
						os.Stdout.Fd(),
						os.Stderr.Fd(),
			},
			Sys:                   &sys,
		},
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	process := os.Process{Pid: pid}
	process.Wait()
}

func loader() {
	fmt.Println("     LOADER", os.Getpid())

	test(1, "")

	if *_fork == false {
		fmt.Println("     CALL FUNC")
		unshare()
	} else {
		fmt.Println("     FORK EXEC")
		pid, err := syscall.ForkExec(
			"/proc/self/exe",
			[]string{"init", "unshare"},
			&syscall.ProcAttr{
				Env:   []string{
							fmt.Sprintf("_=%d", os.Getpid()),
				},
				Files: []uintptr{
							0,
							os.Stdout.Fd(),
							os.Stderr.Fd(),
				},
			},
		)

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}


		process := os.Process{Pid: pid}
		process.Wait()
	}
}

func unshare() {
	fmt.Println("     UNSHARE", os.Getpid())

	fmt.Println("        mkdir /tmp/vinicio/test")
	fmt.Println("        err mkdir -> ", os.Mkdir("/tmp/vinicio/test", 0755))
	test(2, "")

	fmt.Println("        mount /tmp/vinicio")
	fmt.Println("        err mount -> ",
		syscall.Mount("none", "/tmp/vinicio", "ramfs",
			syscall.MS_NOATIME|syscall.MS_NODEV|syscall.MS_NOSUID, ""))

	fmt.Println("        err mount -> ",
		syscall.Mount("none", "/tmp/vinicio", "ramfs",
			syscall.MS_REMOUNT|syscall.MS_UNBINDABLE, ""))


	test(3, "")

	go func() { test(4, "goroutine") }()

	time.Sleep(1 * time.Second)

	fmt.Println("        mkdir /tmp/vinicio/test1")
	fmt.Println("        err mkdir -> ", os.Mkdir("/tmp/vinicio/test1", 0755))

	fmt.Println("        mkdir /tmp/vinicio/test2")
	fmt.Println("        err mkdir -> ", os.Mkdir("/tmp/vinicio/test2", 0755))

	fmt.Println("        mkdir /tmp/vinicio/test3")
	fmt.Println("        err mkdir -> ", os.Mkdir("/tmp/vinicio/test4", 0755))

	fmt.Println()

	test(5, "with data")

	go func() { test(6, "with data goroutine") }()

	time.Sleep(1 * time.Second)

	go func() {
		go func() {
			test(7, "with data goroutine")
		}()
	}()

	time.Sleep(1 * time.Second)
}


var handler map[string] func()


func dispatch(method string) {
	// logica inicial requerida

	// handler
	if _func, ok := handler[method]; ok {
		_func()
	} else {
		fmt.Println(method, "method not allowed")
	}
}


// REF: https://golang.org/pkg/syscall/#SysProcAttr
// REF: http://man7.org/linux/man-pages/man2/unshare.2.html
func main() {

	flag.Parse()
	args := flag.Args()

	fmt.Println()
	fmt.Println(os.Args)
	fmt.Println(args)

	handler = make(map[string] func(), 2)

	handler["loader"]  = loader
	handler["unshare"] = unshare

	if len(args) > 0 {
		dispatch(args[0])
		os.Exit(0)
	}

	initLoader()
}
