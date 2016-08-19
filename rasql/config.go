package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Config struct {
	source_path string

	Synopsis        string `json:"synopsis"`
	HTTPListen      string `json:"http-listen"`
	HTTPListenTLS   string `json:"http-listen-tls"`
	RESTPathPrefix  string `json:"rest-path-prefix"`
	SQLQuerySet     `json:"sql-query-set"`
	HTTPQueryArgSet `json:"http-query-arg-set"`
	BasicAuthPath   string `json:"basic-auth-path"`

	basic_auth map[string]string

	//  Note:  also want to log slow http requests!
	//         consider moving into WARN section.

	WarnSlowSQLQueryDuration float64 `json:"warn-slow-sql-query-duration"`
	TLSCertPath              string  `json:"tls-cert-path"`
	TLSKeyPath               string  `json:"tls-key-path"`
}

func (cf *Config) load(path string) {

	log("loading config file: %s", path)

	cf.source_path = path

	//  slurp config file into string

	b, err := ioutil.ReadFile(cf.source_path)
	if err != nil {
		die("config read failed: %s", err)
	}

	//  decode json in config file

	dec := json.NewDecoder(strings.NewReader(string(b)))
	err = dec.Decode(&cf)
	if err != nil && err != io.EOF {
		die("config json decoding failed: %s", err)
	}

	if cf.HTTPListen == "" {
		die("config: http-listen not defined or empty")
	}
	log("http listen: %s", cf.HTTPListen)

	if cf.RESTPathPrefix == "" {
		cf.RESTPathPrefix = "/"
	}
	log("rest path prefix: %s", cf.RESTPathPrefix)
	log("warn slow sql query duration: %0.9fs", cf.WarnSlowSQLQueryDuration)

	cf.SQLQuerySet.load()
	cf.HTTPQueryArgSet.load()

	//  wire up sql aliases for the http query arguments

	log("bind http query args to sql variables query args ...")
	for _, ha := range cf.HTTPQueryArgSet {
		a := ha.SQLAlias
		if a == "" {
			if ha.Matches == "" {
				die("query arg: missing \"matches\" regexp: %s",
					ha.name)
			}
			continue
		}

		//  point sql arguments to current http query argument

		found := false
		for _, q := range cf.SQLQuerySet {
			for _, qa := range q.SQLQueryArgSet {
				if qa.path != a {
					continue
				}
				log("  %s -> %s", qa.path, ha.name)
				found = true
				qa.http_arg = ha
			}
		}
		if !found {
			die(`http arg "%s": no sql variable for alias "%s"`,
				ha.name, a)
		}
	}
	cf.load_auth()

	log("%s: loaded", cf.source_path)
}

func (cf *Config) check_basic_auth(
	w http.ResponseWriter,
	r *http.Request,
) bool {

	user, pass, _ := r.BasicAuth()
	if user != "" && cf.basic_auth[user] == pass {
		return true
	}

	//  tell client to authenticate.  only grumble if user attempted
	//  a login and password.  not clear if failed empty attempts ought
	//  to be logged.

	w.Header().Set("WWW-Authenticate", "Basic realm=\"rasql\"")
	if user != "" {
		reply_ERROR(http.StatusUnauthorized, w, r,
			"either password or user %s is invalid", user)
	} else {
		http.Error(w, "missing authorization", http.StatusUnauthorized)
	}
	return false
}

func (cf *Config) new_handler_query_json(sqlq *SQLQuery) http.HandlerFunc {

	//  no authorization required

	if len(cf.basic_auth) == 0 {
		return func(w http.ResponseWriter, r *http.Request) {
			sqlq.handle_query_json(w, r, cf)
		}
	}

	//  authorization required before handling the query

	return func(w http.ResponseWriter, r *http.Request) {

		if cf.check_basic_auth(w, r) == false {
			return
		}
		sqlq.handle_query_json(w, r, cf)
	}
}

func (cf *Config) new_handler_query_tsv(sqlq *SQLQuery) http.HandlerFunc {

	//  no authorization required

	if len(cf.basic_auth) == 0 {
		return func(w http.ResponseWriter, r *http.Request) {
			sqlq.handle_query_tsv(w, r, cf)
		}
	}

	//  authorization required before handling the query

	return func(w http.ResponseWriter, r *http.Request) {

		if cf.check_basic_auth(w, r) == false {
			return
		}
		sqlq.handle_query_tsv(w, r, cf)
	}
}

