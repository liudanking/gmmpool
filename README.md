# gmmpool

A multi level memory pool for Golang:


![](https://ws1.sinaimg.cn/large/44cd29dagy1fociejthjoj20n40ckgm0.jpg)


## Usage


```
package main

import (
	"bytes"
	"log"

	"github.com/liudanking/gmmpool"
)

func main() {
	pool := gmmpool.NewMultiLevelPool([]gmmpool.PoolOpt{
		gmmpool.PoolOpt{Num: 10, Size: 1024},     // level 0
		gmmpool.PoolOpt{Num: 10, Size: 1024 * 2}, // level 1
	})

	buf := pool.Get(1025)
	data, err := buf.ReadAll(bytes.NewReader(make([]byte, 8)))
	if err != nil {
		log.Fatal(err)
	}
	log.Print(data)
}


```



## Credit

[goim](https://github.com/Terry-Mao/goim/)


