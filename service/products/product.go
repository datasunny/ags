package products

import (
	"database/sql"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/featen/ags/service/auth"
	"github.com/featen/ags/service/config"
	log "github.com/featen/utils/log"
	"net/http"
	"strconv"
	"time"
)

type Comment struct {
	Id         int64
	User       int64
	Content    string
	CreateTime time.Time
}

type Product struct {
	Id            int64
	NavName       string
	Status        int64 //0: not for sale, 1: for sale, 2: on sale
	EnName        string
	CnName        string
	CoverPhoto    string
	Introduction  string
	Spec          string
	Price         float64
	Discount      float64
	Photos        []string
	Comments      []Comment
	UserLiked     []int64
	UserFavorited []int64
	CountShared   []int64 //0:fb 1:twitter 2:weibo 3:wechat
	SaleURL       []string
}

type SearchCount struct {
	Total     int64
	PageLimit int
}

const productPageLimit = 10

func Register() {
	log.Info("product registered")
	ws := new(restful.WebService)
	ws.Path("/product").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)
	ws.Route(ws.GET("").To(getAllProducts).Filter(auth.AuthEmployeeFilter))
	ws.Route(ws.GET("/{navname}").To(findProductByNavName))
	ws.Route(ws.POST("").To(addProduct).Filter(auth.AuthEmployeeFilter))
	ws.Route(ws.PUT("").To(saveProduct).Filter(auth.AuthEmployeeFilter))

	ws.Route(ws.GET("/search/{searchtext}/page/{pagenumber}").To(searchProducts).Filter(auth.AuthEmployeeFilter))
	ws.Route(ws.GET("/search/{searchtext}/count").To(searchProductsCount).Filter(auth.AuthEmployeeFilter))
	restful.Add(ws)

	wsDeal := new(restful.WebService)
	wsDeal.Path("/deals").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)
	wsDeal.Route(wsDeal.GET("").To(getAllProducts))
	wsDeal.Route(wsDeal.GET("/{navname}").To(findProductByNavName))
	wsDeal.Route(wsDeal.GET("/page/{pageNumber}").To(getPageDeals))
	restful.Add(wsDeal)
}

func FindProductNames(id string) (string, string) {
	log.Debug("find prodcut names for id %s", id)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	querySql := "SELECT en_name, nav_name FROM product WHERE id=?"
	var eName, navName sql.NullString

	err = dbHandler.QueryRow(querySql, id).Scan(&eName, &navName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("No product found for %s", id)
			return "", ""
		} else {
			log.Error("DB query failed: %v", err)
			return "", ""
		}
	}
	return eName.String, navName.String
}

func FindProductName(id string) string {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	querySql := "SELECT cn_name FROM product WHERE id=? "
	var cnName sql.NullString

	err = dbHandler.QueryRow(querySql, id).Scan(&cnName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("No product found for %s", id)
			return ""
		} else {
			log.Error("DB query failed: %v", err)
			return ""
		}
	}
	return cnName.String

}

