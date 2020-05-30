package main

import (
	"os"
	"fmt"
	"time"
	"flag"
	"strings"
	"syscall"
	"os/exec"
	"io/ioutil"
)

var (
	_unshare = flag.Bool("unshare", false, "mount CLONE_NEWNS")
	_fork    = flag.Bool("fork",    false, "fork unshare")
)

func test(n int, path, info string) {
	fmt.Printf("        test %d [%s] -> ls %s\n", n, path, info)

	files, _ := ioutil.ReadDir(path)
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

	test(1, "/tmp/vinicio", "")

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
	test(2, "/tmp/vinicio", "")

	fmt.Println("        mount /tmp/vinicio")
	fmt.Println("        err mount -> ",
		syscall.Mount("none", "/tmp/vinicio", "ramfs",
			syscall.MS_NOATIME|syscall.MS_NODEV|syscall.MS_NOSUID, ""))

	fmt.Println("        err mount -> ",
		syscall.Mount("none", "/tmp/vinicio", "ramfs",
			syscall.MS_REMOUNT|syscall.MS_UNBINDABLE, ""))


	test(3, "/tmp/vinicio", "")

	go func() { test(4, "/tmp/vinicio", "goroutine") }()

	time.Sleep(1 * time.Second)

	fmt.Println("        mkdir /tmp/vinicio/test1")
	fmt.Println("        err mkdir -> ", os.Mkdir("/tmp/vinicio/test1", 0755))

	fmt.Println("        mkdir /tmp/vinicio/test2")
	fmt.Println("        err mkdir -> ", os.Mkdir("/tmp/vinicio/test2", 0755))

	fmt.Println("        mkdir /tmp/vinicio/test3")
	fmt.Println("        err mkdir -> ", os.Mkdir("/tmp/vinicio/test3", 0755))

	fmt.Println()

	test(5, "/tmp/vinicio", "with data")

	go func() { test(6, "/tmp/vinicio", "with data goroutine") }()

	time.Sleep(1 * time.Second)

	// exit control
	c := make(chan bool)

	go func() {

		// exit control
		x := make(chan bool)
		go func() {
			test(7, "/tmp/vinicio", "with data goroutine")
			fakeRequest()
			x <- true
		}()
		<-x

		test(12, "/tmp/vinicio/test", "")
		c <- true
	}()

	<-c  // 2

	test(13, "/tmp/vinicio/test", "")
}

func fakeRequest() {
	fmt.Println("     FAKE REQUEST", os.Getpid())

	script := `

import os

files = os.listdir('/tmp/vinicio')

print("python -> ls /tmp/vinicio")
for f in files:
    print(' - {}'.format(f))

files = os.listdir('/tmp/vinicio/test')

print("python -> ls /tmp/vinicio/test")
for f in files:
    print(' - {}'.format(f))

print("python -> Goodbye, World!")

	`

	fmt.Println("     END ENGINE")
	fmt.Println()

	var e exec.Cmd
	e.Path = "/proc/self/exe"
	e.Args = []string{"init", "engine"}

	e.Stdout = os.Stdout
	e.Stderr = os.Stderr
	e.Stdin  = strings.NewReader(script)

	e.Run()

	test(11, "/tmp/vinicio/test", "")
}

func engine() {
	fmt.Println("     ENGINE", os.Getpid())
	fmt.Println()

	if err := syscall.Unshare(syscall.CLONE_NEWNS); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("  mkdir /tmp/vinicio/test")
	fmt.Println("  err mkdir -> ", os.Mkdir("/tmp/vinicio/test", 0755))
	test(8, "/tmp/vinicio", "")

	test(9, "/tmp/vinicio/test", "")

	fmt.Println("  mount /tmp/vinicio/test")
	fmt.Println("  err mount -> ",
		syscall.Mount("none", "/tmp/vinicio/test", "ramfs",
			syscall.MS_NOATIME|syscall.MS_NODEV|syscall.MS_NOSUID, ""))

	fmt.Println("  err mount -> ",
		syscall.Mount("none", "/tmp/vinicio/test", "ramfs",
			syscall.MS_REMOUNT|syscall.MS_UNBINDABLE, ""))

	fmt.Println("  mkdir /tmp/vinicio/test/vinicio")
	fmt.Println("  err mkdir -> ", os.Mkdir("/tmp/vinicio/test/vinicio", 0755))

	test(10, "/tmp/vinicio/test", "")

	var e exec.Cmd

	e.Path = "/usr/bin/python3.7"
	e.Args = []string{"self", "-B", "-"}

	e.Stdout = os.Stdout
	e.Stderr = os.Stderr
	e.Stdin  = os.Stdin

	e.Run()
	fmt.Println()

	// move output here
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

	handler = make(map[string] func(), 3)

	handler["loader"]  = loader
	handler["unshare"] = unshare
	handler["engine"]  = engine

	if len(args) > 0 {
		dispatch(args[0])
		os.Exit(0)
	}

	initLoader()
}
