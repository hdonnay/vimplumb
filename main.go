package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"9fans.net/go/plan9/client"
	"9fans.net/go/plumb"
)

func main() {
	flag.Parse()

	srv := exec.Command("vim", "--servername", client.Namespace())
	srv.Stdin = os.Stdin
	srv.Stdout = os.Stdout
	srv.Stderr = os.Stderr
	srv.Args = append(srv.Args, flag.Args()...)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}

	p, err := plumb.Open("edit", 0)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()
	go func() {
		m := &plumb.Message{}
		rd := bufio.NewReader(p)
		for {
			err := m.Recv(rd)
			switch err {
			case nil:
			case io.EOF:
				return
			default:
				log.Println(err)
				return
			}
			var a string
			for p := m.Attr; p != nil; p = p.Next {
				switch p.Name {
				case "addr":
					a = p.Value
					if strings.HasPrefix(a, "/") && !strings.HasSuffix(a, "/") {
						a += "/"
					}
				}
			}
			vim(string(m.Data), a)
		}
	}()
	srv.Wait()
}

func vim(file, addr string) error {
	cmd := exec.Command("vim", "--servername", client.Namespace(), "--remote-tab")
	if addr != "" {
		cmd.Args = append(cmd.Args, "+"+addr)
	}
	cmd.Args = append(cmd.Args, file)
	return cmd.Run()
}
