package users

import (
	"database/sql"
	log "featen/ags/modules/log"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/featen/ags/service/auth"
	"github.com/featen/ags/service/config"
	"github.com/featen/ags/service/mails"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Id, Name, Pass, Email string
	Type                  int64
	Phone, CoverPhoto     string
}

type UserAddress struct {
	Id        string
	Name      string
	Address   string
	City      string
	Province  string
	Postal    string
	Country   string
	Phone     string
	IsDefault int64
}

type VerifyUser struct {
	Email, Pass string
}

type SignoutUser struct {
	Id string
}

type CustomerLog struct {
	CustomerId      string
	OperationType   string
	OperationDetail string
	OperationTime   string
}

type Customer struct {
	Id         string
	Name       string
	CoverPhoto string
	Desc       string
	Phone      string
	Email      string
	Logs       []CustomerLog
}

type SearchCount struct {
	Total     int64
	PageLimit int
}

const (
	OperSignUp = iota
	OperSignIn
	OperCreateArticle
	OperUpdateArticle
	OperDeleteArticle
)

const timeLayout = "2006-01-02 3:04pm"
const customerPageLimit = 10

func Register() {
	ws := new(restful.WebService)
	ws.
		Path("/users").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

	ws.Route(ws.GET("/").To(currUser))
	ws.Route(ws.GET("/id/{user-id}").To(findUser))
	ws.Route(ws.GET("/shippings").To(getShippingOptions).Filter(auth.AuthFilter))

	ws.Route(ws.POST("/").To(createUser))
	ws.Route(ws.POST("/signin").To(signinUser))
	ws.Route(ws.POST("/signout").To(signoutUser))
	ws.Route(ws.POST("/address").To(setShipping).Filter(auth.AuthFilter))
	ws.Route(ws.PUT("/").To(updateUser))
	ws.Route(ws.PUT("/password").To(updateUserPassword))
	ws.Route(ws.DELETE("/{user-id}").To(removeUser))

	restful.Add(ws)

	ws_recover := new(restful.WebService)
	ws_recover.
		Path("/recover").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)
	ws_recover.Route(ws_recover.GET("/{recover_magic}").To(verifyRecover))
	ws_recover.Route(ws_recover.POST("/").To(sendRecover))
	restful.Add(ws_recover)

	ws_customer := new(restful.WebService)
	ws_customer.Path("/customers").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)
	ws_customer.Route(ws_customer.GET("/search/{searchtext}/page/{pagenumber}").To(searchCustomers).Filter(auth.AuthEmployeeFilter))
	ws_customer.Route(ws_customer.GET("/search/{searchtext}/count").To(searchCustomersCount).Filter(auth.AuthEmployeeFilter))
	ws_customer.Route(ws_customer.GET("/{cond}").To(findCustomersByCond).Filter(auth.AuthEmployeeFilter))
	ws_customer.Route(ws_customer.GET("/id/{id}").To(findCustomer).Filter(auth.AuthEmployeeFilter))
	ws_customer.Route(ws_customer.POST("").To(addCustomer).Filter(auth.AuthEmployeeFilter))
	ws_customer.Route(ws_customer.POST("/id").To(saveCustomer).Filter(auth.AuthEmployeeFilter))
	ws_customer.Route(ws_customer.POST("/log").To(addCustomerLog).Filter(auth.AuthEmployeeFilter))

	restful.Add(ws_customer)

	log.Info("user registered! ")
}

