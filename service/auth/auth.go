package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	log "featen/ags/modules/log"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/featen/ags/service/config"
	"github.com/gorilla/sessions"
	"io"
	"net/http"
	"strconv"
	"time"
)

type User struct {
	Id, Name, Pass, Email string
	Type                  int64
	Phone, CoverPhoto     string
}

var sysMagicNumber []byte
var CookieStore *sessions.CookieStore

//var CookieStore = sessions.NewCookieStore(sysMagicNumber)

func SetSysMagicNumber(m []byte) {
	sysMagicNumber = m
	CookieStore = sessions.NewCookieStore(sysMagicNumber)
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

func AuthFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	b, userid := AuthHandler(req.Request, resp.ResponseWriter)
	if !b {
		log.Debug("unauthorized request %s %s", req.Request.Method, req.Request.URL)
		resp.WriteErrorString(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}
	req.SetAttribute("agsuserid", userid)
	chain.ProcessFilter(req, resp)
}

func AuthEmployeeFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	b, userid := authEmployeeHandler(req.Request, resp.ResponseWriter)
	if !b {
		log.Debug("unauthorized request %s %s", req.Request.Method, req.Request.URL)
		resp.WriteErrorString(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}
	req.SetAttribute("agsemployeeid", userid)
	chain.ProcessFilter(req, resp)
}

func saveSesstionTime(id string, t string) int {
	log.Debug("try to update user %s session time", id)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	stmt, err := dbHandler.Prepare("UPDATE user SET session_time=? WHERE id=? ")
	if err != nil {
		log.Error("prepare update user session time failed: %v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	_, err = stmt.Exec(t, id)
	if err != nil {
		log.Error("execute update user session time failed: %v", err)
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func saveVisitorSesstionTime(id string, t string) int {
	log.Debug("try to update user %s session time", id)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	stmt, err := dbHandler.Prepare("UPDATE visitor SET session_time=? WHERE id=? ")
	if err != nil {
		log.Error("prepare update visitor session time failed: %v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	_, err = stmt.Exec(t, id)
	if err != nil {
		log.Error("execute update visitor session time failed: %v", err)
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func CurrCookieUser(req *restful.Request, resp *restful.Response) *User {
	s, err := CookieStore.Get(req.Request, "ags-session")
	if err != nil {
		return nil
	}
	if s.Values["id"] == nil || s.Values["time"] == nil || s.Values["magic"] == nil {
		return nil
	}

	b := Check(s.Values["id"].(string), s.Values["time"].(string), s.Values["magic"].(string))
	if b == true {
		id := s.Values["id"].(string)
		u := DbFindUser(id)
		if u != nil {
			return u
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func updateReviewboardOwner(req *http.Request, id string) {
	log.Debug("Update Reviewboard Owner")
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	b, vid := AuthVisitorHandler(req, nil)
	if !b {
		log.Debug("Auth Visitor failed")
		return
	}
	var updateSql = "UPDATE reviewboard set customer_type=1, customer_id=? WHERE customer_type=2 AND customer_id=?"
	log.Error("Sql: %s", updateSql)
	_, err = dbHandler.Exec(updateSql, id, vid)
	if err != nil {
		log.Error("Sql: %s", updateSql)
		log.Error("Update reviewboard owner failed: %v", err)
	}
}

func AddCookie(req *http.Request, resp http.ResponseWriter, id string) {
	s, _ := CookieStore.Get(req, "ags-session")
	s.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 0,
	}
	t := time.Now().String()
	s.Values["id"] = id
	s.Values["time"] = t
	s.Values["magic"] = GenMagic(id, t)
	s.Save(req, resp)
	saveSesstionTime(id, t)
	updateReviewboardOwner(req, id)
}

func DelCookie(req *restful.Request, resp *restful.Response, id string) {
	s, _ := CookieStore.Get(req.Request, "ags-session")

	s.Values["id"] = ""
	s.Values["time"] = ""
	s.Values["magic"] = ""
	s.Save(req.Request, resp.ResponseWriter)
	saveSesstionTime(id, "")
}

func dbCreateVisitor(session_time string) int64 {

	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	var setSql = "INSERT INTO visitor (session_time) VALUES (?) "
	r, err := dbHandler.Exec(setSql, session_time)
	if err != nil {
		log.Error("SQL: %s", setSql)
		log.Error("DB exec failed, insert visiotr: %s, failed : %v", session_time, err)
		return 0
	}

	id, _ := r.LastInsertId()
	return id
}

func AddVisitorCookie(req *http.Request, resp http.ResponseWriter) int64 {
	//insert a visitor
	t := time.Now().String()
	i := dbCreateVisitor(t)
	if i == 0 {
		return 0
	}
	id := strconv.FormatInt(i, 10)

	s, _ := CookieStore.Get(req, "v")
	s.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 0,
	}
	s.Values["id"] = id
	s.Values["time"] = t
	s.Values["magic"] = GenMagic(id, t)
	s.Save(req, resp)
	//saveVisitorSesstionTime(id, t)
	return i
}

func DelVisitorCookie(req *restful.Request, resp *restful.Response, id string) {
	s, _ := CookieStore.Get(req.Request, "v")

	s.Values["id"] = ""
	s.Values["time"] = ""
	s.Values["magic"] = ""
	s.Save(req.Request, resp.ResponseWriter)
	saveVisitorSesstionTime(id, "")
}

func AuthVisitorHandler(r *http.Request, w http.ResponseWriter) (bool, string) {
	s, err := CookieStore.Get(r, "v")
	if err != nil {
		log.Debug("Cannot get session: %v", err)
		return false, ""
	}

	if s.Values["id"] == nil || s.Values["time"] == nil || s.Values["magic"] == nil {
		return false, ""
	}

	b := Check(s.Values["id"].(string), s.Values["time"].(string), s.Values["magic"].(string))
	if b == true {
		return true, s.Values["id"].(string)
	} else {
		return false, ""
	}
}
func AuthHandler(r *http.Request, w http.ResponseWriter) (bool, string) {
	s, err := CookieStore.Get(r, "ags-session")
	if err != nil {
		log.Debug("Cannot get session: %v", err)
		return false, ""
	}

	if s.Values["id"] == nil || s.Values["time"] == nil || s.Values["magic"] == nil {
		return false, ""
	}

	b := Check(s.Values["id"].(string), s.Values["time"].(string), s.Values["magic"].(string))
	if b == true {
		return true, s.Values["id"].(string)
	} else {
		return false, ""
	}
}

func authEmployeeHandler(r *http.Request, w http.ResponseWriter) (bool, string) {
	s, err := CookieStore.Get(r, "ags-session")
	if err != nil {
		log.Debug("Cannot get session: %v", err)
		return false, ""
	}

	if s.Values["id"] == nil || s.Values["time"] == nil || s.Values["magic"] == nil {
		return false, ""
	}

	b := Check(s.Values["id"].(string), s.Values["time"].(string), s.Values["magic"].(string))
	if b == true {
		u := DbFindUser(s.Values["id"].(string))
		if u != nil && (u.Type == 3 || u.Type == 0) {
			return true, s.Values["id"].(string)
		} else {
			return false, ""
		}
	} else {
		return false, ""
	}
}

func Check(id string, n string, magic string) bool {
	if m := GenMagic(id, n); m == magic {
		return true
	} else {
		return false
	}
}

func GenMagic(id string, n string) string {
	h := md5.New()
	io.WriteString(h, "ags-")
	io.WriteString(h, id)
	io.WriteString(h, "-"+n)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Encode(p string) string {
	key := []byte(sysMagicNumber)
	plaintext := []byte(p)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Debug("%v", err)
		return ""
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Debug("%v", err)
		return ""
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.

	return fmt.Sprintf("%x\n", ciphertext)
}

func Decode(c string) string {
	key := []byte(sysMagicNumber)
	ciphertext, _ := hex.DecodeString(c)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Debug("%v", err)
		return ""
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		log.Debug("ciphertext too short")
		return ""
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s\n", ciphertext)
}
