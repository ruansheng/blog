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
    "encoding/json"
    "time"
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
	urlRow,err:=r.Form["page"]
	var page string
	if err!=true {
		page="1"
	}else{
		page=urlRow[0]
	}
	
	intPage,_:=strconv.Atoi(page)
	limit:=strconv.Itoa((intPage-1)*20)+",20"
	
	var sql string="select article_id,title,content,create_time,show_num from article limit "+limit
    conn:=new(Mysql)
	 
	rows:=conn.connect("blog").selectSql(sql)
	var article_id string 
    var title string
    var content string
    var create_time string
    var show_num string
    
    list:=make([]map[string]interface{},0,20)
	for rows.Next() { 
        rerr := rows.Scan(&article_id, &title,&content,&create_time,&show_num)
        if rerr == nil {
        		item:=make(map[string]interface{})
        		item["article_id"]=article_id
        		item["title"]=title
        		item["content"]=content
        		item["create_time"]=create_time
        		item["show_num"]=show_num
           	list=append(list,item)
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
	var sql string="select article_id,title,content,create_time,show_num from article where article_id="+id+" limit 1"
	rows:=conn.connect("blog").selectSql(sql)
	var article_id string 
    var title string
    var content string
    var create_time string
    var show_num string
    articleInfo:=make(map[string]interface{})
    for rows.Next() { 
        rerr := rows.Scan(&article_id, &title,&content,&create_time,&show_num)
        if rerr == nil {
			articleInfo["article_id"]=article_id
        		articleInfo["title"]=title
        		articleInfo["content"]=content
        		articleInfo["create_time"]=create_time
        		articleInfo["show_num"]=show_num
        }
    }
	t,_ :=template.ParseFiles("article.html")
	t.Execute(w,articleInfo)
}

/**
*  文章列表 Api
*/
func articleListApi(w http.ResponseWriter,r *http.Request){
	r.ParseForm()
	urlRow,err:=r.Form["page"]
	var page string
	if err!=true {
		page="1"
	}else{
		page=urlRow[0]
	}
	
	intPage,_:=strconv.Atoi(page)
	limit:=strconv.Itoa((intPage-1)*20)+",20"
	
	var sql string="select article_id,title,content,create_time,show_num from article limit "+limit
    conn:=new(Mysql)
	 
	rows:=conn.connect("blog").selectSql(sql)
	var article_id string 
    var title string
    var content string
    var create_time string
    var show_num string
    
    list:=make([]map[string]interface{},0,20)
	for rows.Next() { 
        rerr := rows.Scan(&article_id, &title,&content,&create_time,&show_num)
        if rerr == nil {
        		item:=make(map[string]interface{})
        		item["article_id"]=article_id
        		item["title"]=title
        		item["content"]=content
        		item["create_time"]=create_time
        		item["show_num"]=show_num
        		list=append(list,item)
        }
    }
    
    result,_:=json.Marshal(list)
    w.Write(result)
}

/**
* 权限校验 后台登录
*/
func checkManagerAuthority(w http.ResponseWriter,r *http.Request){
	manageId:=getSession(w,r,manageIdKey)
	if manageId==nil {
		io.WriteString(w,"<script type='text/javascript'>location.href='/login'</script>")
	}
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
			io.WriteString(w,"<script type='text/javascript'>location.href='/admin'</script>")
		}
	}else{
		data:=map[string]string{"msg":"验证码错误"}
		t,_ :=template.ParseFiles("login.html")
		t.Execute(w,data)
	}
}

/**
* 主面板
*/
func admin(w http.ResponseWriter,r *http.Request){
	checkManagerAuthority(w,r)
	t,_ :=template.ParseFiles("admin/admin.html")
	t.Execute(w,nil)
}

/**
* 添加文章
*/
func addArticle(w http.ResponseWriter,r *http.Request){
	checkManagerAuthority(w,r)
	data:=map[string]string{"msg":""}
	t,_ :=template.ParseFiles("admin/addArticle.html")
	t.Execute(w,data)
}

/**
* 执行插入文章
*/
func doAddArticle(w http.ResponseWriter,r *http.Request){
	r.ParseForm()
	articleTitleRow,err1:=r.Form["article_title"]
	articleContentRow,err2:=r.Form["article_content"]
	
	if err1!=true || err2!=true {
		data:=map[string]string{"msg":"文章信息不能为空"}
		t,_ :=template.ParseFiles("admin/addArticle.html")
		t.Execute(w,data)
	}
	
	sql:="insert into article(title,content,create_time,update_time,delete_time,is_del,show_num)values(?,?,?,?,?,?,?)"
	_this:=new(Mysql)
	_thisClass:=_this.connect("blog")
	stmt,err := _thisClass.conn.Prepare(sql)
	if err!=nil {
		data:=map[string]string{"msg":"prepare执行插入错误"}
		t,_ :=template.ParseFiles("admin/addArticle.html")
		t.Execute(w,data)
	}
	
	articleTitle:=articleTitleRow[0]
	articleContent:=articleContentRow[0]
	time:=time.Now().Unix()
	
	res,err := stmt.Exec(articleTitle,articleContent,time,time,time,0,0)
	if err!=nil {
		data:=map[string]string{"msg":"stmt执行插入错误"}
		t,_ :=template.ParseFiles("admin/addArticle.html")
		t.Execute(w,data)
	}
	article_id, err := res.LastInsertId()
	if err!=nil {
		data:=map[string]string{"msg":"获取插入id错误"}
		t,_ :=template.ParseFiles("admin/addArticle.html")
		t.Execute(w,data)		
	}
	id:=strconv.Itoa(int(article_id))
	data:=map[string]string{"msg":"插入成功"+id}
	t,_ :=template.ParseFiles("admin/addArticle.html")
	t.Execute(w,data)
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
* 退出
*/
func logout(w http.ResponseWriter,r *http.Request){
	delSession(w,r,manageIdKey)
	io.WriteString(w,"<script type='text/javascript'>location.href='/login'</script>")
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
* delete Session
*/
func delSession(w http.ResponseWriter,r *http.Request,key string) (interface{}){
	sess,_:= globalSessions.SessionStart(w, r)
    defer sess.SessionRelease(w)
    err := sess.Delete(key)
    return err
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
	/*Web*/
	http.HandleFunc("/",index)
	http.HandleFunc("/article",article)
	
	/*Admin*/
	http.HandleFunc("/login",login)
	http.HandleFunc("/image",getImageCode)
	http.HandleFunc("/doLogin",doLogin)
	http.HandleFunc("/admin",admin)
	http.HandleFunc("/addArticle",addArticle)
	http.HandleFunc("/doAddArticle",doAddArticle)
	http.HandleFunc("/logout",logout)
	
	/*Api*/
	http.HandleFunc("/articleListApi",articleListApi)
	
	/*static*/
	http.Handle("/public/",http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	
	err:=http.ListenAndServe(":9999",nil)
	if err!=nil{
		log.Fatal("ListenAndServe:",err)
	}
}
