package io

import (
	"fmt"
	"io/ioutil"
)

func ReadModelFile() (string, error) {
	f, err := ioutil.ReadFile("./template/struct.model")
	if err != nil {
		fmt.Println("read fail", err)
	}
	return string(f), err
}

func GenerateModel(filename string, modelByte []byte) error {
	fileName := filename

	err := ioutil.WriteFile(fileName, modelByte, 0666)
	if err != nil {
		fmt.Println("write fail")
	}
	fmt.Println("write success")
	return err
}
