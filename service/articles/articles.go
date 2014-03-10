package articles

import (
	"database/sql"
	"github.com/emicklei/go-restful"
	"github.com/featen/ags/service/auth"
	"github.com/featen/ags/service/config"
	log "github.com/featen/utils/log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type Article struct {
	Id, Title, NavName, Intro, Content string
	UserId                             int64
	UserName, UserImg                  string
	CreateTime, ModifyTime             string
	CoverPhoto                         string
}

const timeLayout = "2006-01-02 3:04pm"

func Register() {
	log.Info("articles registered")

	ws := new(restful.WebService)
	ws.Path("/articles").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("").To(getAllArticles))
	ws.Route(ws.GET("/totalpage/number").To(getTotalPageNumber))
	ws.Route(ws.GET("/page/{pageNumber}").To(getPageArticles))

	ws.Route(ws.GET("/{article-id}").To(findArticleById).
		Doc("get an article").
		Param(ws.PathParameter("article-id", "id of the article").DataType("string")).
		Writes(Article{}))

	ws.Route(ws.GET("/name/{navname}").To(findArticleByNavName))
	ws.Route(ws.PUT("/{article-id}").To(updateArticle).Filter(auth.AuthEmployeeFilter))
	ws.Route(ws.POST("").To(createArticle).Filter(auth.AuthEmployeeFilter))

	ws.Route(ws.DELETE("/{article-id}").To(removeArticle).Filter(auth.AuthFilter))

	restful.Add(ws)
}

