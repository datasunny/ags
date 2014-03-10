package service

import (
	"database/sql"
	"fmt"
	"github.com/featen/ags/service/articles"
	"github.com/featen/ags/service/auth"
	"github.com/featen/ags/service/config"
	"github.com/featen/ags/service/enquires"
	"github.com/featen/ags/service/products"
	"github.com/featen/ags/service/reports"
	"github.com/featen/ags/service/share"
	"github.com/featen/ags/service/users"
	log "github.com/featen/utils/log"
	_ "github.com/mattn/go-sqlite3"
)

func createDb() {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
		fmt.Println("dbHandler failed", err)
	}
	defer dbHandler.Close()

	sqls := []string{
		/*
		   user -> type : 0: admin, 1: customer, 2: customer without email, 3: employee
		   user -> phone: phone1, phone2, mobile1...
		   user_tag is for employee access control
		*/
		"CREATE TABLE IF NOT EXISTS visitor (id integer NOT NULL PRIMARY KEY, create_time timestamp default current_timestamp, session_time timestamp)",
		"CREATE TABLE IF NOT EXISTS user (id integer NOT NULL PRIMARY KEY, name text, username text, type integer, pass text, cover_photo text, desc text, phone text, phone_2 text, email text unique,  create_time timestamp default current_timestamp, session_time timestamp)",
		"CREATE TABLE IF NOT EXISTS user_address (id integer NOT NULL  PRIMARY KEY, user_id integer, address text, city text, province text, country text, postal text, receiver text, phone text, is_default integer)",
		"CREATE TABLE IF NOT EXISTS user_payment (id integer NOT NULL  PRIMARY KEY, user_id integer, payment_type integer, name_on_card text, card_number text, security_code text)",
		"CREATE TABLE IF NOT EXISTS user_log (id integer NOT NULL PRIMARY KEY, user_id integer, operation_type text, operation_detail text, operation_time timestamp default current_timestamp)",
		"CREATE TABLE IF NOT EXISTS user_recover_pass (id integer PRIMARY KEY, email text unique, temp_password text, magic text)",
		"CREATE TABLE IF NOT EXISTS product (id integer PRIMARY KEY, nav_name varchar unique, status integer, en_name varchar, cn_name varchar, cover_photo varchar, introduction varchar, spec varchar, price real, discount real)",
		"CREATE TABLE IF NOT EXISTS product_photo (id integer PRIMARY KEY, product_id integer, url string, unique(product_id, url) on conflict replace)",
		"CREATE TABLE IF NOT EXISTS product_saleurl (id integer PRIMARY KEY, product_id integer, url string, unique(product_id, url) on conflict replace)",
		"CREATE TABLE IF NOT EXISTS enquires (id integer NOT NULL PRIMARY KEY, status integer, customer_id integer, customer_name text, subject text, message text, employee_id integer, followup text, shipping_address_id integer default 0, create_time timestamp default current_timestamp, last_modify_time timestamp)",
		"CREATE TABLE IF NOT EXISTS enquire_product (id integer NOT NULL  PRIMARY KEY, enquire_id integer, user_id integer, product_id integer, product_navname text, product_name text, cover_photo text, price real)",
		"CREATE TABLE IF NOT EXISTS reviewboard (id integer NOT NULL PRIMARY KEY, customer_type integer,  customer_id integer, status integer, product_id integer, product_navname text, product_name text, cover_photo text, price real)",
		"CREATE TABLE IF NOT EXISTS article (id integer NOT NULL PRIMARY KEY, title text, navname text unique, cover_photo text, intro text, content text, create_by_user_id integer, create_time timestamp default current_timestamp, last_modify_time timestamp)",
	}

	for _, s := range sqls {
		_, err := dbHandler.Exec(s)
		if err != nil {
			log.Fatal("%q: %s\n", err, s)
		}
	}

}

func updateAdminUser() {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
		fmt.Println("dbHandler failed", err)
	}
	defer dbHandler.Close()
	updateAdminSql := fmt.Sprintf("UPDATE user set name=?, pass=?, email=? WHERE type=0")
	dbHandler.Exec(updateAdminSql, config.GetValue("AdminName"), config.GetValue("AdminPassword"), config.GetValue("AdminEmail"))
}

func addAdminUser() {
	dbHandler, err := sql.Open("sqlite3", config.GetValue("DbFile"))
	if err != nil {
		log.Fatal("%v", err)
		fmt.Println("dbHandler failed", err)
	}
	defer dbHandler.Close()
	addAdminSql := fmt.Sprintf("INSERT INTO user (name, type, pass, email) VALUES ('%s', 0, '%s', '%s')", config.GetValue("AdminName"), config.GetValue("AdminPassword"), config.GetValue("AdminEmail"))
	dbHandler.Exec(addAdminSql)
}

func RegService() {
	config.InitConfigs("data/ags.config")
	auth.SetSysMagicNumber([]byte(config.GetValue("SysMagicNumber")))
	inited := config.IsConfigInited()
	if !inited {
		createDb()
		addAdminUser()
		config.SetValue("dbInited", "Y")
	} else {
		updateAdminUser()
	}

	users.Register()
	articles.Register()
	share.Register()
	products.Register()
	enquires.Register()
	reports.Register()
}
