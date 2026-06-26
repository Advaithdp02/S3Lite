package main

import (
	"fmt"
	"os"
	"github.com/Advaithdp02/s3lite/internal/storage"
)

func main(){
	if(len(os.Args)<2){
		fmt.Println("Usage: s3lite <commands>")
		return
	}

	store:=storage.New("storage")

	command:=os.Args[1]

	switch command{
	case "Upload":
		err:=store.Upload(os.Args[2])
		if(err!=nil){
			fmt.Printf("Upload failed :%v",err)
			return 
		}
		fmt.Println("Upload done successfully")
	default:
		fmt.Println("Unknown command")
	}
	


	
}
