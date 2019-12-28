package mvc

import (
	freeFishGo "github.com/freefishgo/freefishgo"
)

// 默认的MvcWebConfig配置
var DefaultMvcWebConfig *MvcWebConfig

// 默认的MvcApp
var DefaultMvcApp *MvcApp

type MvcApp struct {
	handlers *controllerRegister
	//Server   *http.Server
	Config *MvcWebConfig
}

// Web服务逻辑处理程序
func (mvc *MvcApp) Middleware(ctx *freeFishGo.HttpContext, next freeFishGo.Next) (c *freeFishGo.HttpContext) {
	c = ctx
	ctx = mvc.handlers.AnalysisRequest(ctx)
	return next(ctx)
}

// 框架注册完成时  进行最后的配置
func (mvc *MvcApp) LastInit(cnf *freeFishGo.Config) {
	mvc.handlers.WebConfig = mvc.Config
	mvc.handlers.MainRouterNil()
}

// 实例化生成一个Mvc对象
func NewFreeFishMvcApp() *MvcApp {
	freeFish := new(MvcApp)
	freeFish.handlers = newControllerRegister()
	freeFish.Config = freeFish.handlers.WebConfig
	return freeFish
}

func checkDefaultMvcApp() {
	if DefaultMvcApp == nil {
		DefaultMvcApp = NewFreeFishMvcApp()
	}
	if DefaultMvcWebConfig == nil {
		DefaultMvcWebConfig = NewWebConfig()
	}
	DefaultMvcApp.Config = DefaultMvcWebConfig
	DefaultMvcApp.handlers.WebConfig = DefaultMvcWebConfig
}

// 将Controller控制器注册到Mvc框架对象中 即使添加路由动作
func (app *MvcApp) AddHandlers(ic ...IController) {
	for i := 0; i < len(ic); i++ {
		app.handlers.AddHandlers(ic[i])
	}
}

// 将Controller控制器注册到默认的Mvc框架对象中 即使添加路由动作
func AddHandlers(ic ...IController) {
	checkDefaultMvcApp()
	DefaultMvcApp.AddHandlers(ic...)
}

// 主节点路由匹配原则注册     目前系统变量支持格式为 `/{ Controller}/{Action}/{id:int}/{who:string}/{allString}`
//
// 如果不进行路由注册  默认为/{ Controller}/{Action}   router.ControllerActionInfo中 ControllerActionFuncName不用设置  设置了也不会生效
func (app *MvcApp) AddMainRouter(list ...*MainRouter) {
	for _, v := range list {
		if app.Config.homeController == "" || app.Config.indexAction == "" && (v.HomeController != "" && v.IndexAction != "") {
			app.Config.homeController = v.HomeController
			app.Config.indexAction = v.IndexAction
			app.handlers.AddMainRouter(v)
		} else {
			v.IndexAction = ""
			v.HomeController = ""
			app.handlers.AddMainRouter(v)
		}
	}
}

// 默认mvc框架 主节点路由匹配原则注册     目前系统变量支持格式为 `/{ Controller}/{Action}/{id:int}/{who:string}/{allString}`
//
// 如果不进行路由注册  默认为/{ Controller}/{Action}   router.ControllerActionInfo中 ControllerActionFuncName不用设置  设置了也不会生效
func AddMainRouter(list ...*MainRouter) {
	checkDefaultMvcApp()
	DefaultMvcApp.AddMainRouter(list...)
}

// 主路由
type MainRouter struct {
	//路由设置  如：/{Controller}/{Action}/{id:int}
	// /home/index/123可以匹配成功
	RouterPattern string
	// Controller名称
	HomeController string
	// 动作名称
	IndexAction string
}
