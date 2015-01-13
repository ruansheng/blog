package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"database/sql" 
    _"./mysql"
    "./session"
    ."./ImageCode"
    "strconv"
)

var globalSessions *session.Manager
var sessionHandler *session.MemSessionStore

func init() {
    globalSessions, _ = session.NewManager("memory", `{"cookieName":"gosessionid","gclifetime":3600}`)
    go globalSessions.GC()
}

/**
* 定义Mysql类
*/
type Mysql struct {
	host string
	port string
	user string
	passwd string
	db string
	conn *sql.DB
}

/**
* 初始化变量
*/
func (mysqlClass * Mysql) initVar(){
	if mysqlClass.host==""{
		mysqlClass.host="localhost"
	}
	if mysqlClass.port==""{
		mysqlClass.port="3306"
	}
	if mysqlClass.user==""{
		mysqlClass.user="root"
	}
	if mysqlClass.passwd==""{
		mysqlClass.passwd="19900918"
	}
	if mysqlClass.db==""{
		mysqlClass.db="test"
	}
}

/**
* 连接mysql
*/
func (mysqlClass *Mysql) connect(db string)(*Mysql){
	mysqlClass.db=db
	mysqlClass.initVar()
	str:=mysqlClass.user+":"+mysqlClass.passwd+"@tcp("+mysqlClass.host+":"+mysqlClass.port+")/"+mysqlClass.db+"?charset=utf8"
	conn, err := sql.Open("mysql", str)
	if err!=nil{
		panic(err.Error())
	}
	mysqlClass.conn=conn
	return mysqlClass
}

/**
* 查询
*/
func (mysqlClass *Mysql) selectSql(sql string)(*sql.Rows){
	rows, err := mysqlClass.conn.Query(sql)
	if err != nil {
        panic(err.Error())
    }
    return rows
}

/**
* 关闭资源
*/
func (mysqlClass * Mysql)closeDb(){
	mysqlClass.conn.Close()
}

type Item struct{
	Id int
	Title string
//	Person map[string]interface{}
}

/**
* 首页
*/
func index(w http.ResponseWriter,r *http.Request){
    conn:=new(Mysql)
	var sql string="select goods_id,activity_id,goods_name from qg_goods limit 999"
	rows:=conn.connect("qgzs").selectSql(sql)
	var goods_id int 
    var activity_id int
    var goods_name string
    
    var list [1000]Item
    var i int=0
	for rows.Next() { 
        rerr := rows.Scan(&goods_id, &activity_id,&goods_name)
        if rerr == nil {
        		var item Item
        		item.Id=goods_id
        		item.Title=goods_name
           	list[i]=item
           	i++
        }
    }
	t,_ :=template.ParseFiles("index.html")
	t.Execute(w,list)
}

/**
* 文章内容页
*/
func article(w http.ResponseWriter,r *http.Request){
	r.ParseForm()
	id:=r.Form["article_id"][0]
	conn:=new(Mysql)
	var sql string="select goods_id,activity_id,goods_name from qg_goods where goods_id="+id+" limit 1"
	rows:=conn.connect("qgzs").selectSql(sql)
	var goods_id int 
    var activity_id int
    var goods_name string
    articleInfo:=make(map[string]interface{})
    for rows.Next() { 
        rerr := rows.Scan(&goods_id, &activity_id,&goods_name)
        if rerr == nil {
			articleInfo["Id"]=goods_id
			articleInfo["Title"]=goods_name
        }
    }
	t,_ :=template.ParseFiles("article.html")
	t.Execute(w,articleInfo)
}

/**
* 后台登录
*/
func login(w http.ResponseWriter,r *http.Request){
	t,_ :=template.ParseFiles("login.html")
	t.Execute(w,nil)
}

/**
* 登录处理
*/
func doLogin(w http.ResponseWriter,r *http.Request){
	fmt.Println(r.Method)
	imagecode:=getSession(w,r,"imagecode")
	fmt.Println(imagecode)
	t,_ :=template.ParseFiles("admin/admin.html")
	t.Execute(w,nil)
}

/**
* 获取验证码
*/
func getImageCode(w http.ResponseWriter,r *http.Request){
	d := make([]byte, 4)
	s := NewLen(4)
	ss := ""
	d = []byte(s)
	for v := range d {
		d[v] %= 10
		ss += strconv.FormatInt(int64(d[v]), 32)
	}
	w.Header().Set("Content-Type", "image/png")
	NewImage(d, 100, 40).WriteTo(w)
	setSession(w,r,"imagecode",ss)
}

/**
* set Session
*/
func setSession(w http.ResponseWriter,r *http.Request,key string,value string) bool{
	sess,_:= globalSessions.SessionStart(w, r)
    defer sess.SessionRelease(w)
    sess.Set(key,value)
    return true
}

/**
* get Session
*/
func getSession(w http.ResponseWriter,r *http.Request,key string) (interface{}){
	sess,_:= globalSessions.SessionStart(w, r)
    defer sess.SessionRelease(w)
    value := sess.Get(key)
    fmt.Println(value)
    return value
}

/**
* 入口函数
*/
func main(){
	http.HandleFunc("/",index)
	http.HandleFunc("/article",article)
	http.HandleFunc("/login",login)
	http.HandleFunc("/image",getImageCode)
	http.HandleFunc("/doLogin",doLogin)
	
	http.Handle("/public/",http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	
	err:=http.ListenAndServe(":9999",nil)
	if err!=nil{
		log.Fatal("ListenAndServe:",err)
	}
}
