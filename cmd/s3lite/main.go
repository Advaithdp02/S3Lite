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

	fmt.Println(store)

	
}