func getAllArticles(req *restful.Request, resp *restful.Response) {
	allArticles, ret := dbGetAllArticles()
	if ret == http.StatusOK {
		resp.WriteEntity(allArticles)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func getTotalPageNumber(req *restful.Request, resp *restful.Response) {
	pageNumber, ret := dbGetTotalPageNumber()
	if ret == http.StatusOK {
		log.Debug("pageNumber is %f", pageNumber)
		resp.WriteEntity(pageNumber)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func getPageArticles(req *restful.Request, resp *restful.Response) {
	pagenumber, err := strconv.ParseInt(req.PathParameter("pageNumber"), 10, 64)
	var ret = http.StatusBadRequest
	if err != nil {
		resp.WriteErrorString(ret, http.StatusText(ret))
		return
	}
	pageArticles, ret := dbGetPageArticles(pagenumber)
	if ret == http.StatusOK {
		resp.WriteEntity(pageArticles)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func findArticleById(req *restful.Request, resp *restful.Response) {
	article := new(Article)
	article.Id = req.PathParameter("article-id")
	ret := dbFindArticleById(article)
	if ret == http.StatusOK {
		resp.WriteEntity(article)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func findArticleByNavName(req *restful.Request, resp *restful.Response) {
	article := new(Article)
	article.NavName = req.PathParameter("navname")
	ret := dbFindArticleByNavName(article)
	if ret == http.StatusOK {
		resp.WriteEntity(article)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func updateArticle(req *restful.Request, resp *restful.Response) {
	article := new(Article)
	err := req.ReadEntity(&article)
	if err == nil {
		if ret := dbUpdateArticle(article); ret == http.StatusOK {
			resp.WriteEntity(article)
		} else {
			resp.WriteErrorString(ret, http.StatusText(ret))
		}
	} else {
		resp.WriteError(http.StatusInternalServerError, err)
	}
}

func createArticle(req *restful.Request, resp *restful.Response) {
	article := new(Article)
	err := req.ReadEntity(&article)
	if err == nil {
		article.UserId, err = strconv.ParseInt(req.Attribute("agsemployeeid").(string), 10, 64)
		ret := dbCreateArticle(article)
		if ret == http.StatusOK {
			resp.WriteHeader(http.StatusCreated)
			resp.WriteEntity(article)
		} else {
			resp.WriteErrorString(ret, http.StatusText(ret))
		}
	} else {
		resp.WriteError(http.StatusInternalServerError, err)
	}
}

func removeArticle(req *restful.Request, resp *restful.Response) {
	id := req.PathParameter("article-id")
	ret := dbDeleteArticle(id)
	if ret == http.StatusOK {
		resp.WriteHeader(http.StatusOK)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func dbGetTotalPageNumber() (float64, int) {
	log.Debug("get total page number")

	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	pageArticlesLimit, err := strconv.ParseInt(config.GetValue("DealsPerPage"), 10, 64)
	if err != nil {
		return 1, http.StatusOK
	}

	var n sql.NullFloat64
	queryLogSql := "SELECT count(*) FROM article"
	dbHandler.QueryRow(queryLogSql).Scan(&n)

	return math.Ceil(float64(n.Float64 / float64(pageArticlesLimit))), http.StatusOK
}

func dbGetAllArticles() ([]Article, int) {
	log.Debug("get all articles")
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	stmt, err := dbHandler.Prepare("SELECT a.id, a.title, a.navname,a.cover_photo, a.intro, a.content, a.create_by_user_id, u.name, a.create_time, a.last_modify_time from article a, user u WHERE a.create_by_user_id=u.id ORDER BY a.id DESC")
	if err != nil {
		log.Error("%v", err)
		return nil, http.StatusInternalServerError
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("%v", err)
		return nil, http.StatusInternalServerError
	}
	defer rows.Close()

	allarticles := make([]Article, 10)
	for rows.Next() {
		var title, navname, cover_photo, intro, content, username sql.NullString
		var articleid, userid sql.NullInt64
		var createtime, modifytime time.Time
		rows.Scan(&articleid, &title, &navname, &cover_photo, &intro, &content, &userid, &username, &createtime, &modifytime)

		allarticles = append(allarticles, Article{strconv.FormatInt(articleid.Int64, 10), title.String, navname.String, intro.String, content.String, userid.Int64, username.String, "", createtime.Format(timeLayout), modifytime.Format(timeLayout), cover_photo.String})
	}
	rows.Close()
	return allarticles, http.StatusOK
}

func dbGetPageArticles(pagenumber int64) ([]Article, int) {
	log.Debug("get page articles for %d", pagenumber)

	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	pageArticlesLimit, err := strconv.ParseInt(config.GetValue("DealsPerPage"), 10, 64)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	if pagenumber <= 0 {
		return nil, http.StatusBadRequest
	}
	offset := (pagenumber - 1) * pageArticlesLimit

	stmt, err := dbHandler.Prepare("SELECT a.id, a.title, a.navname, a.cover_photo, a.intro,  a.create_by_user_id, u.name, a.create_time, a.last_modify_time from article a, user u WHERE a.create_by_user_id=u.id ORDER BY a.id DESC  limit ? offset ?")
	if err != nil {
		log.Error("%v", err)
		return nil, http.StatusInternalServerError
	}
	defer stmt.Close()
	rows, err := stmt.Query(pageArticlesLimit, offset)
	if err != nil {
		log.Fatal("%v", err)
		return nil, http.StatusInternalServerError
	}
	defer rows.Close()

	allarticles := make([]Article, 0)
	for rows.Next() {
		var title, navname, cover_photo, intro, username sql.NullString
		var articleid, userid sql.NullInt64
		var createtime, modifytime time.Time
		rows.Scan(&articleid, &title, &navname, &cover_photo, &intro, &userid, &username, &createtime, &modifytime)

		allarticles = append(allarticles, Article{strconv.FormatInt(articleid.Int64, 10), title.String, navname.String, intro.String, "", userid.Int64, username.String, "", createtime.Format(timeLayout), modifytime.Format(timeLayout), cover_photo.String})
	}
	rows.Close()
	return allarticles, http.StatusOK
}

func dbFindArticleById(article *Article) int {
	log.Debug("try to find article with id : %v", article.Id)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()
	stmt, err := dbHandler.Prepare("SELECT a.id, a.title, a.navname, a.cover_photo, a,intro, a.content, a.create_by_user_id, u.name, a.create_time, a.last_modify_time from article a, user u WHERE a.Id = ? AND a.create_by_user_id = u.id ")
	if err != nil {
		log.Error("%v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	var title, navname, cover_photo, intro, content, username sql.NullString
	var articleid, userid sql.NullInt64
	var createtime, modifytime time.Time
	err = stmt.QueryRow(article.Id).Scan(&articleid, &title, &navname, &cover_photo, &intro, &content, &userid, &username, &createtime, &modifytime)
	if err != nil {
		log.Error("%v", err)
		if err == sql.ErrNoRows {
			return http.StatusNotFound
		} else {
			return http.StatusInternalServerError
		}
	}

	if !title.Valid {
		return http.StatusNotFound
	} else {
		article.Id = strconv.FormatInt(articleid.Int64, 10)
		article.Title = title.String
		article.NavName = navname.String
		article.CoverPhoto = cover_photo.String
		article.Content = content.String
		article.Intro = intro.String
		article.UserId = userid.Int64
		article.UserName = username.String
		article.CreateTime = createtime.Format(timeLayout)
		article.ModifyTime = modifytime.Format(timeLayout)
		return http.StatusOK
	}
}

func dbFindArticleByNavName(article *Article) int {
	log.Debug("try to find article with title : %v", article.NavName)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	stmt, err := dbHandler.Prepare("SELECT a.id, a.title, a.cover_photo, a.intro, a.content, a.create_by_user_id, u.name, a.create_time, a.last_modify_time from article a, user u WHERE a.navname = ? AND a.create_by_user_id = u.id ")
	if err != nil {
		log.Error("%v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	var title, cover_photo, intro, content, username sql.NullString
	var articleid, userid sql.NullInt64
	var createtime, modifytime time.Time
	err = stmt.QueryRow(article.NavName).Scan(&articleid, &title, &cover_photo, &intro, &content, &userid, &username, &createtime, &modifytime)
	if err != nil {
		log.Error("%v", err)
		if err == sql.ErrNoRows {
			return http.StatusNotFound
		} else {
			return http.StatusInternalServerError
		}
	}

	if !title.Valid {
		return http.StatusNotFound
	} else {
		article.Id = strconv.FormatInt(articleid.Int64, 10)
		article.CoverPhoto = cover_photo.String
		article.Title = title.String
		article.Content = content.String
		article.Intro = intro.String
		article.UserId = userid.Int64
		article.UserName = username.String
		article.CreateTime = createtime.Format(timeLayout)
		article.ModifyTime = modifytime.Format(timeLayout)
		return http.StatusOK
	}
}

func dbUpdateArticle(article *Article) int {
	log.Debug("try to update article %v", article)

	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()
	stmt, err := dbHandler.Prepare("UPDATE article SET cover_photo=?, title=?, content=?, last_modify_time=datetime('now','localtime','utc') WHERE id=?")
	if err != nil {
		log.Error("%v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	_, err = stmt.Exec(article.CoverPhoto, article.Title, article.Content, article.Id)
	if err != nil {
		log.Error("%v", err)
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func dbCreateArticle(article *Article) int {
	log.Debug("try to create article %v", article)

	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	stmt, err := dbHandler.Prepare("INSERT INTO article (title, navname, cover_photo, intro, content, create_by_user_id, last_modify_time) VALUES (?,?,?,?,?,?, datetime('now','localtime','utc'))")
	if err != nil {
		log.Error("%v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	r, err := stmt.Exec(article.Title, article.NavName, article.CoverPhoto, article.Intro, article.Content, article.UserId)
	if err != nil {
		log.Error("%v", err)
		return http.StatusBadRequest
	}
	id, _ := r.LastInsertId()
	article.Id = strconv.FormatInt(id, 10)

	return http.StatusOK
}

func dbDeleteArticle(id string) int {
	log.Debug("try to delete article id %v", id)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()
	stmt, err := dbHandler.Prepare("DELETE FROM article WHERE id=?")
	if err != nil {
		log.Error("%v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		log.Error("%v", err)
		return http.StatusBadRequest
	}
	return http.StatusOK
}
