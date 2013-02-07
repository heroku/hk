package main

// #include <mach-o/dyld.h>
import "C"
import "unsafe"

func binPath() string {
	var n C.uint32_t
	C._NSGetExecutablePath(nil, &n)
	b := make([]byte, int(n))
	C._NSGetExecutablePath((*C.char)(unsafe.Pointer(&b[0])), &n)
	return string(b[:n-1])
}
