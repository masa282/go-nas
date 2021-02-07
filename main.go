package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
	port             = ":8080"
)

var dirname = filepath.Join(".", "data")

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch t.filename {
	case "send.html":
		t.once.Do(func() {
			t.templ = template.Must(template.ParseFiles(filepath.Join("html", t.filename)))
		})

		files, err := ioutil.ReadDir(dirname)
		if err != nil {
			log.Println("[-]Failed to read the directory: ", dirname)
			w.Write([]byte("Failed to read the directory"))
			return
		}

		FileList := make(map[string]string, len(files))
		for _, file := range files {
			//fmt.Println(file.Name())
			FileList[file.Name()] = path.Join(dirname, file.Name())
		}

		fmt.Println(FileList)
		if err := t.templ.Execute(w, FileList); err != nil {
			log.Println("[-]temple.Execute: ", err)
		}
	}
}

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

		io.Copy(dst, src)
	}
	fmt.Println("[+]Success!")
	w.Write([]byte("成功！"))
}

func DataHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	fmt.Println(path)
	dir, _ := os.Getwd()
	//fmt.Println(filepath.FromSlash(path))
	//fmt.Println(filepath.Join(dir, path))
	//file, err := os.Open(filepath.Join(dir, filepath.FromSlash(path)))
	file, err := os.Open(filepath.Join(dir, path))
	defer file.Close()
	if err != nil {
		log.Println("[-]Failed to Open the file: ", err)
	}
	io.Copy(w, file)
}

func main() {
	http.HandleFunc("/data/", DataHandler)
	http.Handle("/send", &templateHandler{filename: "send.html"})
	http.HandleFunc("/recv", RecvHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Someone Acessed!")
		http.ServeFile(w, r, filepath.Join("html", "main.html"))
	})

	var ip_port bytes.Buffer
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	//fmt.Println(addrs)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			switch ipv4[0] {
			case 192:
				myIP := ipv4.String()
				ip_port.WriteString(myIP)
				ip_port.WriteString(port)
				goto out
			case 172:
				myIP := ipv4.String()
				ip_port.WriteString(myIP) //ip_port.WriteByte()
				ip_port.WriteString(port)
				goto out
			}
		}
	}

out:
	fmt.Println(ip_port.String())

	//if err := http.ListenAndServe("240d:1a:6b3:3d00:94de:2131:3349:f3a0", nil); err != nil {
	if err := http.ListenAndServe(ip_port.String(), nil); err != nil {
		log.Println("[-]Falied to start")
	}
}
