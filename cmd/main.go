package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func main() {
	args := os.Args
	if args[1] == "run" {
		run()

	} else if args[1] == "child" {
		child()
	}
}

func createCgroup() *cgroup2.Manager {
	memoryMax := int64(1024 * 1024 * 1024)
	cpuQuota := int64(100000)
	cpuPeriod := uint64(100000)
	pidsMax := int64(100)

	resources := cgroup2.Resources{
		Memory: &cgroup2.Memory{
			Max: &memoryMax,
		},
		CPU: &cgroup2.CPU{
			Max: cgroup2.NewCPUMax(&cpuQuota, &cpuPeriod),
		},
		Pids: &cgroup2.Pids{
			Max: pidsMax,
		},
	}

	manager, err := cgroup2.NewManager(
		"/sys/fs/cgroup/",
		"/crtCont",
		&resources,
	)

	if err != nil {
		panic(err)

	}

	return manager

}

func run() {
	args := os.Args[2:]
	fmt.Println(args)

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, args...)...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	}

	pid := os.Getpid()
	ppid := os.Getppid()

	fmt.Println("PId =", pid, "\n PPid = ", ppid)

	cmd.Run()
}

func child() {
	args := os.Args[1:]
	fmt.Println(args)

	wd, _ := os.Getwd()
	fmt.Println("wd  ", wd)

	syscall.Sethostname([]byte("container"))
	if err := syscall.Chroot("./contRoot/contRoot/"); err != nil {
		fmt.Println(err)
		panic("Chroot failed")

	}
	syscall.Chdir("/")

	wd, _ = os.Getwd()
	fmt.Println("current wd ", wd)

	syscall.Mount("proc", "proc", "proc", 0, "")

	cmd := exec.Command(args[1], args[2:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	pid := os.Getpid()
	ppid := os.Getppid()

	fmt.Println("PId =", pid, "\n PPid = ", ppid)

	cmd.Run()

	syscall.Unmount("/proc", 0)
}
