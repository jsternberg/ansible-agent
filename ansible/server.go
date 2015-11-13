package ansible

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"

	"github.com/go-martini/martini"
)

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Serve(l net.Listener) error {
	r := martini.NewRouter()
	r.Get("/ping", s.Ping)
	r.Post("/exec", s.ExecCommand)
	r.Put("/upload", s.PutFile)

	m := martini.New()
	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return http.Serve(l, m)
}

func (s *Server) Ping() []byte {
	serverInfo := map[string]string{}
	out, _ := json.Marshal(&serverInfo)
	return out
}

func (s *Server) ExecCommand(req *http.Request) (int, interface{}) {
	command := req.FormValue("command")
	if command == "" {
		return 500, "command is a required parameter\n"
	}

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()

	data := map[string]interface{}{}
	if err != nil {
		data["status"] = 1
	} else {
		data["status"] = 0
	}
	data["stdin"] = ""
	data["stdout"] = stdout.String()
	data["stderr"] = stderr.String()

	out, err := json.Marshal(&data)
	if err != nil {
		return 500, err.Error()
	}
	return 200, out
}

func (s *Server) PutFile(req *http.Request) (int, string) {
	dest := req.FormValue("dest")
	src, _, err := req.FormFile("src")
	if err != nil {
		return 500, err.Error()
	}

	f, err := os.Create(dest)
	if err != nil {
		return 500, err.Error()
	}
	defer f.Close()

	io.Copy(f, src)
	return 200, ""
}
