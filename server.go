package main

import (
	"net/http"
	"path/filepath"
	"encoding/json"
	"io/ioutil"
  "os"
  "time"
	
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	jwt "github.com/golang-jwt/jwt/v4"
)


type (
  Login struct {
    Username string `json:"username"`
    Password string `json:"password"`
  }

  User struct {
  	Username string `json:"username"`
    Password string `json:"password"`
    Email 	 string `json:"email"`
  }

  UserList struct {
  	Users []User
  }

  jwtCustomClaims struct {
		Username string `json:"username"`
		Email  	 string `json:"email"`
		jwt.StandardClaims
	}
)

var baseURL = "http://dev3.dansmultipro.co.id";

func login(c echo.Context) error {
	u := new(Login)
  
  if err := c.Bind(u); err != nil {
    return echo.NewHTTPError(http.StatusBadRequest, err.Error())
  }
	
	basePath, _ := os.Getwd()
  dbPath := filepath.Join(basePath, "users.json")
  buf, _ := ioutil.ReadFile(dbPath)

  var data UserList
  err := json.Unmarshal(buf, &data)

  if err != nil {
    return err
  }

  var userFound = false;
  var userData User;
  for _, v := range data.Users {
    if v.Username == u.Username {
    	userFound = true;
      if (v.Password != u.Password) {
      	return echo.ErrUnauthorized;
      } else {
      	userData = v;
      }
    }
	}

	if (!userFound) {
		return echo.ErrNotFound;
	}

	claims := &jwtCustomClaims{
		userData.Username,
		userData.Email,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}

func jobs(c echo.Context) error {
	search := c.QueryParam("search");
	location := c.QueryParam("location");
	full_time := c.QueryParam("full_time");
	page := c.QueryParam("page");

	var err error
  var client = &http.Client{}
  var data interface{}

  request, err := http.NewRequest("GET", baseURL+"/api/recruitment/positions.json", nil)
  if err != nil {
    return err
  }

  q := request.URL.Query()

  if (search != "") {
		q.Add("description", search)
  }

  if (location != "") {
		q.Add("location", location)
  }

	if (full_time != "" && full_time == "true") {
		q.Add("type", "Full Time")
  }

  if (page != "") {
		q.Add("page", page)
  }

	request.URL.RawQuery = q.Encode()

  response, err := client.Do(request)
  if err != nil {
    return err
  }
  defer response.Body.Close()

  err = json.NewDecoder(response.Body).Decode(&data)
  if err != nil {
      return err
  }
  return c.JSON(http.StatusOK, data)
}

func jobsDetail(c echo.Context) error {
	var err error
  var client = &http.Client{}
  var data interface{}

  request, err := http.NewRequest("GET", baseURL+"/api/recruitment/positions/"+c.Param("id"), nil)
  if err != nil {
    return err
  }

  response, err := client.Do(request)
  if err != nil {
    return err
  }
  defer response.Body.Close()

  err = json.NewDecoder(response.Body).Decode(&data)
  if err != nil {
      return err
  }

  return c.JSON(http.StatusOK, data)
}

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Dans Multi Pro!")
	})
	e.POST("/login", login)

	r := e.Group("/jobs")

	config := middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: []byte("secret"),
	}

	r.Use(middleware.JWTWithConfig(config))

	r.GET("", jobs)
	r.GET("/:id", jobsDetail)

	e.Logger.Fatal(e.Start(":1323"))
}