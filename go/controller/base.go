package controller

import (
	"context"
	"fmt"

	"server_golang/common/json"
	"server_golang/common/types"
)

// Result 是控制器方法的返回结果
type Result struct {
	SendUser interface{} // "me"=发给自己, "all"=广播, []int64=指定用户
	Msg      string
}

// BaseController 控制器基类，包含通用方法
type BaseController struct {
	IP     string
	Params types.Map
}

func NewBaseController(params types.Map, ip string) *BaseController {
	return &BaseController{
		Params: params,
		IP:     ip,
	}
}

func (c *BaseController) ResponseSuccessToMe(message types.Map) *Result {
	return &Result{
		SendUser: "me",
		Msg:      c.getResultSuccess(message),
	}
}

func (c *BaseController) ResponseSuccessToAll(message types.Map) *Result {
	return &Result{
		SendUser: "all",
		Msg:      c.getResultSuccess(message),
	}
}

func (c *BaseController) ResponseSuccessToUids(uids []int64, message types.Map) *Result {
	return &Result{
		SendUser: uids,
		Msg:      c.getResultSuccess(message),
	}
}

func (c *BaseController) ResponseError(code int, message string) *Result {
	data := types.Map{
		"c":     c.Params.GetStringE("c"),
		"m":     c.Params.GetStringE("m"),
		"reqid": c.Params.GetStringE("reqid"),
		"code":  code,
		"msg":   message,
	}
	return &Result{
		SendUser: "me",
		Msg:      json.Marshal(data),
	}
}

func (c *BaseController) getResultSuccess(message types.Map) string {
	if message == nil {
		message = make(types.Map)
	}
	message["c"] = c.Params.GetStringE("c")
	message["m"] = c.Params.GetStringE("m")
	message["reqid"] = c.Params.GetStringE("reqid")
	message["code"] = 0
	return json.Marshal(message)
}

// Router 将请求分发到对应的控制器
type Router struct{}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Dispatch(ctx context.Context, params types.Map, ip string) *Result {
	controllerName := params.GetStringE("c")
	actionName := params.GetStringE("m")

	base := NewBaseController(params, ip)

	switch controllerName {
	case "user", "User":
		return dispatchUser(ctx, base, actionName)
	case "pay", "Pay":
		return dispatchPay(ctx, base, actionName)
	case "shinelight", "Shinelight":
		return dispatchShinelight(ctx, base, actionName)
	default:
		return base.ResponseError(3, fmt.Sprintf("route error: %s/%s", controllerName, actionName))
	}
}
