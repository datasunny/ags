package enquires

import (
	"database/sql"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/featen/ags/service/auth"
	"github.com/featen/ags/service/config"
	"github.com/featen/ags/service/products"
	"github.com/featen/ags/service/users"
	log "github.com/featen/utils/log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ReviewboardProduct struct {
	Id         int64
	NavName    string
	Name       string
	CoverPhoto string
	Price      float64
}

type Reviewboard struct {
	Identity string //user, visitor
	Products []ReviewboardProduct
}

type Enquire struct {
	Id           string
	Status       int64 //0:new, 1:not reachable, 2:done
	CustomerId   int64
	CustomerName string
	Subject      string
	Message      string
	EmployeeId   int64
	Followup     string
	Products     []ReviewboardProduct
	CreateTime   string
	ModifyTime   string
}

type SearchCount struct {
	Total     int64
	PageLimit int
}

const enquirePageLimit = 10

const timeLayout = "2006-01-02 3:04pm"

func Register() {
	log.Info("enquire registered")

	ws := new(restful.WebService)
	ws.Path("/enquire").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)
	ws.Route(ws.GET("/search/{searchtext}/page/{pagenumber}").To(searchEnquires).Filter(auth.AuthEmployeeFilter))
	ws.Route(ws.GET("/search/{searchtext}/count").To(searchEnquiresCount).Filter(auth.AuthEmployeeFilter))
	ws.Route(ws.GET("/count/{cond}").To(getEnquiresCountByCond).Filter(auth.AuthEmployeeFilter))
	ws.Route(ws.GET("/cond/{cond}").To(findEnquiresByCond).Filter(auth.AuthEmployeeFilter))
	ws.Route(ws.GET("/id/{EnquireId}").To(getEnquire).Filter(auth.AuthFilter))
	ws.Route(ws.POST("").To(addEnquire).Filter(auth.AuthFilter))
	ws.Route(ws.PUT("/id/{EnquireId}").To(followupEnquire).Filter(auth.AuthEmployeeFilter))
	restful.Add(ws)

	wsr := new(restful.WebService)
	wsr.Path("/reviewboard").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)
	wsr.Route(wsr.GET("").To(getReviewboardDetail))
	wsr.Route(wsr.POST("").To(addProductToReviewboard))
	restful.Add(wsr)
}

//usertype: 1:customer, 2:visitor
func dbGetReviewboardDetail(usertype int, userid int64) ([]ReviewboardProduct, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	querySql := "SELECT id, product_id, product_navname, product_name, cover_photo, price FROM reviewboard WHERE customer_type=? AND customer_id=? "

	stmt, err := dbHandler.Prepare(querySql)
	if err != nil {
		log.Debug("querySql: %s", querySql)
		log.Error("Prepare failed: %v", err)
		return nil, http.StatusInternalServerError
	}
	defer stmt.Close()
	rows, err := stmt.Query(usertype, userid)
	if err != nil {
		log.Error("Query reviewboard detail failed, error : %v", err)
		return nil, http.StatusNotFound
	}
	defer rows.Close()

	ps := make([]ReviewboardProduct, 0)
	for rows.Next() {
		var id, product_id sql.NullInt64
		var product_name, product_navname, cover_photo sql.NullString
		var price sql.NullFloat64
		rows.Scan(&id, &product_id, &product_navname, &product_name, &cover_photo, &price)
		ps = append(ps, ReviewboardProduct{product_id.Int64, product_navname.String, product_name.String, cover_photo.String, price.Float64})
	}
	return ps, http.StatusOK
}

func getVisitorReviewboardDetail(req *restful.Request, resp *restful.Response) ([]ReviewboardProduct, int) {
	b, visitorid := auth.AuthVisitorHandler(req.Request, resp.ResponseWriter)
	if !b {
		log.Debug("Add a new visitor cookie")
		auth.AddVisitorCookie(req.Request, resp.ResponseWriter)
		return nil, http.StatusNotFound
	} else {
		log.Debug("Get detail for an old visitor, %d", visitorid)
		id, err := strconv.ParseInt(visitorid, 10, 64)
		if err != nil {
			return nil, http.StatusNotFound
		}
		return dbGetReviewboardDetail(2, id)
	}
}

