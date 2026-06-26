//Package storage which contains structure of storage and new function to intialize 
package storage

import (
	"os"
	"io"
	"path/filepath"
)

type Storage struct{
	Root string
}
//New is a constructer
func New(root string) *Storage{
	return &Storage{
		Root: root,
	}
}
//Upload is used to upload a single file to destination folder

func (s *Storage) Upload(path string) error{
	if err:=os.MkdirAll(s.Root, os.ModePerm);err!=nil{

		return err
	}
	filename:=filepath.Base(path)

	sourcefile,err:=os.Open(path)

	if(err!=nil){
		return err
	}
	defer sourcefile.Close()
	destinationpath:=filepath.Join(s.Root,filename)
	destinationFile,err:=os.Create(destinationpath)
	if (err!=nil){
		return err
	}
	defer destinationFile.Close()

	_,err=io.Copy(destinationFile, sourcefile)
	if(err!=nil){
		return err

	}
	return nil
} 
