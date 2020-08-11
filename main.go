package main

import "fmt"

type test struct {
	str string
}

func main() {

	var t *test

	fmt.Println(help(t))
}

func help(t *test) *test {

	t.str = "1"

	return t
}
