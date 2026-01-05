
package container

import (
    "fmt"
    "os"
    "path/filepath"
)

type CgroupConfig struct {
    Memory string 
    CPU    int    
    Pids   int    
}


func setupCgroup(pid int, cfg CgroupConfig) (string, error) {
    base := "/sys/fs/cgroup/mdocker"

    if err := os.MkdirAll(base, 0755); err != nil {
        return "", err
    }

    // Enable controllers on mdocker parent
    _ = os.WriteFile(
        filepath.Join(base, "cgroup.subtree_control"),
        []byte("+cpu +memory +pids"),
        0644,
    )

    cgPath := filepath.Join(base, fmt.Sprintf("cg-%d", pid))
    if err := os.MkdirAll(cgPath, 0755); err != nil {
        return "", err
    }

    // Memory limit (example: 100MB)
    if cfg.Memory != "" {
    _ = os.WriteFile(
        filepath.Join(cgPath, "memory.max"),
        []byte(cfg.Memory),
        0644,
    )
    }

    // CPU limit (20%)
    if cfg.CPU > 0 {
    // 100000 = 100ms period
    cpuVal := fmt.Sprintf("%d 100000", cfg.CPU*1000)
    _ = os.WriteFile(
        filepath.Join(cgPath, "cpu.max"),
        []byte(cpuVal),
        0644,
    )
    }

    // PID limit
    if cfg.Pids > 0 {
    _ = os.WriteFile(
        filepath.Join(cgPath, "pids.max"),
        []byte(fmt.Sprint(cfg.Pids)),
        0644,
    )
    }

    // Attach process
    if err := os.WriteFile(
        filepath.Join(cgPath, "cgroup.procs"),
        []byte(fmt.Sprint(pid)),
        0644,
    ); err != nil {
        return "", err
    }

    return cgPath, nil
}

func cleanupCgroup(path string) {
    _ = os.Remove(path)
}

func killCgroup(cgPath string) error {
    return os.WriteFile(
        filepath.Join(cgPath, "cgroup.kill"),
        []byte("1"),
        0644,
    )
}

func freezeCgroup(cgPath string) error {
    return os.WriteFile(
        filepath.Join(cgPath, "cgroup.freeze"),
        []byte("1"),
        0644,
    )
}

func thawCgroup(cgPath string) error {
    return os.WriteFile(
        filepath.Join(cgPath, "cgroup.freeze"),
        []byte("0"),
        0644,
    )
}

func cgroupPathFromPID(pid string) string {
    return filepath.Join("/sys/fs/cgroup/mdocker", "cg-"+pid)
}

func FreezeByPID(pid string) error {
    return freezeCgroup(cgroupPathFromPID(pid))
}

func UnfreezeByPID(pid string) error {
    return thawCgroup(cgroupPathFromPID(pid))
}

func KillByPID(pid string) error {
    return killCgroup(cgroupPathFromPID(pid))
}

