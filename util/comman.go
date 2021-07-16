// Copyright (C) 2021 Gridworkz Co., Ltd.
// KATO, Application Management Platform

// Permission is hereby granted, free of charge, to any person obtaining a copy of this 
// software and associated documentation files (the "Software"), to deal in the Software
// without restriction, including without limitation the rights to use, copy, modify, merge,
// publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons 
// to whom the Software is furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all copies or 
// substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, 
// INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR
// PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE
// FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package util

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gridworkz/kato/util/zip"
	"github.com/sirupsen/logrus"
)

//CheckAndCreateDir
func CheckAndCreateDir(path string) error {
	if subPathExists, err := FileExists(path); err != nil {
		return fmt.Errorf("Could not determine if subPath %s exists; will not attempt to change its permissions", path)
	} else if !subPathExists {
		// Create the sub path now because if it's auto-created later when referenced, it may have an
		// incorrect ownership and mode. For example, the sub path directory must have at least g+rwx
		// when the pod specifies an fsGroup, and if the directory is not created here, Docker will
		// later auto-create it with the incorrect mode 0750
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to mkdir:%s", path)
		}

		if err := os.Chmod(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

//CheckAndCreateDirByMode
func CheckAndCreateDirByMode(path string, mode os.FileMode) error {
	if subPathExists, err := FileExists(path); err != nil {
		return fmt.Errorf("Could not determine if subPath %s exists; will not attempt to change its permissions", path)
	} else if !subPathExists {
		// Create the sub path now because if it's auto-created later when referenced, it may have an
		// incorrect ownership and mode. For example, the sub path directory must have at least g+rwx
		// when the pod specifies an fsGroup, and if the directory is not created here, Docker will
		// later auto-create it with the incorrect mode 0750
		if err := os.MkdirAll(path, mode); err != nil {
			return fmt.Errorf("failed to mkdir:%s", path)
		}

		if err := os.Chmod(path, mode); err != nil {
			return err
		}
	}
	return nil
}

//DirIsEmpty
func DirIsEmpty(dir string) bool {
	infos, err := ioutil.ReadDir(dir)
	if len(infos) == 0 || err != nil {
		return true
	}
	return false
}

//OpenOrCreateFile
func OpenOrCreateFile(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777)
}

//FileExists
func FileExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

//SearchFileBody
func SearchFileBody(filename, searchStr string) bool {
	body, _ := ioutil.ReadFile(filename)
	return strings.Contains(string(body), searchStr)
}

//IsHaveFile
//Except for opening files
func IsHaveFile(path string) bool {
	files, _ := ioutil.ReadDir(path)
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") {
			return true
		}
	}
	return false
}

//SearchFile - search whether there is a specified file in the specified directory, specify the number
//of levels of the search directory, -1 is the full directory search
func SearchFile(pathDir, name string, level int) bool {
	if level == 0 {
		return false
	}
	files, _ := ioutil.ReadDir(pathDir)
	var dirs []os.FileInfo
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file)
			continue
		}
		if file.Name() == name {
			return true
		}
	}
	if level == 1 {
		return false
	}
	for _, dir := range dirs {
		ok := SearchFile(path.Join(pathDir, dir.Name()), name, level-1)
		if ok {
			return ok
		}
	}
	return false
}

//FileExistsWithSuffix - whether the specified directory contains files with the specified suffix
func FileExistsWithSuffix(pathDir, suffix string) bool {
	files, _ := ioutil.ReadDir(pathDir)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), suffix) {
			return true
		}
	}
	return false
}

//CmdRunWithTimeout exec cmd with timeout
func CmdRunWithTimeout(cmd *exec.Cmd, timeout time.Duration) (bool, error) {
	done := make(chan error)
	if cmd.Process != nil { //Restore execution state
		cmd.Process = nil
		cmd.ProcessState = nil
	}
	if err := cmd.Start(); err != nil {
		return false, err
	}
	go func() {
		done <- cmd.Wait()
	}()
	var err error
	select {
	case <-time.After(timeout):
		// timeout
		if err = cmd.Process.Kill(); err != nil {
			logrus.Errorf("failed to kill: %s, error: %s", cmd.Path, err.Error())
		}
		go func() {
			<-done // allow goroutine to exit
		}()
		logrus.Infof("process:%s killed", cmd.Path)
		return true, err
	case err = <-done:
		return false, err
	}
}

