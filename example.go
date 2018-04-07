package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "./queriableHtml"
)

func GetRawHtml(url string) ([]byte, error) {
    resp, err := http.Get(url)

    var body_bytes []byte
    body_bytes, _ = ioutil.ReadAll(resp.Body)

    return body_bytes, err
}

func main() {
    body_bytes, _ := GetRawHtml("https://golang.org/")

    root := queriableHtml.NewQueriableHtml(body_bytes)

    fmt.Println(len(root.Query([]string{"*...", "Attr,href,/dl/"})))
    fmt.Println(len(root.Query([]string{"*", "*"})))
    fmt.Println(len(root.Query([]string{"*...", "Attr,id,learn", "Atom,div"})))
    fmt.Println(len(root.Query([]string{"Atom,html", "Atom,body", "*...", "Atom,a"})))
    fmt.Println(len(root.Query([]string{"Atom,html", "Atom,body", "*", "Atom,a"})))
}