func currUser(req *restful.Request, resp *restful.Response) {
	usr := auth.CurrCookieUser(req, resp)
	if usr == nil {
		resp.WriteErrorString(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	} else {
		resp.WriteEntity(usr)
	}
}

func findUser(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	usr := new(User)
	usr.Id = id
	ret := dbFindUser(usr)
	if ret == http.StatusOK {
		response.WriteEntity(User{id, usr.Name, "", usr.Email, usr.Type, usr.Phone, usr.CoverPhoto})
	} else {
		response.WriteErrorString(ret, http.StatusText(ret))
	}
}

func getShippingOptions(req *restful.Request, resp *restful.Response) {
	userId, err := strconv.ParseInt(req.Attribute("agsuserid").(string), 10, 64)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
	}
	options, ret := dbGetShippingOptions(userId)
	if ret == http.StatusOK {
		resp.WriteEntity(options)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func verifyRecover(req *restful.Request, resp *restful.Response) {
	magic := req.PathParameter("recover_magic")
	log.Debug("magic is %s", magic)
	ret, id := dbVerifyRecover(magic)
	if ret == http.StatusOK {
		auth.AddCookie(req.Request, resp.ResponseWriter, strconv.FormatInt(id, 10))
		http.Redirect(resp.ResponseWriter, req.Request, "/#!/mypassword", http.StatusFound)

	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func signoutUser(req *restful.Request, resp *restful.Response) {
	var id SignoutUser
	err := req.ReadEntity(&id)
	if err == nil {
		auth.DelCookie(req, resp, id.Id)
	} else {
		log.Debug("sign out id %s failed", id.Id)
		resp.WriteErrorString(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}
}

func sendRecover(req *restful.Request, resp *restful.Response) {
	var ru VerifyUser
	err := req.ReadEntity(&ru)
	if err != nil {
		log.Debug("read recover user info %s failed", ru.Email)
		resp.WriteErrorString(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	} else {
		exist := dbEmailExist(ru.Email)
		if !exist {
			log.Debug("not a valid email %s", ru.Email)
			resp.WriteErrorString(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		} else {
			magic := auth.GenMagic(ru.Email, time.Now().String())
			dbInsertRecoverInfo(ru.Email, magic)
			go mails.SendRecoverMail(ru.Email, magic)
			resp.WriteHeader(http.StatusOK)
		}
	}
}

func setShipping(req *restful.Request, resp *restful.Response) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	userId, err := strconv.ParseInt(req.Attribute("agsuserid").(string), 10, 64)
	ua := new(UserAddress)
	err = req.ReadEntity(&ua)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	var setSql string
	var r sql.Result

	if len(ua.Id) != 0 {
		//update
		setSql = "UPDATE user_address SET receiver=?, address=?, city=?, province=?, postal=?, country=?, phone=?, is_default=? WHERE id=?"
		r, err = dbHandler.Exec(setSql, ua.Name, ua.Address, ua.City, ua.Province, ua.Postal, ua.Country, ua.Phone, ua.IsDefault, ua.Id)
	} else {
		//insert
		setSql = "INSERT INTO user_address (user_id, receiver, address, city, province, postal, country, phone, is_default) VALUES (?, ?,?, ?,?,?,?,?,?) "
		r, err = dbHandler.Exec(setSql, userId, ua.Name, ua.Address, ua.City, ua.Province, ua.Postal, ua.Country, ua.Phone, ua.IsDefault)
	}
	if err != nil {
		log.Error("SQL: %s", setSql)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	if len(ua.Id) == 0 {
		id, _ := r.LastInsertId()
		ua.Id = strconv.FormatInt(id, 10)
	}

	resp.WriteEntity(ua)
}

func updateUserPassword(request *restful.Request, response *restful.Response) {
	s, err := auth.CookieStore.Get(request.Request, "ags-session")
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	authed := auth.Check(s.Values["id"].(string), s.Values["time"].(string), s.Values["magic"].(string))
	if !authed {
		response.WriteErrorString(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	usr := new(User)
	err = request.ReadEntity(&usr)

	if err == nil {
		usr.Id = s.Values["id"].(string)

		var ret int
		if len(usr.Pass) == 0 {
			ret = http.StatusBadRequest
		} else {
			ret = dbUpdatePass(usr)
		}
		if ret == http.StatusOK {
			response.WriteEntity(usr)
		} else {
			response.WriteErrorString(ret, http.StatusText(ret))
		}
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}
func updateUser(request *restful.Request, response *restful.Response) {
	s, err := auth.CookieStore.Get(request.Request, "ags-session")
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	authed := auth.Check(s.Values["id"].(string), s.Values["time"].(string), s.Values["magic"].(string))
	if !authed {
		response.WriteErrorString(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	usr := new(User)
	err = request.ReadEntity(&usr)

	if err == nil {
		usr.Id = s.Values["id"].(string)

		ret := dbUpdateUser(usr)
		if ret == http.StatusOK {
			response.WriteEntity(usr)
		} else {
			response.WriteErrorString(ret, http.StatusText(ret))
		}
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func signinUser(request *restful.Request, response *restful.Response) {
	usr := new(User)
	err := request.ReadEntity(&usr)
	if err == nil {
		if ret := dbCheckUser(usr); ret == http.StatusOK {
			auth.AddCookie(request.Request, response.ResponseWriter, usr.Id)
			response.WriteEntity(usr)
		} else {
			response.WriteErrorString(ret, http.StatusText(ret))
		}
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func createUser(request *restful.Request, response *restful.Response) {
	usr := new(User)
	err := request.ReadEntity(&usr)
	if err == nil {
		dbCreateUser(usr)
		response.WriteHeader(http.StatusCreated)
		//AddCookie(request.Request, response.ResponseWriter, usr.Id)
		//response.WriteEntity(usr)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

// DELETE http://localhost:8080/users/1
//
func removeUser(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	dbDeleteUser(id)
}

func searchCustomers(req *restful.Request, resp *restful.Response) {
	t := req.PathParameter("searchtext")
	p, _ := strconv.Atoi(req.PathParameter("pagenumber"))

	customers, ret := dbSearchCustomers(t, p)
	if ret == http.StatusOK {
		resp.WriteEntity(customers)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func searchCustomersCount(req *restful.Request, resp *restful.Response) {
	t := req.PathParameter("searchtext")
	n, ret := dbSearchCustomersCount(t)
	if ret == http.StatusOK {
		resp.WriteEntity(SearchCount{n, customerPageLimit})
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func findCustomersByCond(req *restful.Request, resp *restful.Response) {
	log.Debug("try to find customers with cond : %s", req.PathParameter("cond"))
	cond := req.PathParameter("cond")
	customers, ret := dbFindCustomersByCond(cond)
	if ret == http.StatusOK {
		resp.WriteEntity(customers)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func addCustomer(request *restful.Request, response *restful.Response) {
	c := new(Customer)
	err := request.ReadEntity(&c)
	if err == nil {
		dbCreateCustomer(c)
		response.WriteHeader(http.StatusCreated)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func saveCustomer(request *restful.Request, response *restful.Response) {
	c := new(Customer)
	err := request.ReadEntity(&c)
	if err == nil {
		dbSaveCustomer(c)
		response.WriteHeader(http.StatusCreated)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func addCustomerLog(request *restful.Request, response *restful.Response) {
	c := new(CustomerLog)
	err := request.ReadEntity(&c)
	if err == nil {
		dbCreateCustomerLog(c)
		response.WriteHeader(http.StatusCreated)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func findCustomer(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	c := new(Customer)
	c.Id = id
	ret := dbFindCustomer(c)
	if ret == http.StatusOK {
		response.WriteEntity(c)
	} else {
		response.WriteErrorString(ret, http.StatusText(ret))
	}
}

func DbFindUser(id string) *User {
	var u = &User{Id: id}
	ret := dbFindUser(u)
	if ret != http.StatusOK {
		return nil
	} else {
		return u
	}
}

func FindUserName(id string) string {
	u := DbFindUser(id)
	if u != nil {
		return u.Name
	} else {
		return ""
	}
}

func GetUserAddr(userid int64) *UserAddress {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	querySql := "SELECT id, receiver, address, city, province, postal, country, phone from user_address where is_default=1 and user_id=? "
	var uaid sql.NullInt64
	var receiver, address, city, province, postal, country, phone sql.NullString
	err = dbHandler.QueryRow(querySql, userid).Scan(&uaid, &receiver, &address, &city, &province, &postal, &country, &phone)
	if err != nil {
		log.Error("Sql: %s", querySql)
		log.Error("DB choose shipping failed, %v", err)
		return nil
	}

	return &UserAddress{strconv.FormatInt(uaid.Int64, 10), receiver.String, address.String, city.String, province.String, postal.String, country.String, phone.String, 1}
}

func dbGetShippingOptions(userid int64) ([]UserAddress, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	var getSql = "SELECT id, receiver, address, city, province, postal, country, phone, is_default FROM user_address WHERE user_id=?"
	rows, err := dbHandler.Query(getSql, userid)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	defer rows.Close()

	uas := make([]UserAddress, 0, 2)
	for rows.Next() {
		var id, is_default sql.NullInt64
		var name, address, city, province, postal, country, phone sql.NullString
		rows.Scan(&id, &name, &address, &city, &province, &postal, &country, &phone, &is_default)
		uas = append(uas, UserAddress{strconv.FormatInt(id.Int64, 10), name.String, address.String, city.String, province.String, postal.String, country.String, phone.String, is_default.Int64})
	}
	return uas, http.StatusOK

}

func dbFindUser(user *User) int {
	log.Debug("try to find user with id : %v", user.Id)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	if len(user.Id) == 0 {
		return http.StatusNotFound
	}

	stmt, err := dbHandler.Prepare("SELECT type, name, email, phone, cover_photo FROM user WHERE id=? LIMIT 1")
	if err != nil {
		log.Error("%v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	var name, email, phone, cover_photo sql.NullString
	var user_type sql.NullInt64
	err = stmt.QueryRow(user.Id).Scan(&user_type, &name, &email, &phone, &cover_photo)
	if err != nil {
		log.Error("%v", err)
		if err == sql.ErrNoRows {
			return http.StatusNotFound
		} else {
			return http.StatusInternalServerError
		}
	}

	if !name.Valid {
		return http.StatusNotFound
	} else {
		user.Type = user_type.Int64
		user.Email = email.String
		user.Name = name.String
		user.Phone = phone.String
		user.CoverPhoto = cover_photo.String
		return http.StatusOK
	}
}

func dbEmailExist(email string) bool {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	if len(email) == 0 {
		return false
	}
	var getSql = "SELECT id FROM user WHERE email=?"
	var value sql.NullInt64
	err = dbHandler.QueryRow(getSql, email).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("No email found for %s", email)
			return false
		} else {
			log.Error("DB query failed: %v", err)
			return false
		}
	}
	return true
}

func dbUpdatePass(user *User) int {
	log.Debug("try to update user %v password", user)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	stmt, err := dbHandler.Prepare("UPDATE user SET pass=? WHERE id=? ")
	if err != nil {
		log.Error("%v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Pass, user.Id)
	if err != nil {
		log.Error("%v", err)
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func dbUpdateUser(user *User) int {
	log.Debug("try to update user %v", user)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	stmt, err := dbHandler.Prepare("UPDATE user SET name=?, phone=?, cover_photo=? WHERE id=? ")
	if err != nil {
		log.Error("%v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Name, user.Phone, user.CoverPhoto, user.Id)
	if err != nil {
		log.Error("%v", err)
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func dbUpdateUserType(user *User, user_type int) int {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	var setSql = "UPDATE user set type=? WHERE id=? "
	_, err = dbHandler.Exec(setSql, user_type, user.Id)
	if err != nil {
		log.Error("SQL: %s", setSql)
		log.Error("DB exec failed : %v", err)
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func dbCheckUser(user *User) int {
	log.Debug("try to find user with id : %v | %v", user.Email, user.Pass)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	stmt, err := dbHandler.Prepare("SELECT id, type, name, email, phone, cover_photo FROM user WHERE email=? AND pass=? LIMIT 1")
	if err != nil {
		log.Error("%v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	var id, user_type sql.NullInt64
	var name, email, phone, cover_photo sql.NullString
	err = stmt.QueryRow(user.Email, user.Pass).Scan(&id, &user_type, &name, &email, &phone, &cover_photo)
	if err != nil {
		log.Error("%v", err)
		if err == sql.ErrNoRows {
			return http.StatusNotFound
		} else {
			return http.StatusInternalServerError
		}
	}

	if !id.Valid {
		return http.StatusNotFound
	} else {
		user.Id = strconv.FormatInt(id.Int64, 10)
		user.Type = user_type.Int64
		user.Name = name.String
		user.Email = email.String
		user.Phone = phone.String
		user.CoverPhoto = cover_photo.String
		return http.StatusOK
	}
}

func dbCreateUser(user *User) int {
	log.Debug("try to create user %v", user)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	stmt, err := dbHandler.Prepare("INSERT INTO user (type, name, email, pass) VALUES (1, ?,?,?)")
	if err != nil {
		log.Error("%v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Name, user.Email, user.Pass)
	if err != nil {
		log.Error("%v", err)
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func dbCreateCustomer(c *Customer) int {
	log.Debug("try to create user %v", c)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	var userType = 1
	if len(c.Email) == 0 {
		userType = 2
	}
	stmt, err := dbHandler.Prepare("INSERT INTO user (type, name, email, cover_photo, phone, desc) VALUES (?,?,?,?,?,?)")
	if err != nil {
		log.Error("%v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	_, err = stmt.Exec(userType, c.Name, c.Email, c.CoverPhoto, c.Phone, c.Desc)
	if err != nil {
		log.Error("%v", err)
		return http.StatusBadRequest
	}
	return http.StatusOK
}
func dbCreateCustomerLog(c *CustomerLog) int {
	log.Debug("try to create customer log %v", c)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	stmt, err := dbHandler.Prepare("INSERT INTO user_log (user_id, operation_type, operation_detail) VALUES (?,?,?)")
	if err != nil {
		log.Error("%v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	_, err = stmt.Exec(c.CustomerId, c.OperationType, c.OperationDetail)
	if err != nil {
		log.Error("%v", err)
		return http.StatusBadRequest
	}
	return http.StatusOK
}
func dbSaveCustomer(c *Customer) int {
	log.Debug("try to save user %v", c)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	var userType = 1
	if len(c.Email) == 0 {
		userType = 2
	}
	stmt, err := dbHandler.Prepare("UPDATE user SET type=?, name=?, email=?, cover_photo=?, phone=?, desc=? WHERE id=?")
	if err != nil {
		log.Error("%v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	_, err = stmt.Exec(userType, c.Name, c.Email, c.CoverPhoto, c.Phone, c.Desc, c.Id)
	if err != nil {
		log.Error("%v", err)
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func dbDeleteUser(id string) int {
	log.Debug("try to delete user id %v", id)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	stmt, err := dbHandler.Prepare("DELETE FROM user WHERE id=?")
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

func dbInsertRecoverInfo(email, magic string) int {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	var setSql = "INSERT OR REPLACE INTO user_recover_pass (email, magic) VALUES (?,?) "
	_, err = dbHandler.Exec(setSql, email, magic)
	if err != nil {
		log.Error("SQL: %s", setSql)
		log.Error("DB exec failed, insert recover info, email: %s, magic %s, failed : %v", email, magic, err)
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func dbVerifyRecover(magic string) (int, int64) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	var getSql = "SELECT u.id FROM user u, user_recover_pass urp WHERE u.email=urp.email AND urp.magic=?"
	var value sql.NullInt64
	err = dbHandler.QueryRow(getSql, magic).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("No magic found for %s", magic)
			return http.StatusNotFound, 0
		} else {
			log.Error("DB query failed: %v", err)
			return http.StatusInternalServerError, 0
		}
	}

	var delSql = "DELETE from user_recover_pass WHERE magic=?"
	_, err = dbHandler.Exec(delSql, magic)
	if err != nil {
		log.Error("SQL: %s", delSql)
		return http.StatusInternalServerError, 0
	}
	return http.StatusOK, value.Int64
}

func buildSqlCond(cond string) (string, int, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

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
		case "name":
			sqlString = append(sqlString, fmt.Sprintf("name like '%%%s%%'", v[1]))
		case "desc":
			sqlString = append(sqlString, fmt.Sprintf("desc like '%%%s%%'", v[1]))
		case "phone":
			sqlString = append(sqlString, fmt.Sprintf("phone like '%%%s%%'", v[1]))
		case "email":
			sqlString = append(sqlString, fmt.Sprintf("email like '%%%s%%'", v[1]))
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

func dbSearchCustomersCount(t string) (int64, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	querySql := fmt.Sprintf("select count(id) from user where type=1 and (name like '%%%s%%' or phone like '%%%s%%' or email like '%%%s%%') ", t, t, t)
	var n sql.NullInt64
	err = dbHandler.QueryRow(querySql).Scan(&n)
	if err != nil {
		log.Debug("sql : %s", querySql)
		log.Error("DB query failed: %v", err)
		return 0, http.StatusInternalServerError
	}
	return n.Int64, http.StatusOK
}

func dbSearchCustomers(t string, p int) ([]Customer, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	offset := customerPageLimit * (p - 1)
	querySql := fmt.Sprintf("select id, name, cover_photo, phone, desc, email from user where type=1 and (name like '%%%s%%' or phone like '%%%s%%' or email like '%%%s%%') order by id limit %d offset %d", t, t, t, customerPageLimit, offset)

	stmt, err := dbHandler.Prepare(querySql)
	if err != nil {
		log.Debug("querySql: %s", querySql)
		log.Error("Prepare failed : %v", err)
		return nil, http.StatusInternalServerError
	}

	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("Query customers failed, something changed on db schema? : %v ", err)
		return nil, http.StatusNotFound
	}
	defer rows.Close()

	customers := make([]Customer, 0)
	for rows.Next() {
		var customerId sql.NullInt64
		var name, coverPhoto, phone, desc, email sql.NullString
		rows.Scan(&customerId, &name, &coverPhoto, &phone, &desc, &email)
		customers = append(customers, Customer{strconv.FormatInt(customerId.Int64, 10), name.String, coverPhoto.String, desc.String, phone.String, email.String, nil})
	}
	return customers, http.StatusOK

}

func dbFindCustomersByCond(cond string) ([]Customer, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	sqlCond, limit, offset := buildSqlCond(cond)
	log.Debug("get customers for %s", cond)
	querySql := fmt.Sprintf("SELECT id, name, cover_photo, desc, phone, email FROM user WHERE type=1 AND %s LIMIT %d OFFSET %d ", sqlCond, limit, offset)

	stmt, err := dbHandler.Prepare(querySql)
	if err != nil {
		log.Debug("querySql: %s", querySql)
		log.Error("Prepare failed : %v", err)
		return nil, http.StatusInternalServerError
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("Query customers failed, something changed on db schema? : %v ", err)
		return nil, http.StatusNotFound
	}
	defer rows.Close()

	customers := make([]Customer, 0, limit)
	for rows.Next() {
		var customerId sql.NullInt64
		var name, coverPhoto, desc, phone, email sql.NullString
		rows.Scan(&customerId, &name, &coverPhoto, &desc, &phone, &email)
		customers = append(customers, Customer{strconv.FormatInt(customerId.Int64, 10), name.String, coverPhoto.String, "", phone.String, email.String, nil})
	}
	return customers, http.StatusOK
}

func dbFindCustomer(c *Customer) int {
	log.Debug("get customer detail for %d", c.Id)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	querySql := "SELECT id, name, cover_photo, desc, phone, email FROM user WHERE type in (1,2) AND id=? "
	var id sql.NullInt64
	var name, coverPhoto, desc, phone, email sql.NullString
	err = dbHandler.QueryRow(querySql, c.Id).Scan(&id, &name, &coverPhoto, &desc, &phone, &email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("No customer found for %d", c.Id)
			return http.StatusNotFound
		} else {
			log.Debug("sql : %s", querySql)
			log.Error("DB query failed: %v", err)
			return http.StatusInternalServerError
		}
	}

	c.Name = name.String
	c.CoverPhoto = coverPhoto.String
	c.Desc = desc.String
	c.Phone = phone.String
	c.Email = email.String

	queryLogSql := "SELECT operation_type, operation_detail, operation_time FROM user_log WHERE user_id=? ORDER BY id DESC LIMIT 100"
	rows, err := dbHandler.Query(queryLogSql, c.Id)
	defer rows.Close()

	logs := make([]CustomerLog, 0, 100)
	for rows.Next() {
		var operation_type, operation_detail sql.NullString
		var operation_time time.Time
		rows.Scan(&operation_type, &operation_detail, &operation_time)
		logs = append(logs, CustomerLog{c.Id, operation_type.String, operation_detail.String, operation_time.Format(timeLayout)})
	}
	c.Logs = logs

	return http.StatusOK
}