//ReadHostID - read the current machine ID
//ID is the unique identifier of the node, acp_node will maintain the binding relationship 
//between ID and machine information in etcd
func ReadHostID(filePath string) (string, error) {
	if filePath == "" {
		if runtime.GOOS == "windows" {
			filePath = "c:\\kato\\node_host_uuid.conf"
		} else {
			filePath = "/opt/kato/etc/node/node_host_uuid.conf"
		}
	}
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			uid, err := CreateHostID()
			if err != nil {
				return "", err
			}
			err = ioutil.WriteFile(filePath, []byte("host_uuid="+uid), 0777)
			if err != nil {
				logrus.Error("Write host_uuid file error.", err.Error())
			}
			return uid, nil
		}
		return "", err
	}
	body, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	info := strings.Split(strings.TrimSpace(string(body)), "=")
	if len(info) == 2 {
		return info[1], nil
	}
	return "", fmt.Errorf("Invalid host uuid from file")
}

//CreateHostID create host id by mac addr
func CreateHostID() (string, error) {
	macAddrs := getMacAddrs()
	if macAddrs == nil || len(macAddrs) == 0 {
		return "", fmt.Errorf("read macaddr error when create node id")
	}
	ip, _ := LocalIP()
	hash := md5.New()
	hash.Write([]byte(macAddrs[0] + ip.String()))
	uid := fmt.Sprintf("%x", hash.Sum(nil))
	if len(uid) >= 32 {
		return uid[:32], nil
	}
	for i := len(uid); i < 32; i++ {
		uid = uid + "0"
	}
	return uid, nil
}

func getMacAddrs() (macAddrs []string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v", err)
		return macAddrs
	}

	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}
		macAddrs = append(macAddrs, macAddr)
	}
	return macAddrs
}

//LocalIP Get this machine ip
// Get the first non loopback ip
func LocalIP() (net.IP, error) {
	tables, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, t := range tables {
		addrs, err := t.Addrs()
		if err != nil {
			return nil, err
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() {
				continue
			}
			if v4 := ipnet.IP.To4(); v4 != nil {
				return v4, nil
			}
		}
	}
	return nil, fmt.Errorf("cannot find local IP address")
}

//GetIDFromKey - get id from etcd key
func GetIDFromKey(key string) string {
	index := strings.LastIndex(key, "/")
	if index < 0 {
		return ""
	}
	if strings.Contains(key, "-") { //build in task, in order to distinguish between different nodes
		return strings.Split(key[index+1:], "-")[0]
	}

	return key[index+1:]
}

//Deweight - remove array duplication
func Deweight(data *[]string) {
	var result []string
	if len(*data) < 1024 {
		// When the slice length is less than 1024, loop to filter
		for i := range *data {
			flag := true
			for j := range result {
				if result[j] == (*data)[i] {
					flag = false // There are duplicate elements, the flag is false
					break
				}
			}
			if flag && (*data)[i] != "" { // Identified as false, do not add to the result
				result = append(result, (*data)[i])
			}
		}
	} else {
		// When greater than, filter by map
		var tmp = make(map[string]byte)
		for _, d := range *data {
			l := len(tmp)
			tmp[d] = 0
			if len(tmp) != l && d != "" { // After adding the map, the length of the map changes, and the elements are not repeated
				result = append(result, d)
			}
		}
	}
	*data = result
}

//GetDirSizeByCmd get dir sizes by du command
//return kb
func GetDirSizeByCmd(path string) float64 {
	out, err := CmdExec(fmt.Sprintf("du -sk %s", path))
	if err != nil {
		fmt.Println(err)
		return 0
	}
	info := strings.Split(out, "	")
	fmt.Println(info)
	if len(info) < 2 {
		return 0
	}
	i, _ := strconv.Atoi(info[0])
	return float64(i)
}

//GetFileSize
func GetFileSize(path string) int64 {
	if fileInfo, err := os.Stat(path); err == nil {
		return fileInfo.Size()
	}
	return 0
}

