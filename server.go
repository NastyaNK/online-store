package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	id         int
	name, pass string
}

type Cart struct {
	id, count, item, user int
	cost                  float32
	status                string
}

type Category struct {
	id                    int
	name, logo, image, bg string
}

type Item struct {
	id, category         int
	name, faculty, image string
	price                float32
}

type Result struct {
	success int
	message string
}

func main() {
	var err error
	db, err = connectToDB()
	if err != nil {
		log.Fatal("Failed connect to db:", err)
	}

	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/index", index)
	http.HandleFunc("/cart", cart)
	http.HandleFunc("/items", items)
	http.HandleFunc("/cart-items", cartItems)
	http.HandleFunc("/auth", authorize)
	http.HandleFunc("/add", add)
	http.HandleFunc("/buy", buy)

	fmt.Println("Start server: http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

var db *sql.DB

func connectToDB() (*sql.DB, error) {
	pass := "1234" //"Lumos"
	return sql.Open("mysql", "root:"+pass+"@/база")
}

func Replace(f []byte, oldNews ...string) string {
	str := string(f)
	for i := 0; i < len(oldNews); i += 2 {
		str = strings.ReplaceAll(str, "*"+oldNews[i]+"*", oldNews[i+1])
	}
	return str
}

func Show(w http.ResponseWriter, content string) {
	w.Write([]byte(content))
}

func index(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select * from categories")
	if err != nil {
		log.Fatal("Failed query:", err)
	}
	defer rows.Close()

	ctg := Category{}
	page, _ := ioutil.ReadFile("categories.html")
	category, _ := ioutil.ReadFile("cat.html")
	categories := ""
	for rows.Next() {
		rows.Scan(&ctg.id, &ctg.bg, &ctg.logo, &ctg.name, &ctg.image)
		categories += Replace(category, "id", strconv.Itoa(ctg.id), "название", ctg.name, "картинка", ctg.image)
	}

	prof, _ := ioutil.ReadFile("profile.html")
	cart, _ := ioutil.ReadFile("cart.html")
	profile := ""
	if user.id != 0 {
		profile = Replace(prof, "зарегистрирован", "true", "имя", user.name)
	} else {
		profile = Replace(prof, "зарегистрирован", "false", "имя", "")
	}
	Show(w, Replace(page, "категории", categories, "профиль", profile, "корзина", string(cart)))
}

func items(w http.ResponseWriter, r *http.Request) {
	ctg := Category{}
	itm := Item{}

	if r.Method == "GET" {
		params := r.URL.Query()
		category, ok := params["id"]
		if !ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		rows, err := db.Query("select * from categories WHERE id = " + category[0])
		if err != nil {
			log.Fatal("Failed query:", err)
		}

		if rows.Next() {
			rows.Scan(&ctg.id, &ctg.bg, &ctg.logo, &ctg.name, &ctg.image)
		}
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	rows, err := db.Query("select * from items WHERE category = " + strconv.Itoa(ctg.id))
	if err != nil {
		log.Fatal("Failed query:", err)
	}
	defer rows.Close()

	page, _ := ioutil.ReadFile("items.html")
	item, _ := ioutil.ReadFile("item.html")
	items := ""
	for rows.Next() {
		rows.Scan(&itm.id, &itm.name, &itm.price, &itm.category, &itm.faculty, &itm.image)
		items += Replace(item, "id", strconv.Itoa(itm.id), "описание", itm.name, "цена",
			fmt.Sprintf("%.2f", itm.price),
			"категория", strconv.Itoa(itm.category), "факультет", itm.faculty, "картинка", itm.image)
	}
	prof, _ := ioutil.ReadFile("profile.html")
	cart, _ := ioutil.ReadFile("cart.html")
	profile := ""
	if user.id != 0 {
		profile = Replace(prof, "зарегистрирован", "true", "имя", user.name)
	} else {
		profile = Replace(prof, "зарегистрирован", "false", "имя", "")
	}
	Show(w, Replace(page, "товары", items, "приветствие", ctg.logo, "фон", ctg.bg, "профиль", profile, "корзина", string(cart)))
}

func cartItems(w http.ResponseWriter, r *http.Request) {
	if user.id == 0 {
		return
	}

	rows, err := db.Query(fmt.Sprintf("SELECT items.*, cart.count, cart.cost FROM cart LEFT JOIN items ON cart.item = items.id WHERE cart.status = 'cart' AND cart.user = %d AND cart.count > 0", user.id))
	if err != nil {
		log.Fatal("Failed query:", err)
	}
	defer rows.Close()

	page := ""
	item, _ := ioutil.ReadFile("cart-item.html")

	cart := Cart{}
	itm := Item{}
	for rows.Next() {
		rows.Scan(&itm.id, &itm.name, &itm.price, &itm.category, &itm.faculty, &itm.image, &cart.count, &cart.cost)
		page += Replace(item, "id", strconv.Itoa(itm.id), "описание", itm.name, "цена",
			fmt.Sprintf("%.2f", itm.price), "категория", strconv.Itoa(itm.category), "факультет",
			itm.faculty, "картинка", itm.image, "стоимость", fmt.Sprintf("%.2f", cart.cost), "кол-во", fmt.Sprintf("%d", cart.count))
	}
	Show(w, page)
}

func buy(w http.ResponseWriter, r *http.Request) {
	if user.id == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	rows, err := db.Query(fmt.Sprintf("SELECT cart.cost AS cost FROM cart LEFT JOIN items ON cart.item = items.id WHERE cart.status = 'cart' AND cart.user = %d AND cart.count > 0", user.id))
	if err != nil {
		log.Fatal("Failed query:", err)
	}
	defer rows.Close()
	sum := 0
	cost := 0
	for rows.Next() {
		rows.Scan(&cost)
		sum += cost
	}
	_, err = db.Exec(fmt.Sprintf("UPDATE cart SET status = 'buyed' WHERE cart.user = %d AND cart.count > 0", user.id))
	if err != nil {
		log.Fatal("Failed query:", err)
	}
	rows, err = db.Query(fmt.Sprintf("SELECT items.*, cart.count, cart.cost FROM cart LEFT JOIN items ON cart.item = items.id WHERE cart.status = 'buyed' AND cart.user = %d", user.id))
	if err != nil {
		log.Fatal("Failed query:", err)
	}
	defer rows.Close()

	page, _ := ioutil.ReadFile("buyed.html")
	category, _ := ioutil.ReadFile("buyed-item.html")
	items := ""
	cart := Cart{}
	itm := Item{}
	for rows.Next() {
		rows.Scan(&itm.id, &itm.name, &itm.price, &itm.category, &itm.faculty, &itm.image, &cart.count, &cart.cost)
		items += Replace(category, "id", strconv.Itoa(itm.id), "описание", itm.name, "цена",
			fmt.Sprintf("%.2f", itm.price), "категория", strconv.Itoa(itm.category), "факультет",
			itm.faculty, "картинка", itm.image, "стоимость", fmt.Sprintf("%.2f", cart.cost), "кол-во", fmt.Sprintf("%d", cart.count))
	}

	prof, _ := ioutil.ReadFile("profile.html")
	profile := ""
	if user.id != 0 {
		profile = Replace(prof, "зарегистрирован", "true", "имя", user.name)
	} else {
		profile = Replace(prof, "зарегистрирован", "false", "имя", "")
	}
	Show(w, Replace(page, "товары", items, "профиль", profile, "сумма", strconv.Itoa(sum)))
}

func cart(w http.ResponseWriter, r *http.Request) {
	Show(w, "Нифига")
}

func getCart(item int) (Cart, error) {
	row := db.QueryRow(fmt.Sprintf("SELECT * FROM cart WHERE item = %d AND user = %d AND status = 'cart'", item, user.id))
	cart := Cart{}
	err := row.Scan(&cart.id, &cart.item, &cart.user, &cart.count, &cart.cost, &cart.status)
	return cart, err
}

func getItem(id int) (Item, error) {
	row := db.QueryRow(fmt.Sprintf("SELECT * FROM items WHERE id = %d", id))
	item := Item{}
	err := row.Scan(&item.id, &item.name, &item.price, &item.category, &item.faculty, &item.image)
	return item, err
}

func setToCart(item, count int, result *Result) {
	item_, _ := getItem(item)
	_, err := getCart(item)
	if err == sql.ErrNoRows {
		_, err = db.Exec(fmt.Sprintf("INSERT INTO cart(item,user,count,cost,status) VALUES(%d,%d,%d,%f,'cart');",
			item, user.id, count, item_.price*float32(count)))
		if err == nil {
			result.success = 1
		} else {
			log.Fatal("Error add to cart", err)
		}
	} else {
		_, err = db.Exec(fmt.Sprintf("UPDATE cart SET count = %d, cost = %f WHERE item = %d AND user = %d",
			count, item_.price*float32(count), item, user.id))
		if err == nil {
			result.success = 1
		} else {
			log.Fatal("Error update in cart", err)
		}
	}
}

func add(w http.ResponseWriter, r *http.Request) {
	result := Result{success: 0, message: ""}
	if user.id != 0 && r.Method == "GET" {
		var err error
		var id, count int

		params := r.URL.Query()

		id_, ok := params["id"]
		if !ok {
			result.message += "Не указан id товара<br>"
		} else {
			id, err = strconv.Atoi(id_[0])
			if err != nil {
				result.message += "id не верный<br>"
			}
		}

		count_, ok := params["count"]
		if !ok {
			count = 1
		} else {
			count, err = strconv.Atoi(count_[0])
		}
		if err != nil {
			result.message += "Количество указано не верно<br>"
		}

		if len(result.message) == 0 {
			setToCart(id, count, &result)
		}
	} else {
		result.message = "Авторизуйтесь и попробуйте снова"
	}
	Show(w, fmt.Sprintf("{\"success\": %d, \"message\": \"%s\"}", result.success, result.message))
}

var user User

func login(name, pass string) bool {
	row := db.QueryRow("select * from users WHERE name = '" + name + "' and pass = '" + pass + "'")
	usr := User{}
	err := row.Scan(&usr.id, &usr.name, &usr.pass)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		log.Fatal("Failed query:", err)
		return false
	}
	user = usr
	return true
}

func auth(name, pass string) bool {
	_, err := db.Exec("INSERT INTO users VALUES(0, '" + name + "','" + pass + "')")
	if err != nil {
		log.Fatal("Failed query", err)
		return false
	}
	return login(name, pass)
}

func authorize(w http.ResponseWriter, r *http.Request) {
	result := Result{success: 0, message: ""}

	if r.Method == "GET" {
		params := r.URL.Query()
		name, ok := params["name"]
		if !ok || len(name[0]) == 0 {
			result.message += "Не указано имя<br>"
		} else if len(name[0]) < 4 {
			result.message += "Имя должно содержать хотя бы 4 символа<br>"
		} else if len(name[0]) > 40 {
			result.message += "Имя слишком длинное (не должно превышать 40 символов)<br>"
		}
		pass, ok := params["pass"]
		if !ok || len(pass[0]) == 0 {
			result.message += "Не указан пароль<br>"
		} else if len(pass[0]) < 4 {
			result.message += "Пароль должен содержать хотя бы 4 символа<br>"
		} else if len(pass[0]) > 30 {
			result.message += "Пароль слишком длинный (не должен превышать 30 символов)<br>"
		}
		_, ok = params["sign-exit"]
		if ok {
			user.id = 0
			result.success = 0
			result.message = "Вы вышли из профиля"
		}
		_, ok = params["sign-up"]
		if ok && len(result.message) == 0 {
			if auth(name[0], pass[0]) {
				result.success = 1
				result.message = user.name + ":" + user.pass
			} else {
				result.message = "Не удалось зарегистрироваться, перепроверьте данные"
			}
		}
		_, ok = params["sign-in"]
		if ok && len(result.message) == 0 {
			if login(name[0], pass[0]) {
				result.success = 1
				result.message = user.name + ":" + user.pass
			} else {
				result.message = "Не удалось войти, перепроверьте данные"
			}
		}
	} else {
		result.message = "Не указаны параметры запроса"
	}

	Show(w, fmt.Sprintf("{\"success\": %d, \"message\": \"%s\"}", result.success, result.message))
}