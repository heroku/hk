package s3util_test

import (
	"github.com/kr/s3/s3util"
	"io"
	"os"
)

func ExampleCreate() {
	s3util.DefaultConfig.AccessKey = "...access key..."
	s3util.DefaultConfig.SecretKey = "...secret key..."
	r, _ := os.Open("/dev/stdin")
	w, _ := s3util.Create("https://mybucket.s3.amazonaws.com/log.txt", nil, nil)
	io.Copy(w, r)
	w.Close()
}

func ExampleOpen() {
	s3util.DefaultConfig.AccessKey = "...access key..."
	s3util.DefaultConfig.SecretKey = "...secret key..."
	r, _ := s3util.Open("https://mybucket.s3.amazonaws.com/log.txt", nil)
	w, _ := os.Create("out.txt")
	io.Copy(w, r)
	w.Close()
}