func getReviewboardDetail(req *restful.Request, resp *restful.Response) {
	log.Debug("Try to get reviewboard detail")
	b, userid := auth.AuthHandler(req.Request, resp.ResponseWriter)
	var ret int
	var ps []ReviewboardProduct
	c := new(Reviewboard)
	if !b {
		log.Debug("This is a visitor")
		ps, ret = getVisitorReviewboardDetail(req, resp)
		c.Identity = "Visitor"
		c.Products = ps
	} else {
		id, err := strconv.ParseInt(userid, 10, 64)
		if err != nil {
			ret = http.StatusInternalServerError
		} else {
			ps, ret = dbGetReviewboardDetail(1, id)
			c.Identity = "Customer"
			c.Products = ps
		}
	}

	if ret == http.StatusOK {
		resp.WriteEntity(c)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func dbAddProductToReviewboard(usertype int, userid int64, p *ReviewboardProduct) int {

	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	insertSql := "INSERT INTO reviewboard (customer_type, customer_id, product_id, product_navname, product_name, cover_photo, price) VALUES (?, ?, ?, ?, ?,  ?, ?) "
	_, err = dbHandler.Exec(insertSql, usertype, userid, p.Id, p.NavName, p.Name, p.CoverPhoto, p.Price)
	if err != nil {
		log.Error("Sql: %s", insertSql)
		log.Error("DB Insert product to reviewboard failed, %v", err)
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func addProductToReviewboard(req *restful.Request, resp *restful.Response) {
	log.Debug("Try to add product to reviewboard")
	p := new(ReviewboardProduct)
	err := req.ReadEntity(&p)
	if err != nil {
		resp.WriteErrorString(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	var id int64
	var ret int
	b, userid := auth.AuthHandler(req.Request, resp.ResponseWriter)
	if !b {
		log.Debug("This is a visitor")
		b, visitorid := auth.AuthVisitorHandler(req.Request, resp.ResponseWriter)
		if !b {
			id = auth.AddVisitorCookie(req.Request, resp.ResponseWriter)
		} else {
			id, _ = strconv.ParseInt(visitorid, 10, 64)
		}
		ret = dbAddProductToReviewboard(2, id, p)
	} else {
		id, err := strconv.ParseInt(userid, 10, 64)
		if err != nil {
			ret = http.StatusInternalServerError
		} else {
			ret = dbAddProductToReviewboard(1, id, p)
		}
	}

	if ret == http.StatusOK {
		resp.WriteHeader(http.StatusOK)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func getUserEnquires(req *restful.Request, resp *restful.Response) {
	userId, err := strconv.ParseInt(req.Attribute("agsuserid").(string), 10, 64)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
	}
	cond := fmt.Sprintf("customer_id=%d", userId)
	es, ret := dbFindEnquiresByCond(cond)

	if ret == http.StatusOK {
		resp.WriteEntity(es)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func searchEnquires(req *restful.Request, resp *restful.Response) {
	t := req.PathParameter("searchtext")
	p, _ := strconv.Atoi(req.PathParameter("pagenumber"))

	customers, ret := dbSearchEnquires(t, p)
	if ret == http.StatusOK {
		resp.WriteEntity(customers)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func searchEnquiresCount(req *restful.Request, resp *restful.Response) {
	t := req.PathParameter("searchtext")
	n, ret := dbSearchEnquiresCount(t)
	if ret == http.StatusOK {
		resp.WriteEntity(SearchCount{n, enquirePageLimit})
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func getEnquiresCountByCond(req *restful.Request, resp *restful.Response) {
	log.Debug("try to get enquires count with cond : %s", req.PathParameter("cond"))

	cond := req.PathParameter("cond")
	escount, ret := dbGetEnquiresCountByCond(cond)
	if ret == http.StatusOK {
		resp.WriteEntity(escount)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}

}

func findEnquiresByCond(req *restful.Request, resp *restful.Response) {
	log.Debug("try to find enquires with cond : %s", req.PathParameter("cond"))

	cond := req.PathParameter("cond")
	es, ret := dbFindEnquiresByCond(cond)
	if ret == http.StatusOK {
		resp.WriteEntity(es)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func getEnquire(req *restful.Request, resp *restful.Response) {
	log.Debug("try to get enquire with id : %s", req.PathParameter("EnquireId"))
	userId, err := strconv.ParseInt(req.Attribute("agsuserid").(string), 10, 64)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
	}
	id := req.PathParameter("EnquireId")
	e := new(Enquire)
	e.Id = auth.Decode(id)
	ret := dbGetEnquire(e, userId)
	if ret == http.StatusOK {
		e.Id = id
		resp.WriteEntity(e)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func dbGetEnquire(e *Enquire, userId int64) int {
	log.Debug("get enquire detail for %s", e.Id)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	querySql := "SELECT id, status, customer_id, customer_name,  subject, message, employee_id, followup,  create_time, last_modify_time FROM enquires WHERE id=?"
	var enquire_id, status, customer_id, employee_id sql.NullInt64
	var customer_name, subject, message, followup sql.NullString
	var create_time, last_modify_time time.Time
	err = dbHandler.QueryRow(querySql, e.Id).Scan(&enquire_id, &status, &customer_id, &customer_name, &subject, &message, &employee_id, &followup, &create_time, &last_modify_time)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("No enquire found for %s", e.Id)
			return http.StatusNotFound
		} else {
			log.Debug("sql : %s", querySql)
			log.Error("DB query failed: %v", err)
			return http.StatusInternalServerError
		}
	}
	if userId != customer_id.Int64 {
		u := users.DbFindUser(strconv.FormatInt(userId, 10))
		if u == nil || (u.Type == 1 || u.Type == 2) {
			return http.StatusForbidden
		}
	}

	if !status.Valid {
		return http.StatusNotFound
	} else {
		e.Products = dbGetEnquireProducts(e.Id)
		e.Status = status.Int64
		e.CustomerId = customer_id.Int64
		e.CustomerName = customer_name.String
		e.Subject = subject.String
		e.Message = message.String
		e.EmployeeId = employee_id.Int64
		e.Followup = followup.String
		e.CreateTime = create_time.Format(timeLayout)
		e.ModifyTime = last_modify_time.Format(timeLayout)

		return http.StatusOK
	}
}

func dbGetEnquireProducts(id string) []ReviewboardProduct {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	querySql := "SELECT id, product_id, product_navname, product_name, cover_photo, price FROM enquire_product WHERE enquire_id=?"

	stmt, err := dbHandler.Prepare(querySql)
	if err != nil {
		log.Debug("querySql: %s", querySql)
		log.Error("Prepare failed: %v", err)
		return nil
	}
	defer stmt.Close()
	rows, err := stmt.Query(id)
	if err != nil {
		log.Error("Query enquire products failed, error : %v", err)
		return nil
	}
	defer rows.Close()

	ps := make([]ReviewboardProduct, 0)
	for rows.Next() {
		var id, product_id sql.NullInt64
		var product_name, product_navname, cover_photo sql.NullString
		var price sql.NullFloat64
		rows.Scan(&id, &product_id, &product_navname, &product_name, &cover_photo, &price)
		ps = append(ps, ReviewboardProduct{product_id.Int64, product_navname.String, product_name.String, cover_photo.String, price.Float64})
	}
	return ps
}

func dbSearchEnquiresCount(t string) (int64, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	querySql := fmt.Sprintf("select count(id) from enquires where subject like '%%%s%%' or customer_name like '%%%s%%' ", t, t)
	var n sql.NullInt64
	err = dbHandler.QueryRow(querySql).Scan(&n)
	if err != nil {
		log.Debug("sql : %s", querySql)
		log.Error("DB query failed: %v", err)
		return 0, http.StatusInternalServerError
	}
	return n.Int64, http.StatusOK
}

func dbSearchEnquires(t string, p int) ([]Enquire, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	offset := enquirePageLimit * (p - 1)
	querySql := fmt.Sprintf("select id, status, customer_id, customer_name, subject, message, followup, create_time, last_modify_time from enquires where subject like '%%%s%%' or customer_name like '%%%s%%' order by id desc limit %d offset %d", t, t, enquirePageLimit, offset)

	stmt, err := dbHandler.Prepare(querySql)
	if err != nil {
		log.Debug("querySql: %s", querySql)
		log.Error("Prepare failed : %v", err)
		return nil, http.StatusInternalServerError
	}

	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("Query products failed, something changed on db schema? : %v ", err)
		return nil, http.StatusNotFound
	}
	defer rows.Close()

	es := make([]Enquire, 0)
	for rows.Next() {
		var enquire_id, status, customer_id sql.NullInt64
		var customer_name, subject, message, followup sql.NullString
		var create_time, last_modify_time time.Time
		rows.Scan(&enquire_id, &status, &customer_id, &customer_name, &subject, &message, &followup, &create_time, &last_modify_time)
		es = append(es, Enquire{auth.Encode(strconv.FormatInt(enquire_id.Int64, 10)), status.Int64, customer_id.Int64, customer_name.String, subject.String, message.String, 1, followup.String, nil, create_time.Format(timeLayout), last_modify_time.Format(timeLayout)})
	}
	return es, http.StatusOK
}

func buildSqlCond(cond string) (string, int, int) {
	var offset = 0
	var limit = 100
	conds := strings.Split(cond, "&")
	if len(conds) == 0 || len(conds) > 10 {
		return "id>0", limit, offset
	}

	var sqlString = make([]string, 0, 10)
	for _, c := range conds {
		v := strings.Split(c, "=")
		if len(v) != 2 {
			return "id>0", limit, offset
		}
		switch v[0] {
		case "id":
			sqlString = append(sqlString, "id="+v[1])
		case "status":
			sqlString = append(sqlString, "status="+v[1])
		case "employee_id":
			sqlString = append(sqlString, "employee_id="+v[1])
		case "customer_id":
			sqlString = append(sqlString, "customer_id="+v[1])
		case "customer_name":
			sqlString = append(sqlString, fmt.Sprintf("customer_name like '%%%s%%'", v[1]))
		case "offset":
			o, err := strconv.Atoi(v[1])
			if err != nil {
				offset = 0
			} else {
				offset = o
			}
		case "limit":
			l, err := strconv.Atoi(v[1])
			if err != nil {
				limit = 100
			} else {
				limit = l
			}

		}
	}

	if len(sqlString) == 0 {
		return "id>0", limit, offset
	}

	return strings.Join(sqlString, " AND "), limit, offset
}

func dbGetEnquiresCountByCond(cond string) (int64, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	sqlCond, _, _ := buildSqlCond(cond)

	log.Debug("get enquires for %s", cond)
	querySql := "SELECT count(id) FROM enquires WHERE ? "
	var escount sql.NullInt64
	err = dbHandler.QueryRow(querySql, sqlCond).Scan(&escount)
	if err != nil {
		log.Debug("sql : %s", querySql)
		log.Error("DB query failed: %v", err)
		return 0, http.StatusInternalServerError
	}
	return escount.Int64, http.StatusOK
}

func dbFindEnquiresByCond(cond string) ([]Enquire, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	sqlCond, limit, offset := buildSqlCond(cond)

	log.Debug("get enquires for %s, %d, %d", sqlCond, limit, offset)
	querySql := fmt.Sprintf("SELECT id, status, customer_id, customer_name, subject, message, employee_id, followup, create_time, last_modify_time FROM enquires WHERE %s ORDER BY id DESC LIMIT %d OFFSET %d ", sqlCond, limit, offset)

	stmt, err := dbHandler.Prepare(querySql)
	if err != nil {
		log.Debug("querySql: %s", querySql)
		log.Error("Prepare failed : %v", err)
		return nil, http.StatusInternalServerError
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("Query enquire failed, something changed on db schema? : %v ", err)
		return nil, http.StatusNotFound
	}
	defer rows.Close()

	es := make([]Enquire, 0)
	for rows.Next() {
		var enquire_id, status, customer_id, employee_id sql.NullInt64
		var customer_name, subject, message, followup sql.NullString
		var create_time, last_modify_time time.Time
		rows.Scan(&enquire_id, &status, &customer_id, &customer_name, &subject, &message, &employee_id, &followup, &create_time, &last_modify_time)
		es = append(es, Enquire{auth.Encode(strconv.FormatInt(enquire_id.Int64, 10)), status.Int64, customer_id.Int64, customer_name.String, subject.String, message.String, employee_id.Int64, followup.String, nil, create_time.Format(timeLayout), last_modify_time.Format(timeLayout)})
	}
	return es, http.StatusOK
}

func addEnquire(req *restful.Request, resp *restful.Response) {
	e := new(Enquire)
	err := req.ReadEntity(&e)
	if err == nil {
		e.CustomerId, err = strconv.ParseInt(req.Attribute("agsuserid").(string), 10, 64)
		ret := dbAddEnquire(e)
		if ret == http.StatusOK {
			resp.WriteHeader(http.StatusCreated)
			e.Id = auth.Encode(e.Id)
			resp.WriteEntity(e)
		} else {
			resp.WriteErrorString(ret, http.StatusText(ret))
		}
	} else {
		resp.WriteError(http.StatusInternalServerError, err)
	}
}

func dbAddEnquire(e *Enquire) int {
	log.Debug("try to add enquire %v", e)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	tx, err := dbHandler.Begin()

	currTime := time.Now()
	stmt, err := dbHandler.Prepare("INSERT INTO enquires(status, customer_id, customer_name, subject, message, employee_id, followup, create_time, last_modify_time) VALUES (0, ?, ?, ?, ?, ?, ?, ?,?)")
	if err != nil {
		log.Debug("stmt is %v", stmt)
		log.Error("err: %v", err)
		tx.Rollback()
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	var name = users.FindUserName(strconv.FormatInt(e.CustomerId, 10))
	r, err := stmt.Exec(e.CustomerId, name, e.Subject, e.Message, 0, "", currTime, currTime)
	if err != nil {
		log.Error("%v", err)
		tx.Rollback()
		return http.StatusBadRequest
	}
	id, _ := r.LastInsertId()
	e.Id = strconv.FormatInt(id, 10)

	pStmt, err := dbHandler.Prepare("INSERT INTO enquire_product (enquire_id, user_id, product_id, product_navname, product_name, cover_photo, price) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		tx.Rollback()
		log.Error("Prepare insert enquire_product failed: %v", err)
		return http.StatusInternalServerError
	}
	defer pStmt.Close()

	for _, op := range e.Products {
		ename, navname := products.FindProductNames(strconv.FormatInt(op.Id, 10))

		_, err = pStmt.Exec(id, e.CustomerId, op.Id, navname, ename, op.CoverPhoto, op.Price)
		if err != nil {
			tx.Rollback()
			log.Error("INSERT enquire product failed: %v", err)
			return http.StatusInternalServerError
		}

	}

	clearCartSql := "DELETE FROM reviewboard WHERE customer_id=? "
	_, err = dbHandler.Exec(clearCartSql, e.CustomerId)
	if err != nil {
		log.Error("Sql: %s", clearCartSql)
		tx.Rollback()
		return http.StatusInternalServerError
	}

	tx.Commit()
	return http.StatusOK

}

func dbFollowupEnquire(e *Enquire) int {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	updateSql := "UPDATE enquires SET employee_id=?, status=?, followup=? where id=?"
	_, err = dbHandler.Exec(updateSql, e.EmployeeId, e.Status, e.Followup, e.Id)
	if err != nil {
		log.Error("Sql: %s", updateSql)
		log.Error("DB followup enquire failed, %v", err)
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func followupEnquire(req *restful.Request, resp *restful.Response) {
	id := req.PathParameter("EnquireId")

	e := new(Enquire)
	err := req.ReadEntity(&e)
	if err == nil {
		e.EmployeeId, err = strconv.ParseInt(req.Attribute("agsemployeeid").(string), 10, 64)
		e.Id = auth.Decode(id)
		ret := dbFollowupEnquire(e)
		if ret == http.StatusOK {
			resp.WriteHeader(http.StatusCreated)
			e.Id = auth.Encode(e.Id)
			resp.WriteEntity(e)
		} else {
			resp.WriteErrorString(ret, http.StatusText(ret))
		}
	} else {
		resp.WriteError(http.StatusInternalServerError, err)
	}
}
