# strudel - omaha upgrades to system-level rkt pods

Strudel is a service responsible for upgrading rkt pods of "systems-level"
applications independently of the operating system upgrade path. Ideally this
means the base CoreOS system can shrink to just a kernel, systemd, strudel and
rkt. This is still a work in progress.
