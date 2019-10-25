package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// go run main.go run <cmd> <args>
func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("help")
	}
}

func run() {
	fmt.Printf("Running %v \n", os.Args[2:])

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWUSER | syscall.CLONE_NEWIPC,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	// // grand this new user root access in the new user namespace
	cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: 0,
		Gid: 0,
	}

	// map the new user (in host: root) to root user in the container
	cmd.SysProcAttr.UidMappings = []syscall.SysProcIDMap{{ContainerID: 0, HostID: 0, Size: 1}}
	cmd.SysProcAttr.GidMappings = []syscall.SysProcIDMap{{ContainerID: 0, HostID: 0, Size: 1}}

	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v \n", os.Args[2:])

	cg()

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot("/var/chroot"))
	must(os.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))
	// must(syscall.Mount("/home/etrucvu/hack19", "/tmp", "tmpfs", 0, ""))

	must(cmd.Run())

	must(syscall.Unmount("proc", 0))
	// must(syscall.Unmount("/home/etrucvu/hack19", 0))
}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	cpu := filepath.Join(cgroups, "cpu")
	memory := filepath.Join(cgroups, "memory")

	// cgroup name is "medium"
	os.Mkdir(filepath.Join(pids, "medium"), 0755)
	os.Mkdir(filepath.Join(cpu, "medium"), 0755)
	os.Mkdir(filepath.Join(memory, "medium"), 0755)

	// configure the cgroup "medium", for PIDs
	must(ioutil.WriteFile(filepath.Join(pids, "medium/pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place after the container exits
	must(ioutil.WriteFile(filepath.Join(pids, "medium/notify_on_release"), []byte("1"), 0700))
	// add the pid of the container's process to cgroup 'medium'
	must(ioutil.WriteFile(filepath.Join(pids, "medium/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))

	// configure CPU
	must(ioutil.WriteFile(filepath.Join(cpu, "medium/cpu.cfs_quota_us"), []byte("200000"), 0700))
	must(ioutil.WriteFile(filepath.Join(cpu, "medium/cpu.cfs_period_us"), []byte("1000000"), 0700))
	must(ioutil.WriteFile(filepath.Join(cpu, "medium/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
	must(ioutil.WriteFile(filepath.Join(cpu, "medium/notify_on_release"), []byte("1"), 0700))
	// configure memory
	must(ioutil.WriteFile(filepath.Join(memory, "medium/memory.limit_in_bytes"), []byte("1024000000"), 0700))
	must(ioutil.WriteFile(filepath.Join(memory, "medium/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
	must(ioutil.WriteFile(filepath.Join(memory, "medium/notify_on_release"), []byte("1"), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
