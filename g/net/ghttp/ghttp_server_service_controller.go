// Copyright 2018 gf Author(https://gitee.com/johng/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://gitee.com/johng/gf.
// 服务注册.

package ghttp

import (
    "errors"
    "strings"
    "reflect"
)

// 绑定控制器，控制器需要实现gmvc.Controller接口
// 这种方式绑定的控制器每一次请求都会初始化一个新的控制器对象进行处理，对应不同的请求会话
// 第三个参数methods用以指定需要注册的方法，支持多个方法名称，多个方法以英文“,”号分隔，区分大小写
func (s *Server)BindController(pattern string, c Controller, methods...string) error {
    methodMap := (map[string]bool)(nil)
    if len(methods) > 0 {
        methodMap = make(map[string]bool)
        for _, v := range strings.Split(methods[0], ",") {
            methodMap[strings.TrimSpace(v)] = true
        }
    }
    // 遍历控制器，获取方法列表，并构造成uri
    m     := make(handlerMap)
    v     := reflect.ValueOf(c)
    t     := v.Type()
    sname := t.Elem().Name()
    for i := 0; i < v.NumMethod(); i++ {
        mname := t.Method(i).Name
        if methodMap != nil && !methodMap[mname] {
            continue
        }
        if mname == "Init" || mname == "Shut" || mname == "Exit"  {
            continue
        }
        key   := s.mergeBuildInNameToPattern(pattern, sname, mname, true)
        m[key] = &handlerItem {
            rtype : gROUTE_REGISTER_CONTROLLER,
            ctype : v.Elem().Type(),
            fname : mname,
            faddr : nil,
        }
        // 如果方法中带有Index方法，那么额外自动增加一个路由规则匹配主URI
        if strings.EqualFold(mname, "Index") {
            p := key
            if strings.EqualFold(p[len(p) - 6:], "/index") {
                p = p[0 : len(p) - 6]
                if len(p) == 0 {
                    p = "/"
                }
            }
            m[p] = &handlerItem {
                rtype : gROUTE_REGISTER_CONTROLLER,
                ctype : v.Elem().Type(),
                fname : mname,
                faddr : nil,
            }
        }
    }
    return s.bindHandlerByMap(m)
}

// 绑定路由到指定的方法执行
func (s *Server)BindControllerMethod(pattern string, c Controller, method string) error {
    m     := make(handlerMap)
    v     := reflect.ValueOf(c)
    e     := v.Type().Elem()
    t     := v.Elem().Type()
    sname := e.Name()
    mname := strings.TrimSpace(method)
    if !v.MethodByName(mname).IsValid() {
        return errors.New("invalid method name:" + mname)
    }
    key    := s.mergeBuildInNameToPattern(pattern, sname, mname, false)
    m[key]  = &handlerItem {
        rtype : gROUTE_REGISTER_CONTROLLER,
        ctype : t,
        fname : mname,
        faddr : nil,
    }
    return s.bindHandlerByMap(m)
}

// 绑定控制器(RESTFul)，控制器需要实现gmvc.Controller接口
// 方法会识别HTTP方法，并做REST绑定处理，例如：Post方法会绑定到HTTP POST的方法请求处理，Delete方法会绑定到HTTP DELETE的方法请求处理
// 因此只会绑定HTTP Method对应的方法，其他方法不会自动注册绑定
// 这种方式绑定的控制器每一次请求都会初始化一个新的控制器对象进行处理，对应不同的请求会话
func (s *Server)BindControllerRest(pattern string, c Controller) error {
    // 遍历控制器，获取方法列表，并构造成uri
    m := make(handlerMap)
    v := reflect.ValueOf(c)
    t := v.Type()
    // 如果存在与HttpMethod对应名字的方法，那么绑定这些方法
    for i := 0; i < v.NumMethod(); i++ {
        name   := t.Method(i).Name
        method := strings.ToUpper(name)
        if _, ok := s.methodsMap[method]; !ok {
            continue
        }
        key   := name + ":" + pattern
        m[key] = &handlerItem {
            rtype : gROUTE_REGISTER_CONTROLLER,
            ctype : v.Elem().Type(),
            fname : name,
            faddr : nil,
        }
    }
    return s.bindHandlerByMap(m)
}
