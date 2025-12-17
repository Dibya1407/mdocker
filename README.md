# MiniDocker
## Overview

MiniDocker (mdocker) is a minimal container runtime implemented in Go to demonstrate how containers work internally using Linux kernel primitives, without relying on Docker or containerd.

## Work Completed So Far

1. Basic CLI Structure
Implemented CLI using cobra in Go with a run command  
it re executes itself to separate parent and child logic  
Run as-

    sudo ./mdocker run /bin/sh

2. Linux Namespace Isolation  
    Following have been implemented so far-
      
    PID Namespace
        Processes inside container cannot see host processes  
        The container has its own PID 1  
    
    Mount Namespace  
        Isolates mount operations, doesnt affect host  
    
    UTS Namespace  
        Separate hostname inside container, used as a visible proof of isolation  

3. Safe Re-exec and File Descriptor Handling  
initially whenever I ran the container, my Arch linux setup would basically stop working, i lose accesss to all my files for some reason, apparently that was because the container inherited file descriptors from the user session, which caused problems in host. Luckily rebooting did fix my setup.

    
    To fix this:  
        All file descriptors except stdin, stdout, and stderr are explicitly closed before re-executing.
        This prevents leaking DBus, Wayland, or other host session sockets into the container.
        This step was critical to make the runtime safe to test.  

5. Filesystem Isolation using pivot_root  
   An Alpine Linux minirootfs is used as the container root filesystem.(I have included the rootfs in repo for testing, Ofc I didnt make my own rootfs)  
    pivot_root is implemented instead of chroot for stronger isolation.  
    The old root filesystem is unmounted and removed after pivoting.


    After this step:  
    The container cannot access the host filesystem.  
    The container sees only its own root filesystem.  

7. Other Minor Stuff  
   A minimal reaper to kill zombie processes to prevent accumulation
   The container doesnt inherit environment variables automatically, so PATH is explicitly set inside container
   The path to rootfs is resolved dynamically relative to the executable instead of being hardcoded.  


## Challenges Faced
1. File Descriptor Leaks  
   Initially caused host desktop issues  
    Learned that re-exec without closing FDs is dangerous.  
    Fixed by explicitly closing all unnecessary descriptors.  

3. Environment Loss after pivot_root  
   Commands failed due to missing PATH.  
    Realized container runtimes must explicitly set environment variables.

4. Terminal / TTY Complexity(Main issue till now)  
   As of now, On starting the container i get the warning "/bin/sh: can't access tty; job control turned off", the rest of the container does work though.  
    Attempted to implement PTY and job control support.  
    Caused the same issue, I lose access to files, i have to reboot to fix  
    Decision made to defer PTY support and focus on core correctness as of now.  


## Research/Concepts Learned  
Golang basics,Cobra fundamentals  
How Containers work,Their purpose  
Linux Namespaces,cgroups  
Difference between chroot and pivot_root  
How container root filesystems work  

## Planned future work  
Adding cgroups for CPU and memory limits  
adding networking(maybe)  
improving usability, further documentation and cleanup  
