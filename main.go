package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
)

var (
	fileRoutines = 10
)

type MuxMap struct {
	result map[string]int64
	mx     sync.Mutex
}

func NewMuxMap() *MuxMap {
	return &MuxMap{
		result: make(map[string]int64),
	}
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("send absolute path to folder as first argument and count of routines as second")
		return
	}
	fmt.Println("path is " + os.Args[1])
	if len(args) > 2 {
		var err error
		if fileRoutines, err = strconv.Atoi(os.Args[2]); err != nil {
			panic(err)
		}
		if fileRoutines <= 0 {
			fileRoutines = 10
		}
	}
	fmt.Println(fmt.Sprintf("process start at %d routines", fileRoutines))
	filesToAnalise, err := getFiles(os.Args[1])
	if err != nil {
		panic(err)
	}
	muxMap := NewMuxMap()
	chCnt := make(chan int, fileRoutines)
	chErr := make(chan error, len(filesToAnalise))
	for i := range filesToAnalise {
		go muxMap.analyseFile(filesToAnalise[i], chCnt, chErr)
	}
	for range filesToAnalise {
		<-chErr
	}
	fmt.Println(muxMap.result)
}

func (mm *MuxMap) analyseFile(path string, chCnt chan int, chErr chan error) {
	var err error
	chCnt <- 1
	fmt.Println("start to parse " + path)
	file, err := os.Open(path)
	if err != nil {
		_ = <-chCnt
		chErr <- err
		return
	}
	defer file.Close()
	r := bufio.NewReader(file)
	fileMap := make(map[string]int64)
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			break
		}
		for i := range line {
			vl := fileMap[string(line[i])]
			fileMap[string(line[i])] = vl + 1
		}
	}
	mm.mx.Lock()
	for nm := range fileMap {
		z := mm.result[nm] + fileMap[nm]
		mm.result[nm] = z
	}
	mm.mx.Unlock()
	_ = <-chCnt
	chErr <- err
}

func getFiles(path string) ([]string, error) {
	var res []string
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for i := range files {
			if files[i].IsDir() {
				continue
			}
			res = append(res, path+"/"+files[i].Name())
		}
	} else {
		res = append(res, path)
	}
	return res, nil
}
