// Copyright 2012 by sdm. All rights reserved.
// license that can be found in the LICENSE file.

package wk

import (
	"log"
	"net"
	"net/http"
	//_ "net/http/pprof"
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

// http server
type HttpServer struct {
	// config
	Config *WebConfig

	// net.Listener 
	Listener net.Listener

	// *ServeMux
	Mux *http.ServeMux

	// http server
	server *http.Server

	// Processes
	Processes ProcessTable

	// RouteTable
	RouteTable *RouteTable

	//server variables
	Variables map[string]interface{}
}

// Fire can fire a event
func (srv *HttpServer) Fire(moudle, name string, source, data interface{}, context *HttpContext) {
	if LogLevel >= LogInfo {
		Logger.Println("fire event", moudle, name)
	}

	var e *EventContext

	for _, sub := range Subscribers {
		if (sub.Moudle == _any || sub.Moudle == moudle) && (sub.Name == _any || sub.Name == name) {
			if e == nil {
				e = &EventContext{
					Moudle:  moudle,
					Name:    name,
					Source:  source,
					Data:    data,
					Context: context,
				}
			}

			sub.Handler.On(e)
		}
	}
}

// DefaultServer create a http server with default config
func NewDefaultServer() (srv *HttpServer, err error) {
	var conf *WebConfig

	conf, err = ReadDefaultConfigFile()
	if err != nil {
		conf = NewDefaultConfig()
	}

	return NewHttpServer(conf)
}

// NewHttpServer create a http server with config 
func NewHttpServer(config *WebConfig) (srv *HttpServer, err error) {
	srv = &HttpServer{
		Config: config,
	}
	srv.init()
	return srv, nil
}

func (srv *HttpServer) init() error {
	srv.Variables = make(map[string]interface{})
	srv.RouteTable = newRouteTable()

	// copy hander, maybe does not need this?	
	l := len(Processes)
	srv.Processes = make([]*Process, l)
	for i := 0; i < l; i++ {
		Processes[i].Handler.Register(srv)
		srv.Processes[i] = Processes[i]
	}

	return nil
}

func (srv *HttpServer) listenAndServe() (err error) {
	// if srv.Listener, err = net.Listen("tcp", srv.Config.Address); err != nil {
	// 	return err
	// }

	srv.Mux = http.NewServeMux()
	srv.Mux.Handle("/", srv)
	srv.server = &http.Server{
		Addr:           srv.Config.Address,
		Handler:        srv.Mux,
		ReadTimeout:    time.Duration(srv.Config.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(srv.Config.WriteTimeout) * time.Second,
		MaxHeaderBytes: srv.Config.MaxHeaderBytes,
	}

	return srv.server.ListenAndServe()
}

// error return error message to client
func (srv *HttpServer) error(ctx *HttpContext, err error) {
	if LogLevel >= LogError {
		Logger.Println(err.Error())
		Logger.Println(debug.Stack())
	}

	ctx.Result = &ErrorResult{
		Tag: "HttpServer",
		Err: err,
	}
}

// Start can start server instance and serve request
func (srv *HttpServer) Start() (err error) {

	if Logger == nil {
		Logger = log.New(os.Stdout, _serverName, log.Ldate|log.Ltime)
	}
	Logger.Println("http server is starting")

	Logger.Println("Address:", srv.Config.Address,
		"\n\t RootDir:", srv.Config.RootDir,
		"\n\t ConfigDir:", srv.Config.ConfigDir,
		"\n\t PublicDir:", srv.Config.PublicDir,
	)

	if err = srv.listenAndServe(); err != nil {
		Logger.Println("http server server fail:", err)
		return
	}

	Logger.Println("http server server is listen on:", srv.Config.Address)
	return nil
}

// // Close can close http server
// func (s *HttpServer) Close() error {
// 	Logger.Println("http server is closing")

// 	if srv.Listener != nil {
// 		srv.Listener.Close()
// 	}
// 	Logger.Println("http server closed")

// 	return nil
// }

// ServeHTTP 
func (srv *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer func() {
		err := recover()
		if err == nil {
			return
		}

		var buf bytes.Buffer
		fmt.Fprintf(&buf, "http server panic %v : %v\n", r.URL, err)
		buf.Write(debug.Stack())
		Logger.Println(buf.String())

		http.Error(w, msgServerInternalErr, codeServerInternaError)

	}()

	ctx := srv.buildContext(w, r)

	if LogLevel >= LogDebug {
		Logger.Println("request start", ctx.Method, ctx.RequestPath)
	}
	srv.Fire(_wkWebServer, _eventStartRequest, srv, nil, ctx)

	srv.doServer(ctx)

	srv.Fire(_wkWebServer, _eventEndRequest, srv, nil, ctx)
	if LogLevel >= LogDebug {
		Logger.Println("request end", ctx.Method, ctx.RequestPath, ctx.Result, ctx.Error)
	}

}

func (srv *HttpServer) doServer(ctx *HttpContext) {

	ctx.SetHeader(HeaderServer, _serverName)

	for _, h := range srv.Processes {
		if h.match(ctx) {
			srv.exeProcess(ctx, h)
		}
	}
}

// buildContext 
func (s *HttpServer) buildContext(w http.ResponseWriter, r *http.Request) *HttpContext {
	_ = r.ParseForm()
	return &HttpContext{
		Resonse:     w,
		Request:     r,
		Method:      r.Method,
		RequestPath: cleanPath(strings.TrimSpace(r.URL.Path)),
	}
}

// execute process
func (srv *HttpServer) exeProcess(ctx *HttpContext, p *Process) (err error) {

	if LogLevel >= LogDebug {
		Logger.Println("process start:", p.Name)
	}

	defer func() {
		if x := recover(); x != nil {
			if LogLevel >= LogError {
				Logger.Println("execute process recover:", p.Name, x)
				Logger.Println(string(debug.Stack()))
			}

			if e, ok := x.(error); ok {
				err = e
			} else {
				err = errors.New(fmt.Sprintln(x))
			}
			ctx.Error = err
		}
	}()

	srv.Fire(p.Name, _eventStartExecute, p, nil, ctx)

	p.Handler.Execute(ctx)

	srv.Fire(p.Name, _eventEndExecute, p, nil, ctx)

	if LogLevel >= LogDebug {
		Logger.Println("process end", p.Name, err)
	}

	return nil
}

// // MapPath return physical path	 
// func (srv *HttpServer) MapPath(p string) string {

// 	f := path.Join(srv.Config.PublicDir, p)
// 	info, err := os.Stat(f)
// 	if err != nil || info.IsDir() {
// 		return ""
// 	}
// 	return f
// }
