/*
 * passthrough.go
 *
 * Copyright 2017 Bill Zissimopoulos
 */
/*
 * This file is part of Cgofuse.
 *
 * You can redistribute it and/or modify it under the terms of the GNU
 * General Public License version 3 as published by the Free Software
 * Foundation.
 */

package main

import (
	"cgofuse/fuse"
	"os"
	"path/filepath"
	"syscall"
)

type Ptfs struct {
	fuse.FileSystemBase
	root string
}

func errno(err error) int {
	if nil != err {
		return -int(err.(syscall.Errno))
	} else {
		return 0
	}
}

func (self *Ptfs) Statfs(path string, stat *fuse.Statfs_t) (errc int) {
	defer Trace(path)(&errc, stat)
	path = filepath.Join(self.root, path)
	stgo := syscall.Statfs_t{}
	errc = errno(syscall.Statfs(path, &stgo))
	copyFusestatfsFromGostatfs(stat, &stgo)
	return
}

func (self *Ptfs) Mknod(path string, mode uint32, dev uint64) (errc int) {
	defer Trace(path, mode, dev)(&errc)
	path = filepath.Join(self.root, path)
	return errno(syscall.Mknod(path, mode, int(dev)))
}

func (self *Ptfs) Mkdir(path string, mode uint32) (errc int) {
	defer Trace(path, mode)(&errc)
	path = filepath.Join(self.root, path)
	return errno(syscall.Mkdir(path, mode))
}

func (self *Ptfs) Unlink(path string) (errc int) {
	defer Trace(path)(&errc)
	path = filepath.Join(self.root, path)
	return errno(syscall.Unlink(path))
}

func (self *Ptfs) Rmdir(path string) (errc int) {
	defer Trace(path)(&errc)
	path = filepath.Join(self.root, path)
	return errno(syscall.Rmdir(path))
}

func (self *Ptfs) Link(oldpath string, newpath string) (errc int) {
	defer Trace(oldpath, newpath)(&errc)
	oldpath = filepath.Join(self.root, oldpath)
	newpath = filepath.Join(self.root, newpath)
	return errno(syscall.Link(oldpath, newpath))
}

func (self *Ptfs) Symlink(target string, newpath string) (errc int) {
	defer Trace(target, newpath)(&errc)
	target = filepath.Join(self.root, target)
	newpath = filepath.Join(self.root, newpath)
	return errno(syscall.Symlink(target, newpath))
}

func (self *Ptfs) Readlink(path string) (errc int, target string) {
	defer Trace(path)(&errc, &target)
	path = filepath.Join(self.root, path)
	buff := [1024]byte{}
	n, e := syscall.Readlink(path, buff[:])
	if e != nil {
		return errno(e), ""
	}
	return 0, string(buff[:n])
}

func (self *Ptfs) Rename(oldpath string, newpath string) (errc int) {
	defer Trace(oldpath, newpath)(&errc)
	oldpath = filepath.Join(self.root, oldpath)
	newpath = filepath.Join(self.root, newpath)
	return errno(syscall.Rename(oldpath, newpath))
}

func (self *Ptfs) Chmod(path string, mode uint32) (errc int) {
	defer Trace(path, mode)(&errc)
	path = filepath.Join(self.root, path)
	return errno(syscall.Chmod(path, mode))
}

func (self *Ptfs) Chown(path string, uid uint32, gid uint32) (errc int) {
	defer Trace(path, uid, gid)(&errc)
	path = filepath.Join(self.root, path)
	return errno(syscall.Chown(path, int(uid), int(gid)))
}

func (self *Ptfs) Utimens(path string, tmsp1 []fuse.Timespec) (errc int) {
	defer Trace(path, tmsp1)(&errc)
	path = filepath.Join(self.root, path)
	tmsp := [2]syscall.Timespec{}
	tmsp[0].Sec, tmsp[0].Nsec = tmsp1[0].Sec, tmsp1[0].Nsec
	tmsp[1].Sec, tmsp[1].Nsec = tmsp1[1].Sec, tmsp1[1].Nsec
	return errno(syscall.UtimesNano(path, tmsp[:]))
}

