package main

import (
	"fmt"

	"../reddit_refresh_online"
)

func main() {
	fmt.Println(reddit_refresh_online.ValidateSub("thisisamemeandfuckme"))
	fmt.Println(reddit_refresh_online.ValidateSub("gamedeals"))
}
