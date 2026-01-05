package container

import (
	"os"
	"os/exec"
	"syscall"
	"path/filepath"
	"fmt"
	"strconv"
)


func getProjectRoot() (string, error) {
	exePath, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}


func Run(args []string, cfg CgroupConfig) error {
    if os.Getenv("MDOCKER_CHILD") == "1" {
        return child(args)
    }
    return parent(args, cfg)
}

func closeExtraFiles() error {
	var rlim syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim); err != nil {
		return err
	}

	maxFD := uintptr(rlim.Cur)

	for fd := uintptr(3); fd < maxFD; fd++ {
		_ = syscall.Close(int(fd))
	}
	return nil
}

func mustRun(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}


func parent(args []string, cfg CgroupConfig) error {

	// CRITICAL: close inherited FDs
	if err := closeExtraFiles(); err != nil {
		return err
	}

	// FIX: Prepend "run" so the re-executed command is parsed correctly
	// as "./mdocker run <args>" instead of "./mdocker <args>"
    cmd := exec.Command("/proc/self/exe", append([]string{"run"}, args...)...)
	
    cmd.Env = append(os.Environ(), "MDOCKER_CHILD=1")
	cmd.ExtraFiles = nil
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
        Cloneflags: syscall.CLONE_NEWUTS |
            syscall.CLONE_NEWPID |
            syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET,
    }

    if err := cmd.Start(); err != nil {
    return err
	}

	pid := cmd.Process.Pid
	fmt.Println("container pid:", pid)

	// cleanup old veth if exists
	exec.Command("ip", "link", "del", "veth-host").Run()

	mustRun(exec.Command("ip", "link", "add",
		"veth-host", "type", "veth",
		"peer", "name", "veth-cont"))

	mustRun(exec.Command("ip", "link", "set",
		"veth-cont", "netns", strconv.Itoa(pid)))

	mustRun(exec.Command("ip", "addr", "add",
		"10.0.0.1/24", "dev", "veth-host"))

	mustRun(exec.Command("ip", "link", "set",
		"veth-host", "up"))

	mustRun(exec.Command("nsenter", "-t", strconv.Itoa(pid), "-n",
		"ip", "link", "set", "lo", "up"))

	mustRun(exec.Command("nsenter", "-t", strconv.Itoa(pid), "-n",
		"ip", "addr", "add", "10.0.0.2/24", "dev", "veth-cont"))

	mustRun(exec.Command("nsenter", "-t", strconv.Itoa(pid), "-n",
		"ip", "link", "set", "veth-cont", "up"))

	mustRun(exec.Command("nsenter", "-t", strconv.Itoa(pid), "-n",
	"ip", "route", "add", "default", "via", "10.0.0.1"))


    cgPath, err := setupCgroup(cmd.Process.Pid, cfg)
    if err != nil {
        return err
    }
    defer func() {
		killCgroup(cgPath)
		cleanupCgroup(cgPath)
	}()

	return cmd.Wait()

}

func child(args []string) error {
	// Safety: ensure no host FDs survive
	_ = closeExtraFiles()

	// Set container hostname (visual proof of isolation)
	if err := syscall.Sethostname([]byte("mdocker")); err != nil {
		return err
	}

	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")


	// Get project root directory
	projectRoot, err := getProjectRoot()
	if err != nil {
		return err
	}

	rootfs := filepath.Join(projectRoot, "rootfs")

	/// ---- Pivot Root ----

	// Make rootfs a mount point
	if err := syscall.Mount(rootfs, rootfs, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return err
	}

	putOld := rootfs + "/.old_root"
	if err := os.MkdirAll(putOld, 0700); err != nil {
		return err
	}

	// Pivot root
	if err := syscall.PivotRoot(rootfs, putOld); err != nil {
	return err
	}

	// Change working directory to new root
	if err := syscall.Chdir("/"); err != nil {
	return err
	}

	// Unmount old root
	if err := syscall.Unmount("/.old_root", syscall.MNT_DETACH); err != nil {
	return err
	}
	_ = os.RemoveAll("/.old_root")



	// Ensure /proc exists
	_ = os.MkdirAll("/proc", 0555)

	// Mount proc for this PID namespace
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return err
	}

	// Set PATH environment variable
	os.Setenv("PATH", "/bin:/usr/bin:/sbin:/usr/sbin")

	// Mount devtmpfs for this PID namespace
	syscall.Mount("devtmpfs", "/dev", "devtmpfs", 0, "")
	os.MkdirAll("/dev", 0755)

	/// ---- PID 1 INIT LOGIC STARTS HERE ----

	// Starting the actual container command
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	// Removed reaper loop
	return cmd.Wait()
}
