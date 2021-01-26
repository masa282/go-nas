package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
	port             = ":8080"
)

func RecvHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(defaultMaxMemory)
	fhds := r.MultipartForm.File["image"]
	for _, fhd := range fhds {
		src, err := fhd.Open()
		defer src.Close()
		if err != nil {
			log.Println(err)
			return
		}

		dst, err := os.Create(filepath.Join("data", fhd.Filename))
		defer dst.Close()
		if err != nil {
			log.Println(err)
			return
		}

		data, err := ioutil.ReadAll(src)
		if err != nil {
			log.Println(err)
			return
		}
		dst.Write(data)
	}

	fmt.Println("Success!")
	w.Write([]byte("成功！"))
}

func main() {
	http.HandleFunc("/recv", RecvHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Someone Acessed!")
		http.ServeFile(w, r, filepath.Join("html", "recv.html"))
	})

	var ip_port bytes.Buffer
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			switch ipv4[0] {
			case 192:
				myIP := ipv4.String()
				ip_port.WriteString(myIP)
				ip_port.WriteString(port)
			case 172:
				myIP := ipv4.String()
				ip_port.WriteString(myIP) //ip_port.WriteByte()
				ip_port.WriteString(port)
			}
		}
	}

	fmt.Println(ip_port.String())
	if err := http.ListenAndServe(ip_port.String(), nil); err != nil {
		log.Println("falied to start")
	}
}