func searchProducts(req *restful.Request, resp *restful.Response) {
	t := req.PathParameter("searchtext")
	p, _ := strconv.Atoi(req.PathParameter("pagenumber"))

	customers, ret := dbSearchProducts(t, p)
	if ret == http.StatusOK {
		resp.WriteEntity(customers)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func searchProductsCount(req *restful.Request, resp *restful.Response) {
	t := req.PathParameter("searchtext")
	n, ret := dbSearchProductsCount(t)
	if ret == http.StatusOK {
		resp.WriteEntity(SearchCount{n, productPageLimit})
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func getAllProducts(req *restful.Request, resp *restful.Response) {
	log.Debug("get all products")
	allProducts, ret := dbGetAllProducts()
	if ret == http.StatusOK {
		log.Debug("write all products info")
		resp.WriteEntity(allProducts)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func getPageDeals(req *restful.Request, resp *restful.Response) {
	pagenumber, err := strconv.ParseInt(req.PathParameter("pageNumber"), 10, 64)
	var ret = http.StatusBadRequest
	if err != nil {
		resp.WriteErrorString(ret, http.StatusText(ret))
		return
	}
	pageDeals, ret := dbGetPageDeals(pagenumber)
	if ret == http.StatusOK {
		resp.WriteEntity(pageDeals)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func findProductByNavName(req *restful.Request, resp *restful.Response) {
	log.Debug("try to find product with nav name : %s", req.PathParameter("navname"))
	product := new(Product)
	product.NavName = req.PathParameter("navname")
	ret := dbFindProductByNavName(product)
	if ret == http.StatusOK {
		resp.WriteEntity(product)
	} else {
		resp.WriteErrorString(ret, http.StatusText(ret))
	}
}

func dbFindProductByNavName(p *Product) int {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	navname := p.NavName
	log.Debug("get product detail for %s", navname)
	querySql := "SELECT id, nav_name, status, en_name, cn_name, cover_photo, introduction, spec, price, discount FROM product WHERE nav_name=? "
	var productId, status sql.NullInt64
	var navName, enName, cnName, coverPhoto, introduction, spec sql.NullString
	var price, discount sql.NullFloat64
	err = dbHandler.QueryRow(querySql, navname).Scan(&productId, &navName, &status, &enName, &cnName, &coverPhoto, &introduction, &spec, &price, &discount)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("No product found for %s", navname)
			return http.StatusNotFound
		} else {
			log.Error("DB query failed: %v", err)
			return http.StatusInternalServerError
		}
	}
	p.Id = productId.Int64
	p.NavName = navName.String
	p.Status = status.Int64
	p.EnName = enName.String
	p.CnName = cnName.String
	p.CoverPhoto = coverPhoto.String
	p.Introduction = introduction.String
	p.Spec = spec.String
	p.Price = price.Float64
	p.Discount = discount.Float64

	queryPhotoSql := "SELECT url FROM product_photo WHERE product_id=? ORDER BY id DESC LIMIT 100"
	rows, err := dbHandler.Query(queryPhotoSql, p.Id)
	defer rows.Close()

	ps := make([]string, 0, 100)
	for rows.Next() {
		var url sql.NullString
		rows.Scan(&url)
		if len(url.String) > 0 {
			ps = append(ps, url.String)
		}
	}
	p.Photos = ps

	querySaleUrlSql := "SELECT url FROM product_saleurl WHERE product_id=? ORDER BY id DESC LIMIT 100"
	surows, err := dbHandler.Query(querySaleUrlSql, p.Id)
	defer surows.Close()

	su := make([]string, 0, 100)
	for surows.Next() {
		var url sql.NullString
		surows.Scan(&url)
		if len(url.String) > 0 {
			su = append(su, url.String)
		}
	}
	p.SaleURL = su

	return http.StatusOK
}

func dbGetPageDeals(pagenumber int64) ([]Product, int) {
	log.Debug("get page deals for %d", pagenumber)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	pageDealsLimit, err := strconv.ParseInt(config.GetValue("DealsPerPage"), 10, 64)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	if pagenumber <= 0 {
		return nil, http.StatusBadRequest
	}

	//show latest deals only
	//offset := (pagenumber - 1) * pageDealsLimit
	offset := 0
	stmt, err := dbHandler.Prepare("SELECT id, nav_name, status, en_name, cn_name, cover_photo, price, discount FROM product WHERE status!=0 ORDER BY id desc limit ? offset ?")
	if err != nil {
		log.Error("Prepare to get page deal failed : %v", err)
		return nil, http.StatusInternalServerError
	}
	defer stmt.Close()
	rows, err := stmt.Query(pageDealsLimit, offset)
	if err != nil {
		log.Fatal("Query page deals failed: %v ", err)
		return nil, http.StatusNotFound
	}
	defer rows.Close()

	products := make([]Product, 0)
	for rows.Next() {
		var productId, status sql.NullInt64
		var navName, enName, cnName, coverPhoto sql.NullString
		var price, discount sql.NullFloat64
		rows.Scan(&productId, &navName, &status, &enName, &cnName, &coverPhoto, &price, &discount)
		products = append(products, Product{productId.Int64, navName.String, status.Int64, enName.String, cnName.String, coverPhoto.String, "", "", price.Float64, discount.Float64, nil, nil, nil, nil, nil, nil})
	}
	return products, http.StatusOK
}

func dbSearchProductsCount(t string) (int64, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	querySql := fmt.Sprintf("select count(id) from product where en_name like '%%%s%%' ", t)
	var n sql.NullInt64
	err = dbHandler.QueryRow(querySql).Scan(&n)
	if err != nil {
		log.Debug("sql : %s", querySql)
		log.Error("DB query failed: %v", err)
		return 0, http.StatusInternalServerError
	}
	return n.Int64, http.StatusOK
}

func dbSearchProducts(t string, p int) ([]Product, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	offset := productPageLimit * (p - 1)
	querySql := fmt.Sprintf("select id, nav_name, status, en_name, cover_photo, price from product WHERE en_name like '%%%s%%' order by id limit %d offset %d", t, productPageLimit, offset)

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

	products := make([]Product, 0)
	for rows.Next() {
		var productId, status sql.NullInt64
		var navName, enName, coverPhoto sql.NullString
		var price sql.NullFloat64
		rows.Scan(&productId, &navName, &status, &enName, &coverPhoto, &price)
		products = append(products, Product{productId.Int64, navName.String, status.Int64, enName.String, "", coverPhoto.String, "", "", price.Float64, 0, nil, nil, nil, nil, nil, nil})
	}
	return products, http.StatusOK
}
func dbGetAllProducts() ([]Product, int) {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	var getSql = "SELECT count(*) FROM product ORDER BY id DESC "
	var product_count sql.NullInt64
	err = dbHandler.QueryRow(getSql).Scan(&product_count)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info("sql error")
			return nil, http.StatusInternalServerError
		} else {
			log.Error("DB query failed: %v", err)
			return nil, http.StatusInternalServerError
		}
	}
	if product_count.Int64 == 0 {
		log.Error("No product in db")
		return nil, http.StatusNotFound
	}

	stmt, err := dbHandler.Prepare("SELECT id, nav_name, status, en_name, cn_name, cover_photo, introduction, spec, price, discount FROM product ORDER BY id desc ")
	if err != nil {
		log.Error("Prepare all product failed : %v", err)
		return nil, http.StatusInternalServerError
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("Query all product failed, something changed on db schema? : %v ", err)
		return nil, http.StatusNotFound
	}
	defer rows.Close()

	products := make([]Product, 0, product_count.Int64)
	for rows.Next() {
		var productId, status sql.NullInt64
		var navName, enName, cnName, coverPhoto, introduction, spec sql.NullString
		var price, discount sql.NullFloat64
		rows.Scan(&productId, &navName, &status, &enName, &cnName, &coverPhoto, &introduction, &spec, &price, &discount)
		products = append(products, Product{productId.Int64, navName.String, status.Int64, enName.String, cnName.String, coverPhoto.String, introduction.String, spec.String, price.Float64, discount.Float64, nil, nil, nil, nil, nil, nil})
	}
	return products, http.StatusOK
}

func saveProduct(req *restful.Request, resp *restful.Response) {
	product := new(Product)
	err := req.ReadEntity(&product)
	if err == nil {
		ret := dbSaveProduct(product)
		if ret == http.StatusOK {
			resp.WriteHeader(http.StatusOK)
			resp.WriteEntity(product)
		} else {
			resp.WriteErrorString(ret, http.StatusText(ret))
		}
	} else {
		resp.WriteError(http.StatusInternalServerError, err)
	}
}

func dbSaveProduct(p *Product) int {
	log.Debug("try to save product %v", p)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	if p.Photos != nil && len(p.Photos) > 0 {
		p.CoverPhoto = p.Photos[0]
	}

	tx, err := dbHandler.Begin()
	pSql := "UPDATE product SET status=?, en_name=?, cn_name=?, cover_photo=?, introduction=?, spec=?, price=?, discount=? WHERE nav_name=?"
	_, err = dbHandler.Exec(pSql, p.Status, p.EnName, p.CnName, p.CoverPhoto, p.Introduction, p.Spec, p.Price, p.Discount, p.NavName)
	if err != nil {
		tx.Rollback()
		log.Error("SQL: %s, err: %v", pSql, err)
		return http.StatusInternalServerError
	}

	dpSql := "DELETE FROM product_photo where product_id=?"
	_, err = dbHandler.Exec(dpSql, p.Id)
	if err != nil {
		tx.Rollback()
		log.Error("SQL: %s, err: %v", dpSql, err)
		return http.StatusInternalServerError
	}

	stmt, err := dbHandler.Prepare("INSERT INTO product_photo (product_id, url) VALUES (?,?)")
	if err != nil {
		tx.Rollback()
		log.Error("prepare failed : %v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	for _, url := range p.Photos {
		_, err = stmt.Exec(p.Id, url)
		if err != nil {
			tx.Rollback()
			log.Error("insert product_photo failed: %v", err)
			return http.StatusInternalServerError
		}
	}

	dpSql = "DELETE FROM product_saleurl where product_id=?"
	_, err = dbHandler.Exec(dpSql, p.Id)
	if err != nil {
		tx.Rollback()
		log.Error("SQL: %s, err: %v", dpSql, err)
		return http.StatusInternalServerError
	}

	stmt2, err := dbHandler.Prepare("INSERT INTO product_saleurl (product_id, url) VALUES (?,?)")
	if err != nil {
		tx.Rollback()
		log.Error("prepare product_saleurl failed : %v", err)
		return http.StatusInternalServerError
	}
	defer stmt2.Close()

	for _, url := range p.SaleURL {
		_, err = stmt2.Exec(p.Id, url)
		if err != nil {
			tx.Rollback()
			log.Error("insert product_photo failed: %v", err)
			return http.StatusInternalServerError
		}
	}
	tx.Commit()

	return http.StatusOK
}

func addProduct(req *restful.Request, resp *restful.Response) {
	product := new(Product)
	err := req.ReadEntity(&product)
	if err == nil {
		ret := dbAddProduct(product)
		if ret == http.StatusOK {
			resp.WriteHeader(http.StatusCreated)
			resp.WriteEntity(product)
		} else {
			resp.WriteErrorString(ret, http.StatusText(ret))
		}
	} else {
		resp.WriteError(http.StatusInternalServerError, err)
	}
}

func dbAddProduct(p *Product) int {
	log.Debug("try to add new product %v", p)
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
	}
	defer dbHandler.Close()

	if p.Photos != nil && len(p.Photos) > 0 {
		p.CoverPhoto = p.Photos[0]
	}

	tx, err := dbHandler.Begin()
	insertSql := "INSERT INTO product (nav_name, status, en_name, cn_name, cover_photo, introduction, spec, price, discount) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	r, err := dbHandler.Exec(insertSql, p.NavName, p.Status, p.EnName, p.CnName, p.CoverPhoto, p.Introduction, p.Spec, p.Price, p.Discount)
	if err != nil {
		tx.Rollback()
		log.Error("SQL: %s, err: %v", insertSql, err)
		return http.StatusInternalServerError
	}
	id, _ := r.LastInsertId()

	stmt, err := dbHandler.Prepare("INSERT INTO product_photo (product_id, url) VALUES (?,?)")
	if err != nil {
		tx.Rollback()
		log.Error("prepare failed : %v", err)
		return http.StatusInternalServerError
	}
	defer stmt.Close()

	for _, url := range p.Photos {
		_, err = stmt.Exec(id, url)
		if err != nil {
			tx.Rollback()
			log.Error("insert product_photo failed: %v", err)
			return http.StatusInternalServerError
		}
	}

	stmt2, err := dbHandler.Prepare("INSERT INTO product_saleurl (product_id, url) VALUES (?,?)")
	if err != nil {
		tx.Rollback()
		log.Error("prepare product_saleurl failed : %v", err)
		return http.StatusInternalServerError
	}
	defer stmt2.Close()

	for _, url := range p.SaleURL {
		_, err = stmt2.Exec(id, url)
		if err != nil {
			tx.Rollback()
			log.Error("insert product_photo failed: %v", err)
			return http.StatusInternalServerError
		}
	}
	tx.Commit()

	return http.StatusOK
}