func (cf *Config) new_handler_query_csv(sqlq *SQLQuery) http.HandlerFunc {

	//  no authorization required

	if len(cf.basic_auth) == 0 {
		return func(w http.ResponseWriter, r *http.Request) {
			sqlq.handle_query_csv(w, r, cf)
		}
	}

	//  authorization required before handling the query

	return func(w http.ResponseWriter, r *http.Request) {

		if cf.check_basic_auth(w, r) == false {
			return
		}
		sqlq.handle_query_csv(w, r, cf)
	}
}

func (cf *Config) new_handler_query_html(sqlq *SQLQuery) http.HandlerFunc {

	//  no authorization required

	if len(cf.basic_auth) == 0 {
		return func(w http.ResponseWriter, r *http.Request) {
			sqlq.handle_query_html(w, r, cf)
		}
	}

	//  authorization required before handling the query

	return func(w http.ResponseWriter, r *http.Request) {

		if cf.check_basic_auth(w, r) == false {
			return
		}
		sqlq.handle_query_html(w, r, cf)
	}
}

//  Note: consider using package https://github.com/abbot/go-http-auth

func (cf *Config) load_auth() {

	ba_log := func(format string, args ...interface{}) {

		log("basic auth: %s", fmt.Sprintf(format, args...))
	}

	ba_die := func(format string, args ...interface{}) {

		die("basic auth: %s", fmt.Sprintf(format, args...))
	}

	cf.basic_auth = nil
	if cf.BasicAuthPath == "" {
		ba_log("password file not defined")
		ba_log("no password required to access queries")
		return
	}
	cf.basic_auth = make(map[string]string)
	ba_log("password file: %s", cf.BasicAuthPath)

	f, err := os.Open(cf.BasicAuthPath)
	if err != nil {
		ba_die("can not open password file: %s", err)
	}
	defer f.Close()

	in := bufio.NewReader(f)
	white_re := regexp.MustCompile(`^\s*$`)
	comment_re := regexp.MustCompile(`^\s*#`)
	entry_re := regexp.MustCompile(`^([[:alpha:]0-9]{1,32}):(..*)`)

	ba_log("loading passwords ...")
	lc := uint32(0)
	for {
		line, err := in.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			ba_die("error reading password: %s", err)
		}
		lc++
		if white_re.MatchString(line) || comment_re.MatchString(line) {
			continue
		}
		fields := entry_re.FindStringSubmatch(line)
		if fields == nil {
			die("syntax error in password file near line %d", lc)
		}
		if len(fields) != 3 {
			panic("len(password entry) != 3")
		}
		cf.basic_auth[fields[1]] = fields[2]
	}
	log("loaded %d password entries", len(cf.basic_auth))
}

//  Note: why only json output?  Need to generalize!

func (cf *Config) handle_query_index_json(
	w http.ResponseWriter,
	r *http.Request,
) {
	if !cf.check_basic_auth(w, r) {
		return
	}
	putf := func(format string, args ...interface{}) {
		fmt.Fprintf(w, format, args...)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	//  write bytes string to client

	putb := func(b []byte) {
		_, err := w.Write(b)
		if err != nil {
			panic(err)
		}
	}

	puts := func(s string) {
		putb([]byte(s))
	}

	// put json string to client

	putjs := func(s string) {
		b, err := json.Marshal(s)
		if err != nil {
			panic(err)
		}
		putb(b)
	}

	//  write a json array to the client

	puta := func(a []string) {
		b, err := json.Marshal(a)
		if err != nil {
			panic(err)
		}
		putb(b)
	}

	puts(`[
    "duration,colums,rows",
    0.0,
    `,
	)

	//  write the columns

	var columns = [...]string{
		"path",
		"synopsis",
		"description",
	}

	puta(columns[:])
	puts(",\n\n    [\n")

	count := uint64(0)
	for n, q := range cf.SQLQuerySet {

		if count > 0 {
			putf(",\n")
		}
		count++

		puts("[")
		putjs(n)
		puts(",")
		putjs(q.synopsis)
		puts(",")
		putjs(q.description)

		putf("]")
	}
	putf("\n    ]\n]\n")
}