func (self *Ptfs) Access(path string, mask uint32) (errc int) {
	defer Trace(path, mask)(&errc)
	path = filepath.Join(self.root, path)
	return errno(syscall.Access(path, mask))
}

func (self *Ptfs) Create(path string, mode uint32) (errc int, fh uint64) {
	defer Trace(path, mode)(&errc, &fh)
	path = filepath.Join(self.root, path)
	f, e := syscall.Open(path, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_TRUNC, mode)
	if e != nil {
		return errno(e), ^uint64(0)
	}
	return 0, uint64(f)
}

func (self *Ptfs) Open(path string, flags int) (errc int, fh uint64) {
	defer Trace(path, flags)(&errc, &fh)
	path = filepath.Join(self.root, path)
	f, e := syscall.Open(path, flags, 0)
	if e != nil {
		return errno(e), ^uint64(0)
	}
	return 0, uint64(f)
}

func (self *Ptfs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	defer Trace(path, fh)(&errc, stat)
	stgo := syscall.Stat_t{}
	if ^uint64(0) == fh {
		path = filepath.Join(self.root, path)
		errc = errno(syscall.Stat(path, &stgo))
	} else {
		errc = errno(syscall.Fstat(int(fh), &stgo))
	}
	copyFusestatFromGostat(stat, &stgo)
	return
}

func (self *Ptfs) Truncate(path string, size int64, fh uint64) (errc int) {
	defer Trace(path, size, fh)(&errc)
	if ^uint64(0) == fh {
		path = filepath.Join(self.root, path)
		errc = errno(syscall.Truncate(path, size))
	} else {
		errc = errno(syscall.Ftruncate(int(fh), size))
	}
	return
}

func (self *Ptfs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer Trace(path, buff, ofst, fh)(&n)
	n, e := syscall.Pread(int(fh), buff, ofst)
	if e != nil {
		return errno(e)
	}
	return n
}

func (self *Ptfs) Write(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer Trace(path, buff, ofst, fh)(&n)
	n, e := syscall.Pwrite(int(fh), buff, ofst)
	if e != nil {
		return errno(e)
	}
	return n
}

func (self *Ptfs) Release(path string, fh uint64) (errc int) {
	defer Trace(path, fh)(&errc)
	return errno(syscall.Close(int(fh)))
}

func (self *Ptfs) Fsync(path string, datasync bool, fh uint64) (errc int) {
	defer Trace(path, datasync, fh)(&errc)
	return errno(syscall.Fsync(int(fh)))
}

func (self *Ptfs) Opendir(path string) (errc int, fh uint64) {
	defer Trace(path)(&errc, &fh)
	path = filepath.Join(self.root, path)
	f, e := syscall.Open(path, syscall.O_RDONLY|syscall.O_DIRECTORY, 0)
	if e != nil {
		return errno(e), ^uint64(0)
	}
	return 0, uint64(f)
}

func (self *Ptfs) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	defer Trace(path, fill, ofst, fh)(&errc)
	file, e := os.Open(path)
	if e != nil {
		return errno(e)
	}
	defer file.Close()
	nams, e := file.Readdirnames(0)
	if e != nil {
		return errno(e)
	}
	for _, name := range nams {
		if !fill(name, nil, 0) {
			break
		}
	}
	return 0
}

func (self *Ptfs) Releasedir(path string, fh uint64) (errc int) {
	defer Trace(path, fh)(&errc)
	return errno(syscall.Close(int(fh)))
}

func main() {
	ptfs := Ptfs{}
	args := os.Args
	if 3 <= len(args) && '-' != args[len(args)-2][0] && '-' != args[len(args)-1][0] {
		ptfs.root, _ = filepath.Abs(args[len(args)-2])
		args = append(args[:len(args)-2], args[len(args)-1])
	}
	host := fuse.NewFileSystemHost(&ptfs)
	host.Mount(args)
}