//GetDirSize - kb as unit
func GetDirSize(path string) float64 {
	if ok, err := FileExists(path); err != nil || !ok {
		return 0
	}

	fileSizes := make(chan int64)
	concurrent := make(chan int, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go walkDir(path, &wg, fileSizes, concurrent)

	go func() {
		wg.Wait() //Wait for the goroutine to end
		close(fileSizes)
	}()
	var nfiles, nbytes int64
loop:
	for {
		select {
		case size, ok := <-fileSizes:
			if !ok {
				break loop
			}
			nfiles++
			nbytes += size
		}
	}
	return float64(nbytes / 1024)
}

//Get the file size in the directory dir
func walkDir(dir string, wg *sync.WaitGroup, fileSizes chan<- int64, concurrent chan int) {
	defer wg.Done()
	concurrent <- 1
	defer func() {
		<-concurrent
	}()
	for _, entry := range listDirNonSymlink(dir) {
		if entry.IsDir() { //dictionary
			wg.Add(1)
			subDir := filepath.Join(dir, entry.Name())
			go walkDir(subDir, wg, fileSizes, concurrent)
		} else {
			fileSizes <- entry.Size()
		}
	}
}

//sema is a counting semaphore for limiting concurrency in listDir
var sema = make(chan struct{}, 20)

//Read the file information in the directory dir
func listDir(dir string) []os.FileInfo {
	sema <- struct{}{}
	defer func() { <-sema }()
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		logrus.Errorf("get file sizt: %v\n", err)
		return nil
	}
	return entries
}

// List all entries of non-soft chain type in the specified directory
func listDirNonSymlink(dir string) []os.FileInfo {
	sema <- struct{}{}
	defer func() { <-sema }()
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		logrus.Errorf("get file sizt: %v\n", err)
		return nil
	}

	var result []os.FileInfo
	for i := range entries {
		if entries[i].Mode()&os.ModeSymlink == 0 {
			result = append(result, entries[i])
		}
	}
	return result
}

//RemoveSpaces 
func RemoveSpaces(sources []string) (re []string) {
	for _, s := range sources {
		if s != " " && s != "" && s != "\t" && s != "\n" && s != "\r" {
			re = append(re, s)
			fmt.Println(re)
		}
	}
	return
}

