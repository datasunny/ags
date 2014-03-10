package reports

import (
	"database/sql"
	"github.com/emicklei/go-restful"
	"github.com/featen/ags/service/auth"
	"github.com/featen/ags/service/config"
	log "github.com/featen/utils/log"
	"net/http"
	"strings"
)

type ReportData struct {
	Timeframe string
	Type      string
	Xvalues   []string
	Yvalues   []float64
}

const timeLayout = "2006-01-02 3:04pm"

func Register() {
	log.Info("report registered")
	ws := new(restful.WebService)
	ws.Path("/report").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)
	ws.Route(ws.GET("/{cond}").To(genDataByCond).Filter(auth.AuthEmployeeFilter))
	restful.Add(ws)
}

func genDataByCond(req *restful.Request, resp *restful.Response) {
	log.Debug("try to gen report with cond : %s", req.PathParameter("cond"))
	cond := req.PathParameter("cond")
	reportData, ret := dbGenDataByCond(cond)
	if ret == http.StatusOK {
		resp.WriteEntity(reportData)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func dbGenDataNewCustomer(reportData *ReportData) (*ReportData, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	var querySql string
	switch reportData.Timeframe {
	case "0":
		querySql = "SELECT count(id), date(create_time) from user WHERE create_time >= (SELECT date(julianday(date('now'))-7)) AND create_time <= (SELECT date(julianday(date('now')))) group by date(create_time)"
	case "1":
		querySql = "SELECT count(id), date(create_time) from user WHERE create_time >= (SELECT date(julianday(date('now'))-30)) AND create_time <= (SELECT date(julianday(date('now')))) group by date(create_time)"
	case "2":
		querySql = "SELECT count(id), date(create_time) from user WHERE create_time >= (SELECT date(julianday(date('now'))-180)) AND create_time <= (SELECT date(julianday(date('now')))) group by date(create_time)"
	case "3":
		querySql = "SELECT count(id), date(create_time) from user group by date(create_time)"

	}

	stmt, err := dbHandler.Prepare(querySql)
	if err != nil {
		log.Debug("querySql: %s", querySql)
		log.Error("Prepare failed : %v", err)
		return nil, http.StatusInternalServerError
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("Query trans failed, something changed on db schema? : %v ", err)
		return nil, http.StatusNotFound
	}
	defer rows.Close()

	for rows.Next() {
		var id_count sql.NullFloat64
		var date_day sql.NullString
		rows.Scan(&id_count, &date_day)
		reportData.Xvalues = append(reportData.Xvalues, date_day.String)
		reportData.Yvalues = append(reportData.Yvalues, id_count.Float64)
	}

	return reportData, http.StatusOK

}
func dbGenDataPageVisits(reportData *ReportData) (*ReportData, int) {
	return nil, http.StatusNotImplemented
}

func dbGenDataSaleAmount(reportData *ReportData) (*ReportData, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	var querySql string
	switch reportData.Timeframe {
	case "0":
		querySql = "SELECT sum(paid_amount), date(create_time) from orders WHERE status>0 AND  create_time >= (SELECT date(julianday(date('now'))-7)) AND create_time <= (SELECT date(julianday(date('now')))) group by date(create_time)"
	case "1":
		querySql = "SELECT sum(paid_amount), date(create_time) from orders WHERE status>0 AND create_time >= (SELECT date(julianday(date('now'))-30)) AND create_time <= (SELECT date(julianday(date('now')))) group by date(create_time)"
	case "2":
		querySql = "SELECT sum(paid_amount), date(create_time) from orders WHERE status>0 AND create_time >= (SELECT date(julianday(date('now'))-180)) AND create_time <= (SELECT date(julianday(date('now')))) group by date(create_time)"
	case "3":
		querySql = "SELECT sum(paid_amount), date(create_time) from orders WHERE status>0 group by date(create_time)"

	}

	stmt, err := dbHandler.Prepare(querySql)
	if err != nil {
		log.Debug("querySql: %s", querySql)
		log.Error("Prepare failed : %v", err)
		return nil, http.StatusInternalServerError
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("Query orders failed, something changed on db schema? : %v ", err)
		return nil, http.StatusNotFound
	}
	defer rows.Close()

	for rows.Next() {
		//var id_count sql.NullInt64
		var saleamount sql.NullFloat64
		var date_day sql.NullString
		rows.Scan(&saleamount, &date_day)
		reportData.Xvalues = append(reportData.Xvalues, date_day.String)
		reportData.Yvalues = append(reportData.Yvalues, saleamount.Float64)
	}

	return reportData, http.StatusOK
}

func dbGenDataByCond(cond string) (*ReportData, int) {
	log.Debug("gen report data for %s", cond)

	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	reportData := new(ReportData)
	conds := strings.Split(cond, "&")
	if len(conds) == 0 {
		return nil, http.StatusBadRequest
	}
	for _, c := range conds {
		v := strings.Split(c, "=")
		if len(v) != 2 {
			return nil, http.StatusBadRequest
		}
		switch v[0] {
		case "timeframe":
			reportData.Timeframe = v[1]
		case "type":
			reportData.Type = v[1]
		}
	}
	switch reportData.Type {
	case "NewCustomers":
		return dbGenDataNewCustomer(reportData)
	case "SaleAmount":
		return dbGenDataSaleAmount(reportData)
	case "PageVisits":
		return dbGenDataPageVisits(reportData)
	}
	return nil, http.StatusInternalServerError
}
