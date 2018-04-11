package filemon

import (
	"sync"
	"fmt"
	"time"
	"os"
	"os/exec"
	"strings"
	"io/ioutil"
	"github.com/howeyc/fsnotify"
)

const (
	filePath   = "/home/ubuntu/GzhuOJ/public"
	hostname   = "root@gzhuacm.cn"
	remotePath = "/home/public"
)

var watcher *fsnotify.Watcher
var mutex sync.Mutex

func Print(args ...interface{}) {
	fmt.Println(time.Now(), args)
}
func isDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		Print("error:", err.Error())
		return false
	}
	return fileInfo.IsDir()
}
func watchPath(filePath string) {
	Print("watchPath:", filePath)
	err := watcher.Watch(filePath)
	if err != nil {
		Print(err.Error())
		return
	}
}
func broweDir(path string) {
	Print("broweDir:", path)
	dir, err := os.Open(path)
	if err != nil {
		Print("error:", err.Error())
		return
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		Print("error:", err.Error())
		return
	}
	for _, name := range names {
		dirPath := path + "/" + name
		if !isDir(dirPath) {
			continue
		}
		watchPath(dirPath)
		broweDir(dirPath)
	}
}

func Run_file_sync() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	broweDir(filePath)
	watchPath(filePath)
	dealWatch()
}
func copy(event *fsnotify.FileEvent) *exec.Cmd {
	return exec.Command(
		"scp",
		"-r",
		"-P 23456",
		event.Name,
		hostname+":"+remotePath+strings.TrimPrefix(event.Name, filePath))
}
func remove(event *fsnotify.FileEvent) *exec.Cmd {
	return exec.Command(
		"ssh",
		"-p 23456",
		hostname,
		`rm -r `+remotePath+strings.TrimPrefix(event.Name, filePath)+``)
}
func dealWatch() {
	for {
		func() {
			//mutex.Lock()
			//defer mutex.Unlock()
			select {
			case event := <-watcher.Event:
				Print("event: ", event)
				var cmd *exec.Cmd
				if event.IsCreate() || event.IsModify() {
					cmd = copy(event)
				}
				if event.IsDelete() || event.IsRename() {
					cmd = remove(event)
				}
				Print("cmd:", cmd.Args)
				stderr, err := cmd.StderrPipe()
				if err != nil {
					Print(err.Error())
					return
				}
				defer stderr.Close()
				stdout, err := cmd.StdoutPipe()
				if err != nil {
					Print(err.Error())
					return
				}
				defer stdout.Close()
				if err = cmd.Start(); err != nil {
					Print(err.Error())
					return
				}
				errBytes, err := ioutil.ReadAll(stderr)
				if err != nil {
					Print(err.Error())
					return
				}
				outBytes, err := ioutil.ReadAll(stdout)
				if err != nil {
					Print(err.Error())
					return
				}
				if len(errBytes) != 0 {
					Print("errors:", string(errBytes))
				}
				if len(outBytes) != 0 {
					Print("output:", string(outBytes))
				}
				if err = cmd.Wait(); err != nil {
					Print(err.Error())
				}
			case err := <-watcher.Error:
				Print("error: ", err.Error())
			}
		}()
	}
}