//CmdExec
func CmdExec(args string) (string, error) {
	out, err := exec.Command("bash", "-c", args).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

//Zip - zip compressing source dir to target file
func Zip(source, target string) error {
	if err := CheckAndCreateDir(filepath.Dir(target)); err != nil {
		return err
	}
	zipfile, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		//set file uid and
		elem := reflect.ValueOf(info.Sys()).Elem()
		uid := elem.FieldByName("Uid").Uint()
		gid := elem.FieldByName("Gid").Uint()
		header.Comment = fmt.Sprintf("%d/%d", uid, gid)
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		file, err := os.OpenFile(path, os.O_RDONLY, 0)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

//Unzip archive file to target dir
func Unzip(archive, target string) error {
	reader, err := zip.OpenDirectReader(archive)
	if err != nil {
		return fmt.Errorf("error opening archive: %v", err)
	}
	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}
	for _, file := range reader.File {
		run := func() error {
			path := filepath.Join(target, file.Name)
			if file.FileInfo().IsDir() {
				os.MkdirAll(path, file.Mode())
				if file.Comment != "" && strings.Contains(file.Comment, "/") {
					guid := strings.Split(file.Comment, "/")
					if len(guid) == 2 {
						uid, _ := strconv.Atoi(guid[0])
						gid, _ := strconv.Atoi(guid[1])
						if err := os.Chown(path, uid, gid); err != nil {
							return fmt.Errorf("error changing owner: %v", err)
						}
					}
				}
				return nil
			}

			fileReader, err := file.Open()
			if err != nil {
				return fmt.Errorf("fileReader; error opening file: %v", err)
			}
			defer fileReader.Close()
			targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return fmt.Errorf("targetFile; error opening file: %v", err)
			}
			defer targetFile.Close()

			if _, err := io.Copy(targetFile, fileReader); err != nil {
				return fmt.Errorf("error copy file: %v", err)
			}
			if file.Comment != "" && strings.Contains(file.Comment, "/") {
				guid := strings.Split(file.Comment, "/")
				if len(guid) == 2 {
					uid, _ := strconv.Atoi(guid[0])
					gid, _ := strconv.Atoi(guid[1])
					if err := os.Chown(path, uid, gid); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if err := run(); err != nil {
			return err
		}
	}

	return nil
}

// CopyFile copy source file to target
// direct io read and write file
// Keep the permissions user and group
func CopyFile(source, target string) error {
	sfi, err := os.Stat(source)
	if err != nil {
		return err
	}
	elem := reflect.ValueOf(sfi.Sys()).Elem()
	uid := elem.FieldByName("Uid").Uint()
	gid := elem.FieldByName("Gid").Uint()
	sf, err := os.OpenFile(source, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer sf.Close()
	tf, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, sfi.Mode())
	if err != nil {
		return err
	}
	defer tf.Close()
	_, err = io.Copy(tf, sf)
	if err != nil {
		return err
	}

	if err := os.Chown(target, int(uid), int(gid)); err != nil {
		return err
	}

	return nil
}

//GetParentDirectory
func GetParentDirectory(dirctory string) string {
	return substr(dirctory, 0, strings.LastIndex(dirctory, "/"))
}

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

//Rename - move file
func Rename(old, new string) error {
	_, err := os.Stat(GetParentDirectory(new))
	if err != nil {
		if err == os.ErrNotExist || strings.Contains(err.Error(), "no such file or directory") {
			if err := os.MkdirAll(GetParentDirectory(new), 0755); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return os.Rename(old, new)
}

//MergeDir
//if Subdirectories already exist, Don't replace
func MergeDir(fromdir, todir string) error {
	files, err := ioutil.ReadDir(fromdir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Rename(path.Join(fromdir, f.Name()), path.Join(todir, f.Name())); err != nil {
			if !strings.Contains(err.Error(), "file exists") {
				return err
			}
		}
	}
	return nil
}

//CreateVersionByTime create version number
func CreateVersionByTime() string {
	now := time.Now()
	return now.Format("20060102150405")
}

// GetDirList get all lower level dir
func GetDirList(dirpath string, level int) ([]string, error) {
	var dirlist []string
	list, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, err
	}
	for _, f := range list {
		if f.IsDir() {
			if level <= 1 {
				dirlist = append(dirlist, filepath.Join(dirpath, f.Name()))
			} else {
				list, err := GetDirList(filepath.Join(dirpath, f.Name()), level-1)
				if err != nil {
					return nil, err
				}
				dirlist = append(dirlist, list...)
			}
		}
	}
	return dirlist, nil
}

//GetFileList
func GetFileList(dirpath string, level int) ([]string, error) {
	var dirlist []string
	list, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, err
	}
	for _, f := range list {
		if !f.IsDir() && level <= 1 {
			dirlist = append(dirlist, filepath.Join(dirpath, f.Name()))
		} else if level > 1 && f.IsDir() {
			list, err := GetFileList(filepath.Join(dirpath, f.Name()), level-1)
			if err != nil {
				return nil, err
			}
			dirlist = append(dirlist, list...)
		}
	}
	return dirlist, nil
}

// GetDirNameList get all lower level dir
func GetDirNameList(dirpath string, level int) ([]string, error) {
	var dirlist []string
	list, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, err
	}
	for _, f := range list {
		if f.IsDir() {
			if level <= 1 {
				dirlist = append(dirlist, f.Name())
			} else {
				list, err := GetDirList(filepath.Join(dirpath, f.Name()), level-1)
				if err != nil {
					return nil, err
				}
				dirlist = append(dirlist, list...)
			}
		}
	}
	return dirlist, nil
}

//GetCurrentDir
func GetCurrentDir() string {
	dir, err := filepath.Abs("./")
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

//IsDir path is dir
func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

var reg = regexp.MustCompile(`(?U)\$\{.*\}`)

//ParseVariable parse and replace variable in source str
func ParseVariable(source string, configs map[string]string) string {
	resultKey := reg.FindAllString(source, -1)
	for _, sourcekey := range resultKey {
		key, defaultValue := getVariableKey(sourcekey)
		if value, ok := configs[key]; ok {
			source = strings.Replace(source, sourcekey, value, -1)
		} else if defaultValue != "" {
			source = strings.Replace(source, sourcekey, defaultValue, -1)
		}
	}
	return source
}

func getVariableKey(source string) (key, value string) {
	if len(source) < 4 {
		return "", ""
	}
	left := strings.Index(source, "{")
	right := strings.Index(source, "}")
	k := source[left+1 : right]
	if strings.Contains(k, ":") {
		re := strings.Split(k, ":")
		if len(re) > 1 {
			return re[0], re[1]
		}
		return re[0], ""
	}
	return k, ""
}

// Getenv returns env by key or default value.
func Getenv(key, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
}