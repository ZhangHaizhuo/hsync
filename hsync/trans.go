package hsync

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type User struct {
	Name string
	Psw  string
	Home string
}

type Trans struct {
	Home string
}

func NewTrans(home string) *Trans {
	home, _ = filepath.Abs(home)
	return &Trans{Home: home}
}

type FileStat struct {
	Mtime    time.Time
	Size     int64
	Md5      string
	FileMode os.FileMode
	Exists   bool
}

func (stat *FileStat) IsDir() bool {
	return stat.FileMode.IsDir() //&& stat.FileMode&os.ModeSymlink != 1
}

type MyFile struct {
	Name string
	Data []byte
	Stat *FileStat
	Gzip bool
}

func (f *MyFile) ToString() string {
	return fmt.Sprintf("Name:%s,Mode:%v,Size:%d", f.Name, f.Stat.FileMode, f.Stat.Size)
}

func (trans *Trans) cleanFileName(rel_name string) (fullName string, relName string, err error) {
	fullName, err = filepath.Abs(trans.Home + "/" + rel_name)
	if err != nil {
		return
	}
	relName, err = filepath.Rel(trans.Home, fullName)
	return
}

func (trans *Trans) FileStat(relName string, result *FileStat) (err error) {
	glog.Infoln("Call FileStat", relName)
	fullName, _, err := trans.cleanFileName(relName)
	if err != nil {
		return err
	}
	err = fileGetStat(fullName, result)
	return err
}

func (trans *Trans) CopyFile(myFile *MyFile, result *int) error {
	glog.Infoln("Call CopyFile ", myFile.ToString())
	fullName, _, err := trans.cleanFileName(myFile.Name)
	if err != nil {
		glog.Warningln("CopyFile err:", err)
		return fmt.Errorf("wrong file name")
	}
	dir := fullName
	if !myFile.Stat.IsDir() {
		dir = filepath.Dir(fullName)
	}
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		os.MkdirAll(dir, myFile.Stat.FileMode)
	}
	if err != nil {
		return err
	}
	if !myFile.Stat.IsDir() {
		err = ioutil.WriteFile(fullName, myFile.Data, myFile.Stat.FileMode)
	}
	*result = 1
	return err
}

func (trans *Trans) DeleteFile(relName string, result *int) (err error) {
	glog.Infoln("Call DeleteFile", relName)
	fullName, _, err := trans.cleanFileName(relName)
	if err != nil {
		return err
	}
	err = os.Remove(fullName)
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}

func fileGetStat(name string, stat *FileStat) error {
	info, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	stat.Exists = true
	stat.Mtime = info.ModTime()
	stat.Size = info.Size()
	stat.FileMode = info.Mode()
	if !stat.IsDir() {
		stat.Md5 = FileMd5(name)
	}
	return nil
}

func fileGetMyFile(absPath string) (*MyFile, error) {
	stat := new(FileStat)
	err := fileGetStat(absPath, stat)
	if err != nil {
		return nil, err
	}
	f := &MyFile{
		Name: absPath,
		Stat: stat,
		Gzip: false,
	}
	if !stat.IsDir() {
		f.Data, err = ioutil.ReadFile(absPath)
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}
