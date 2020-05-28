package main

import (
	"fmt"
	"os"
	"time"
	"syscall"
	"io/ioutil"
)


// REF: https://golang.org/pkg/syscall/#SysProcAttr
// REF: http://man7.org/linux/man-pages/man2/unshare.2.html
func main() {

	fmt.Println(os.Args)

	if len(os.Args) > 1 {

		if os.Getenv("_") != fmt.Sprintf("%d", os.Getppid()) {
			os.Exit(1)
		}

		// Create session
		// Evitar que muera el hijo si matan el padre
		syscall.Setsid()

		fmt.Println("parent pid:", os.Getppid())
		fmt.Println(os.Environ())

		fmt.Println("[local namespace] test 1 -> mkdir /tmp/vinicio/test1")
		fmt.Println(os.Mkdir("/tmp/vinicio/test1", 0755))

		// Crear un nuevo mount namespace
		if err := syscall.Unshare(syscall.CLONE_NEWNS); err != nil {
			fmt.Println("unshare", err)
		}

		fmt.Println("wait mount")

		time.Sleep(1 * time.Minute)

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

		files, err := ioutil.ReadDir("/tmp/vinicio")

		fmt.Println("[new namespace] test 4 -> ls /tmp/vinicio/")
		for _, file := range files {
			fmt.Println(" - ", file.Name())
		}

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

}
