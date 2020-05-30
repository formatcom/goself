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

	/*
	if len(os.Args) > 1 {

		if os.Getenv("_") != fmt.Sprintf("%d", os.Getppid()) {
			os.Exit(1)
		}

		if os.Args[0] == "unshare" {
			fmt.Println("[unshare] test 6 -> ls /tmp/vinicio/")
			files, _ := ioutil.ReadDir("/tmp/vinicio")
			for _, file := range files {
				fmt.Println("test 6 - ", file.Name())
			}

			go func() {
				fmt.Println("[unshare gorutine] test 7 -> ls /tmp/vinicio/")

				files, _ := ioutil.ReadDir("/tmp/vinicio")
				for _, file := range files {
					fmt.Println("test 7 - ", file.Name())
				}
			}()
			time.Sleep(5 * time.Minute)
			os.Exit(0)
		}

		// Create session
		// Evitar que muera el hijo si matan el padre
		// syscall.Setsid()

		fmt.Println("parent pid:", os.Getppid())
		fmt.Println(os.Environ())

		fmt.Println("[local namespace] test 1 -> mkdir /tmp/vinicio/test1")
		fmt.Println(os.Mkdir("/tmp/vinicio/test1", 0755))

		// Crear un nuevo mount namespace
		if err := syscall.Unshare(syscall.CLONE_NEWNS); err != nil {
			fmt.Println("unshare", err)
		}

		fmt.Println("wait mount")

		time.Sleep(15 * time.Second)

		fmt.Println("mount /tmp/vinicio")

		err := syscall.Mount("none", "/tmp/vinicio", "ramfs",
				syscall.MS_NOATIME|syscall.MS_NODEV|syscall.MS_NOSUID, "")

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = syscall.Mount("none", "/tmp/vinicio", "ramfs",
				syscall.MS_REMOUNT|syscall.MS_UNBINDABLE, "")

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("[new namespace] test 2 -> mkdir /tmp/vinicio/test1")
		fmt.Println(os.Mkdir("/tmp/vinicio/test1", 0755))
		fmt.Println("[new namespace] test 3 -> mkdir /tmp/vinicio/test2")
		fmt.Println(os.Mkdir("/tmp/vinicio/test2", 0755))

		fmt.Println("[new namespace] test 4 -> ls /tmp/vinicio/")

		files, _ := ioutil.ReadDir("/tmp/vinicio")
		for _, file := range files {
			fmt.Println("test 4 - ", file.Name())
		}

		go func() {
			fmt.Println("[gorutine] test 5 -> ls /tmp/vinicio/")

			files, _ := ioutil.ReadDir("/tmp/vinicio")
			for _, file := range files {
				fmt.Println("test 5 - ", file.Name())
			}
		}()


		fmt.Println(syscall.ForkExec(
			"/proc/self/exe",
			[]string{"unshare", "loader"},
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
		))

		/*
		r1, _, err1 := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
		if err1 != 0 {
			os.Exit(1)
		}

		// parent
		if r1 != 0 {
		}

		// child
		if r1 == 0 {
			/*
			pid := os.Getpid()
			syscall.RawSyscall(syscall.SYS_SETPGID, 0, 0, 0)
			syscall.RawSyscall(syscall.SYS_IOCTL, 0, uintptr(syscall.TIOCSPGRP), uintptr(unsafe.Pointer(&pid)))

			fmt.Println("hola mundo")

			go func() {
				fmt.Println("hola mundo 2")
			}()

			//go func() {
				fmt.Fprintln(os.Stdout, "[new namespace] test 6 -> ls /tmp/vinicio/")

				files, _ := ioutil.ReadDir("/tmp/vinicio")
				for _, file := range files {
					fmt.Println("test 6 - ", file.Name())
				}
			//}()
		}

		//if r1 != 0 { return 0 }

		/*
		if r1 == 0 {

		*/

	/*
		time.Sleep(5 * time.Minute)
		os.Exit(0)
	}


	fmt.Println("pid:", os.Getpid())

	pid, err := syscall.ForkExec(
		"/proc/self/exe",
		[]string{"initc", "loader"},
		&syscall.ProcAttr{
			Env:   []string{
						fmt.Sprintf("_=%d", os.Getpid()),
			},
			Files: []uintptr{
						0,
						os.Stdout.Fd(),
						os.Stderr.Fd(),
			},
			Sys:     &syscall.SysProcAttr{
						Unshareflags:  syscall.CLONE_NEWNS,
			},
		},
	)

	if err != nil {
		os.Exit(1)
	}

	fmt.Println("child", pid)

	process := os.Process{Pid: pid}
	process.Wait()
	*/
}
