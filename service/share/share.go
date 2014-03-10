package share

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"github.com/featen/ags/service/auth"
	"github.com/featen/ags/service/config"
	log "github.com/featen/utils/log"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type ShareUrls struct {
	Place, UserId, Urls string
}

func CreateDbSql() []string {
	sqls := []string{
		"CREATE TABLE IF NOT EXISTS session_upload_files (id integer primary key, session varchar unique, urls varchar)",
	}

	return sqls
}

func InitDbSql() []string {
	return []string{}
}

func Register() {
	log.Info("share service registered")

	http.HandleFunc("/uploadphoto", uploadPhotoHandler)
}

func getSessionUploadFileUrls(session string) (int, string) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	var getSql = "SELECT urls FROM session_upload_files WHERE session=?"
	var value sql.NullString
	err = dbHandler.QueryRow(getSql, session).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("No value found for %s", session)
			return http.StatusNotFound, ""
		} else {
			log.Error("DB query failed: %v", err)
			return http.StatusInternalServerError, ""
		}
	}
	return http.StatusOK, value.String
}

func setSessionUploadFileUrls(session, file_urls string) int {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	var setSql = "INSERT OR REPLACE INTO session_upload_files (session, urls) VALUES (?,?) "
	_, err = dbHandler.Exec(setSql, session, file_urls)
	if err != nil {
		log.Error("SQL: %s", setSql)
		log.Error("DB exec failed, insert session: %s, urls %s, failed : %v", session, file_urls, err)
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func uploadPhotoHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s, err := auth.CookieStore.Get(r, "ags-session")
		if err != nil {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		if s.Values["id"] == nil || s.Values["time"] == nil || s.Values["magic"] == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		log.Debug("id: %s, times: %s", s.Values["id"].(string), s.Values["time"].(string))
		_, urls := getSessionUploadFileUrls(s.Values["magic"].(string))
		fmt.Fprintf(w, "%s", urls)
	case "POST":
		//parse the multipart form in the request
		err := r.ParseMultipartForm(100000)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//get a ref to the parsed multipart form
		m := r.MultipartForm

		//get the *fileheaders
		if m == nil {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		files := m.File["files"]
		urls := make([]string, 0, len(files))
		for i, _ := range files {
			//for each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			b, err := ioutil.ReadAll(file)
			h := md5.New()
			h.Write(b)
			filename := fmt.Sprintf("%x", h.Sum(nil))
			//create destination file making sure the path is writeable.
			//dst, err := os.Create("data/upload/" + files[i].Filename)
			fileurl := fmt.Sprintf("/upload/%s", filename)
			dst, err := os.Create("webapp/upload/" + filename)
			defer dst.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//copy the uploaded file to the destination file
			//if _, err := io.Copy(dst, file); err != nil {
			if _, err := dst.Write(b); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			urls = append(urls, fileurl)
		}

		s, err := auth.CookieStore.Get(r, "ags-session")
		if err != nil {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		if s.Values["id"] == nil || s.Values["time"] == nil || s.Values["magic"] == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		setSessionUploadFileUrls(s.Values["magic"].(string), strings.Join(urls, ";"))
		fmt.Fprintf(w, "%s", strings.Join(urls, ";"))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	return
}
