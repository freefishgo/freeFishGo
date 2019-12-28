package freeFishGo

import (
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
)

// DefaultApplicationBuilder is the default ApplicationBuilder used by Serve.
var DefaultApplicationBuilder *ApplicationBuilder
var DefaultConfig *Config

func checkDefaultApplicationBuilderNil() {
	if DefaultApplicationBuilder == nil {
		DefaultApplicationBuilder = NewFreeFishApplicationBuilder()
	}
	if DefaultConfig == nil {
		DefaultConfig = NewConfig()
	}
}

// ApplicationBuilder管道构造器
type ApplicationBuilder struct {
	Config  *Config
	handler *applicationHandler
}

// 向管道注入session去数据的接口
func (app *ApplicationBuilder) InjectionSession(session ISession) {
	app.handler.session = session
}

// 向默认管道注入session去数据的接口
func InjectionSession(session ISession) {
	checkDefaultApplicationBuilderNil()
	DefaultApplicationBuilder.InjectionSession(session)
}

// 创建一个ApplicationBuilder管道
func NewFreeFishApplicationBuilder() *ApplicationBuilder {
	freeFish := new(ApplicationBuilder)
	freeFish.handler = newApplicationHandler()
	freeFish.Config = NewConfig()
	return freeFish
}

// 启动默认中间件web服务
func Run() <-chan error {
	checkDefaultApplicationBuilderNil()
	return DefaultApplicationBuilder.Run()
}

// 启动web服务
func (app *ApplicationBuilder) Run() <-chan error {
	app.middlewareSorting()
	app.handler.config = app.Config
	errChan := make(chan error)
	if app.Config.EnableSession {
		if app.handler.session == nil {
			app.handler.session = NewSessionMgr(app.handler.config.SessionAliveTime)
		}
		app.handler.session.Init(app.handler.config.SessionAliveTime)
	}
	if app.Config.Listen.EnableHTTP {
		addr := app.Config.Listen.HTTPAddr + ":" + strconv.Itoa(app.Config.Listen.HTTPPort)
		go func() {
			log.Println("http on " + addr)
			errChan <- (&http.Server{
				Addr:           addr,
				ReadTimeout:    app.Config.Listen.ServerTimeOut,
				WriteTimeout:   app.Config.Listen.WriteTimeout,
				MaxHeaderBytes: app.Config.Listen.MaxHeaderBytes,
				Handler:        app.handler,
			}).ListenAndServe()
		}()
	}
	if app.Config.Listen.EnableHTTPS {
		addr := app.Config.Listen.HTTPSAddr + ":" + strconv.Itoa(app.Config.Listen.HTTPSPort)
		go func() {
			log.Println("https on " + addr)
			errChan <- (&http.Server{
				Addr:           addr,
				ReadTimeout:    app.Config.Listen.ServerTimeOut,
				WriteTimeout:   app.Config.Listen.WriteTimeout,
				MaxHeaderBytes: app.Config.Listen.MaxHeaderBytes,
				Handler:        app.handler,
			}).ListenAndServeTLS(app.Config.Listen.HTTPSCertFile, app.Config.Listen.HTTPSKeyFile)
		}()
	}
	return errChan
}

func newApplicationHandler() *applicationHandler {
	return new(applicationHandler)
}

type applicationHandler struct {
	middlewareList []IMiddleware
	middlewareLink *MiddlewareLink
	config         *Config
	session        ISession
}

// http服务逻辑处理程序
func (app *applicationHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := new(HttpContext)
	ctx.SetContext(rw, r)
	if app.config.EnableSession {
		ctx.Response.SetISession(app.session)
		ctx.Response.SessionCookieName = app.config.SessionCookieName
		ctx.Response.SessionAliveTime = app.config.SessionAliveTime
		cookie, err := ctx.Request.Cookie(app.config.SessionCookieName)
		if err == nil {
			ctx.Response.SessionId = cookie.Value
		}
	}
	defer func() {
		if ctx != nil && ctx.Response.Gzip != nil {
			ctx.Response.Gzip.Close()
		}
	}()
	defer func() {
		if app.config.EnableSession {
			ctx.Response.UpdateSession()
		}
		if err := recover(); err != nil {
			err, _ := err.(error)
			if app.config.RecoverPanic {
				//app.config.RecoverFunc(ctx, err, debug.Stack())
			} else {
				if ctx != nil {
					ctx.Response.WriteHeader(500)
					ctx.Response.Write([]byte(`<html><body><div style="color: red;color: red;margin: 150px auto;width: 800px;"><div>` + "服务器内部错误 500:" + err.Error() + "\r\n\r\n\r\n</div><pre>" + string(debug.Stack()) + `</pre></div></body></html>`))
				}
			}
		}
	}()
	ctx.Response.IsOpenGzip = app.config.EnableGzip
	ctx.Response.NeedGzipLen = app.config.NeedGzipLen
	ctx = app.middlewareLink.val.Middleware(ctx, app.middlewareLink.next.innerNext)
	if !ctx.Response.Started {
		ctx.Response.ResponseWriter.WriteHeader(ctx.Response.ReadStatusCode())
	}
}

// 下一个中间件
type Next func(*HttpContext) *HttpContext

// 中间件类型接口
type IMiddleware interface {
	Middleware(ctx *HttpContext, next Next) *HttpContext
	//注册框架后 框架会自动调用这个函数
	LastInit(*Config)
}

type MiddlewareLink struct {
	val  IMiddleware
	next *MiddlewareLink
}

// 执行下一个中间件
func (link *MiddlewareLink) innerNext(ctx *HttpContext) *HttpContext {
	return link.val.Middleware(ctx, link.next.innerNext)
}

// 中间件注册接口
func (app *ApplicationBuilder) UseMiddleware(middleware IMiddleware) {
	if app.handler.middlewareList == nil {
		app.handler.middlewareList = []IMiddleware{}
	}
	app.handler.middlewareList = append(app.handler.middlewareList, middleware)
}

// 向默认中间件注册接口
func UseMiddleware(middleware IMiddleware) {
	checkDefaultApplicationBuilderNil()
	DefaultApplicationBuilder.UseMiddleware(middleware)
}

// 中间件排序
func (app *ApplicationBuilder) middlewareSorting() *ApplicationBuilder {
	app.handler.middlewareLink = new(MiddlewareLink)
	tmpMid := app.handler.middlewareLink
	for i := 0; i < len(app.handler.middlewareList); i++ {
		tmpMid.val = app.handler.middlewareList[i]
		tmpMid.val.LastInit(app.Config)
		tmpMid.next = new(MiddlewareLink)
		tmpMid = tmpMid.next
	}
	if tmpMid.val == nil {
		tmpMid.val = &lastFrameMiddleware{}
		tmpMid.val.LastInit(app.Config)
	}
	return app
}

// 框架最后一个中间件
type lastFrameMiddleware struct {
}

func (last *lastFrameMiddleware) Middleware(ctx *HttpContext, next Next) *HttpContext {
	return ctx
}
func (last *lastFrameMiddleware) LastInit(config *Config) {

}
