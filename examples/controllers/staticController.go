package controllers

import (
	"freeFishGo/examples/fishgo"
	"freeFishGo/router"
	"io/ioutil"
	"log"
	"path/filepath"
)

type staticController struct {
	router.Controller
}

func init() {
	fishgo.AddHandlers(&staticController{})
}

type data struct {
	Path string `json:"path"`
}

func (static *staticController) StaticFile(d *data) {
	if b, err := ioutil.ReadFile(filepath.Join("static", d.Path)); err == nil {
		static.HttpContext.Response.Write(b)
	} else {
		static.HttpContext.Response.WriteHeader(404)
		static.HttpContext.Response.Write([]byte(err.Error()))
	}
}

func (static *staticController) SetInfo() []*router.ControllerActionInfo {
	log.Println("不是默认GetControllerInfo")
	tmp := make([]*router.ControllerActionInfo, 0)
	tmp = append(tmp, &router.ControllerActionInfo{RouterPattern: "static/{path:allString}", ControllerActionFuncName: "StaticFile"})
	return tmp
}
