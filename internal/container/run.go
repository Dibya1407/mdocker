package container

import (
	"os"
	"os/exec"
	"syscall"
)

const rootfs = "/home/ctrtest/rootfs"


func Run(args []string) error {
    if os.Getenv("MDOCKER_CHILD") == "1" {
        return child(args)
    }
    return parent(args)
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


func parent(args []string) error {

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
            syscall.CLONE_NEWNS,
    }

    return cmd.Run()
}

func child(args []string) error {
	// Safety: ensure no host FDs survive
	_ = closeExtraFiles()

	// Set container hostname (visual proof of isolation)
	if err := syscall.Sethostname([]byte("mdocker")); err != nil {
		return err
	}

	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")


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


	/// ---- PID 1 INIT LOGIC STARTS HERE ----

	// Starting the actual container command
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	// Reaper loop
	for {
		var status syscall.WaitStatus
		pid, err := syscall.Wait4(-1, &status, 0, nil)

		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return err
		}

		// Exiting if the main child exited
		if pid == cmd.Process.Pid {
			if status.Exited() {
				os.Exit(status.ExitStatus())
			}
			if status.Signaled() {
				os.Exit(128 + int(status.Signal()))
			}
		}
	}
}


