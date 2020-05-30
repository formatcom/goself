package main

import (
	"fmt"
	"os"
	"time"
	"syscall"
	"os/exec"
	"io/ioutil"
)


// REF: https://golang.org/pkg/syscall/#SysProcAttr
// REF: http://man7.org/linux/man-pages/man2/unshare.2.html
func main() {

	fmt.Println("LOADER", os.Getpid())
	fmt.Println(os.Args)

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

		time.Sleep(5 * time.Minute)
		os.Exit(0)
	}


	fmt.Println("pid:", os.Getpid())

	cmd := exec.Cmd{
		Path: os.Args[0],
		Args: []string{"initc", "loader"},
		Env:  []string{
			 fmt.Sprintf("_=%d", os.Getpid()),
		},
		SysProcAttr: &syscall.SysProcAttr{Unshareflags: syscall.CLONE_NEWNS},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	fmt.Println(cmd.Run())
	/*
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

	c := make(chan error)

	go func() {
		state, err := process.Wait()
		if err == nil {
			fmt.Println("pid",    state.Pid(),
				    "exited", state.Exited(),
				    "status", state.ExitCode())
		}
		c <- err
	}()

	go func() {
		var DISCARD string
		var state rune
		for {
			//time.Sleep(15 * time.Millisecond)
			time.Sleep(1 * time.Second)
			stat, _ := ioutil.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
			fmt.Sscanf(string(stat), "%s %s %c", &DISCARD, &DISCARD, &state)

			// REF: https://www.man7.org/linux/man-pages/man1/ps.1.html#PROCESS_STATE_CODES
			if state != 'R' && state != 'S' {
				fmt.Println("STATE", string(state))
			}
		}
	}()

	<-c
	*/

}
