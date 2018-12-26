/**
 *                            ___.
 * ____________ _______   ____\_ |__   ____   ____
 * \_  __ \__  \\_  __ \_/ __ \| __ \_/ __ \_/ __ \
 *  |  | \// __ \|  | \/\  ___/| \_\ \  ___/\  ___/
 *  |__|  (____  /__|    \___  >___  /\___  >\___  >
 *            \/            \/    \/     \/     \/
 *
 * Licensed under The Apache Licence
 * For full copyright and license information, please see the LICENSE
 * Redistributions of files must retain the above copyright notice.
 *
 * @author      rare bee<rarebee@163.com>
 * @copyright   rare bee<rarebee@163.com>
 * @license     http://www.apache.org/licenses/LICENSE-2.0.html Apache Licence, Version 2.0
 */
package main

import (
	"TKServer/Config"
	"TKServer/Core"
	"TKServer/Handle"
	"TKServer/Helper"
	_session "github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"log"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(8)

	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	db := Helper.NewDbHubHandle()
	defer db.CloseAll()

	cache := Helper.InitRedisPool()
	defer cache.CloseAll()

	server, err := Core.NewTKServer()

	go server.Logs()

	if err != nil {

	}

	server.HandleConnect(func(s *Core.Session) {
		log.Println("HandleConnect:", s.Sid)
	})

	server.HandleMessage(func(s *Core.Session, msg []byte) {
		// log.Println("HandleMessage:", string(msg))
		s.OnExchange(msg)
	})

	server.HandleSentMessage(func(s *Core.Session, msg []byte) {
		// fmt.Println("HandleSentMessage", s, string(msg))
	})

	server.HandlePong(func(s *Core.Session) {
		log.Println("HandlePong:", s.Id())
	})

	server.HandleClose(func(s *Core.Session, code int, msg string) error {
		log.Println("HandleClose:", s, code, msg)
		return nil
	})

	server.HandleError(func(s *Core.Session, err error) {
		log.Println("HandleError:", err)
	})

	server.HandleDisconnect(func(s *Core.Session) {
		// fmt.Println("HandleDisconnect:", s)
	})

	// Http api start from here!
	cookieDomain := Config.DomainProd
	if Config.Debug {
		cookieDomain = Config.DomainDev
	}

	// MiddleWage Recovery
	r.Use(gin.Recovery())

	// MiddleWare Session
	store := cookie.NewStore([]byte("loveCode"))
	store.Options(_session.Options{Path: "/", Domain: cookieDomain})
	r.Use(_session.Sessions("TSession", store))

	// Set Origin
	r.Use(func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", ctx.Request.Header.Get("Origin"))
		ctx.Header("Access-Control-Allow-Credentials", "true")
	})

	// Set template parse options
	r.Delims("{[{", "}]}")
	r.LoadHTMLFiles("./Html/views.html", "./Html/index.html", "./Html/customer.html", "./Html/index_test.html", "./Html/customer_test.html")

	// [ Router ] /io
	r.GET("/io", func(c *gin.Context) {
		server.ServeHTTP(c)
	})

	// Api Router
	Api := &Handle.ApiHandle{}

	r.GET("/", Api.Index())
	r.GET("/getSessions", Api.GetSessions(server))
	r.GET("/getGroups", Api.GetGroups(server))
	r.GET("/getSession", Api.GetSession(server))
	r.GET("/getUser/:sid", Api.GetUser(server))
	r.GET("/getWXConfig", Api.GetWXConfig())
	r.GET("/customer", Api.Customer())
	r.GET("/views", Api.Views())
	r.GET("/static/*file", Api.Static())
	r.GET("/voiceConvert", Api.VoiceConvert())
	r.GET("/viewNotify", Api.ViewNotify(server))
	// r.GET("/getFriend", func(ctx *gin.Context) {})
	r.GET("/getBusiness", Api.GetBusiness())
	r.GET("/getCustomer", Api.GetCustomer())
	r.GET("/getUserCustomer", Api.GetUserCustomer())
	r.GET("/getFriends", Api.GetFriends())
	r.GET("/readAck", Api.ReadAck())
	r.GET("/setNewCustomerAck", Api.SetNewCustomerAck())
	r.GET("/getNewCustomerCount", Api.GetNewCustomerCount())
	r.GET("/SetInterested", Api.SetInterested())
	r.GET("/GetUserNewMessage", Api.GetUserNewMessage())
	r.GET("/getLineStatus", Api.GetLineStatus(server))
	r.POST("/sendMessage", Api.SendMessage(server))
	r.GET("/getJsConfig", Api.GetJsConfig())

	r.Run(":" + Config.Port)
}