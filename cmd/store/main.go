package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	file1, err := os.Open("C:\\Users\\李豪\\Desktop\\project\\kitten\\cmd\\store\\2022-04-06.txt")
	if err != nil {
		fmt.Println(err)
	}
	result, err := os.OpenFile("C:\\Users\\李豪\\Desktop\\project\\kitten\\cmd\\store\\result.txt", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer result.Close()
	defer file1.Close()

	stat, err := file1.Stat()
	if err != nil {
		panic(err)
	}
	fmt.Println(stat.Size())
	buf := bufio.NewReader(file1)
	writer := bufio.NewWriter(result)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		if strings.Contains(line, "\"status_code\":502") {
			writer.WriteString(line + "\n")
		}
	}
	writer.Flush()
}
