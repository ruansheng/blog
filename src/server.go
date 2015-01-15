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
    "crypto/sha1"
    "io"
)

/**
* 管理员id session key
*/
const manageIdKey="manage_id"

/**
* 图片验证码 session key
*/
const imageCodeKey="image_code"

/**
* session 资源句柄
*/
var globalSessions *session.Manager
var sessionHandler *session.MemSessionStore

/**
* 初始化函数
*/
func init() {
    globalSessions, _ = session.NewManager("file",`{"cookieName":"gosessionid","gclifetime":3600,"ProviderConfig":"./tmp"}`)
    go globalSessions.GC()
}

/**
* 文章 数据结构
*/
type Item struct{
	article_id string
	title string
	content string
	create_time string
	show_num string
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

/**
* 首页
*/
func index(w http.ResponseWriter,r *http.Request){
	r.ParseForm()
	page:=r.Form["page"][0]
	if page=="" {
		page="1"
	}
	
	intPage,_:=strconv.Atoi(page)
	limit:=strconv.Itoa(intPage*20)+",20"
	
    conn:=new(Mysql)
	var sql string="select article_id,title,content,create_time,show_num from article limit "+limit
	fmt.Println(sql)
	fmt.Println(conn)
	
	/*
	rows:=conn.connect("blog").selectSql(sql)
	var article_id string 
    var title string
    var content string
    var create_time string
    var show_num string
    
    var list [20]Item
    var i int=0
	for rows.Next() { 
        rerr := rows.Scan(&article_id, &title,&content,&create_time,&show_num)
        if rerr == nil {
        		var item Item
        		item.article_id=article_id
        		item.title=title
        		item.content=content
        		item.create_time=create_time
        		item.show_num=show_num
           	list[i]=item
           	i++
        }
    }
	t,_ :=template.ParseFiles("index.html")
	t.Execute(w,list)
	*/
}

/**
* 文章内容页
*/
func article(w http.ResponseWriter,r *http.Request){
//	r.ParseForm()
//	id:=r.Form["article_id"][0]
//	conn:=new(Mysql)
//	var sql string="select goods_id,activity_id,goods_name from qg_goods where goods_id="+id+" limit 1"
//	rows:=conn.connect("qgzs").selectSql(sql)
//	var goods_id int 
//    var activity_id int
//    var goods_name string
//    articleInfo:=make(map[string]interface{})
//    for rows.Next() { 
//        rerr := rows.Scan(&goods_id, &activity_id,&goods_name)
//        if rerr == nil {
//			articleInfo["Id"]=goods_id
//			articleInfo["Title"]=goods_name
//        }
//    }
//	t,_ :=template.ParseFiles("article.html")
//	t.Execute(w,articleInfo)
}

/**
* 后台登录
*/
func login(w http.ResponseWriter,r *http.Request){
	data:=map[string]string{"msg":""}
	t,_ :=template.ParseFiles("login.html")
	t.Execute(w,data)
}

/**
* 登录处理
*/
func doLogin(w http.ResponseWriter,r *http.Request){
	r.ParseForm()
	manageName:=r.Form["manage_name"][0]
	password:=r.Form["password"][0]
	code:=r.Form["code"][0]
	
	if manageName==""||password==""||code==""{
		data:=map[string]string{"msg":"登录信息不能为空"}
		t,_ :=template.ParseFiles("login.html")
		t.Execute(w,data)
	}
	
	//判断验证码是否正确
	imagecode:=getSession(w,r,imageCodeKey)
	if code==imagecode {
		//判断账号和密码是否正确
		userInfo,err:=checkLogin(manageName,getSha1(password))
		if err!=nil {
			data:=map[string]interface{}{"msg":err}
			t,_ :=template.ParseFiles("login.html")
			t.Execute(w,data)
		}else{
			setSession(w,r,manageIdKey,userInfo["manage_id"])
			t,_ :=template.ParseFiles("admin/admin.html")
			t.Execute(w,nil)
		}
	}else{
		data:=map[string]string{"msg":"验证码错误"}
		t,_ :=template.ParseFiles("login.html")
		t.Execute(w,data)
	}
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
	setSession(w,r,imageCodeKey,ss)
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
    return value
}

/**
* 校验登录
*/
func checkLogin(manageName string,password string)(map[string]string,interface{}){
	conn:=new(Mysql)
	var sql string="select manage_id,manage_name from manage where manage_name='"+manageName+"' and password='"+password+"' limit 1"
	rows:=conn.connect("blog").selectSql(sql)
	var manage_id string 
    var manage_name string
    userInfo:=make(map[string]string)
    for rows.Next() { 
        rerr := rows.Scan(&manage_id, &manage_name)
        if rerr == nil {
			userInfo["manage_id"]=manage_id
			userInfo["manage_name"]=manage_name
        }
    }
    var err interface{}
    if userInfo!=nil{
    		err=nil
    }else{
    		err="账号或密码不正确"
    	}
    return userInfo,err
}

/**
* 对字符串进行SHA1哈希
*/
func getSha1(str string) string {
	t := sha1.New();
	io.WriteString(t,str);
	return fmt.Sprintf("%x",t.Sum(nil));
}

//-----------------------------------------------

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
