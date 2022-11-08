package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	http.HandleFunc("/", hello) // 注册自己业务处理的Hander
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	server := http.Server{Addr: ":8080"}
	go func() {
		if err := server.ListenAndServe(); err != nil { // 监听处理
			fmt.Println("server start failed")
		}
	}()

	// 通过信号量的方式停止服务，如果有一部分请求进行到一半，处理完成再关闭服务器
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	fmt.Printf("接收信号：%s\n", s)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("server shutdown failed")
	}
	fmt.Println("server exit")
}

type Message struct {
	SUC      bool
	TIME     string
	DATA     obj
	ALL_DATA []string
}

type obj struct {
	TITLE string
	DATE  string
	NEWS  []string
	WEIYU string
}

func hello(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("index")
	if id == "" {
		id = "0"
	}
	tpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}
	var revMsg Message
	get := httpGet(id)
	err = json.Unmarshal([]byte(get), &revMsg)
	if err != nil {
		tpl.Execute(w, Message{
			SUC:  false,
			TIME: "获取失败",
			DATA: obj{
				DATE:  "请观看其他日期的早报",
				NEWS:  []string{"当前接口获取失败，无内容，请切换尝试"},
				WEIYU: "暂时没有微语，请切换",
			},
		})
	} else {
		tpl.Execute(w, revMsg)
	}
}

func httpGet(id string) string {
	url := fmt.Sprintf("https://bpi.icodeq.com/163news?index=%s", id)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}
