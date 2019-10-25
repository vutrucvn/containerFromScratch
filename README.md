# container_from_scratch

# install a file system (to be used for the container)
 https://help.ubuntu.com/community/BasicChroot

The code hardcoded the file system under "/var/chroot"

# run the code like this
go run main.go run /bin/bash

# test 
hostname    # UTS
ip addr     # netns
ipcs        # IPC namespace
id          # USER, we use root for both
ps, top     # PID
ls /proc    # NS (mount)

cat /sys/fs/cgroup/pids/medium/cgroup.procs # see current pids running on this cgroup pids, named "medium"

sudo readlink /proc/<pid>/ns/<kind>     # see the inode (real namespace id) that the pid link to