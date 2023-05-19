package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

//  Note: added SQLQuerySetFileFilter for slurping all matching sql files
//        in a directory.

type Config struct {
	source_path string

	Synopsis        string `json:"synopsis"`
	HTTPListen      string `json:"http_listen"`
	RESTPathPrefix  string `json:"rest_path_prefix"`
	SQLQuerySet     `json:"sql_query_set"`
	HTTPQueryArgSet `json:"http_query_arg_set"`

	BasicAuthPath string `json:"basic_auth_path"`

	basic_auth map[string]string

	//  Note:  also want to log slow http requests!
	//         consider moving into WARN section.

	WarnSlowSQLQueryDuration float64 `json:"warn_slow_sql_query_duration"`

	//  https paramters

	TLSHTTPListen string `json:"tls_http_listen"`
	TLSCertPath   string `json:"tls_cert_path"`
	TLSKeyPath    string `json:"tls_key_path"`
}

func (cf *Config) load(path string) {

	INFO("loading config file: %s", path)

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
	INFO("http listen: %s", cf.HTTPListen)

	if cf.RESTPathPrefix == "" {
		cf.RESTPathPrefix = "/"
	}
	INFO("rest path prefix: %s", cf.RESTPathPrefix)
	INFO("warn slow sql query duration: %0.9fs",
		cf.WarnSlowSQLQueryDuration,
	)

	cf.SQLQuerySet.load()
	cf.HTTPQueryArgSet.load()

	//  wire up sql aliases for the http query arguments

	INFO("bind http query args to sql variables query args ...")
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
				INFO("  %s -> %s", qa.path, ha.name)
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

	INFO("%s: loaded", cf.source_path)
}

func (cf *Config) check_basic_auth(
	w http.ResponseWriter,
	r *http.Request,
) bool {

	if cf.BasicAuthPath == "" {
		return true
	}

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

	ba_INFO := func(format string, args ...interface{}) {

		INFO("basic auth: "+format, args...)
	}

	ba_die := func(format string, args ...interface{}) {

		die("basic auth: "+format, args...)
	}

	cf.basic_auth = nil
	if cf.BasicAuthPath == "" {
		ba_INFO("password file not defined")
		ba_INFO("no password required to access queries")
		return
	}
	cf.basic_auth = make(map[string]string)
	ba_INFO("password file: %s", cf.BasicAuthPath)

	f, err := os.Open(cf.BasicAuthPath)
	if err != nil {
		ba_die("can not open password file: %s", err)
	}
	defer f.Close()

	in := bufio.NewReader(f)
	white_re := regexp.MustCompile(`^\s*$`)
	comment_re := regexp.MustCompile(`^\s*#`)
	entry_re := regexp.MustCompile(`^([[:alpha:]0-9]{1,32}):(..*)`)

	ba_INFO("loading passwords ...")
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
	INFO("loaded %d password entries", len(cf.basic_auth))
}

func (cf *Config) handle_query_index_json(
	w http.ResponseWriter,
	r *http.Request,
) {
	var columns = [...]string{
		"name",
		"synopsis",
		"description",
	}

	now := time.Now()
	if !cf.check_basic_auth(w, r) {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	reply := &JSONQueryReply{
		Status:		"ok",
		Columns:	columns[:],
	}
	for _, q := range cf.SQLQuerySet {
		var r [3]interface{}

		r[0], r[1], r[2] = q.name, q.synopsis, q.description
		reply.Rows = append(reply.Rows, r[:])
	}
	reply.Duration = time.Since(now).Seconds()

	buf, err := json.Marshal(reply)
	if err != nil {
		panic(err)
	}
	_, err = w.Write(buf)
	if err != nil {
		ERROR("write(handle_query_json) to %s failed: %s",
			r.RemoteAddr,
			err,
		)
		return
	}
}

const query_index_thead = `
<table>
 <caption>Index of %d REST Quer%s</caption>
 <thead>
  <tr>
   <th>Name</th>
   <th>Synopsis</th>
   <th>Description</th>
  </tr>
 </thead>
 <tbody>
`

const query_index_tr = `
  <tr>
   <td>%s</td>
   <td>%s</td>
   <td>%s</td>
  </tr>
`

// Note: is a colspan=3 needed?
const query_index_tfoot = `
 </tbody>
 <tfoot>
  <tr>
   <th>%s to Execute</th>
  </tr>
 </tfoot>
</table>
`

func (cf *Config) handle_query_index_html(
	w http.ResponseWriter,
	r *http.Request,
) {
	now := time.Now()

	if !cf.check_basic_auth(w, r) {
		return
	}

	w.Header().Set("Content-Type", "text/html;  charset=utf-8")

	var plural = "ies"
	if len(cf.SQLQuerySet) == 1 {
		plural = "y"
	}
	buf := bytes.NewBufferString(fmt.Sprintf(
			query_index_thead,
			len(cf.SQLQuerySet),
			plural,
	))

	// build the <tr> row set
	for _, q := range cf.SQLQuerySet {
		buf.Write([]byte(fmt.Sprintf(query_index_tr,
				html.EscapeString(q.name),
				html.EscapeString(q.synopsis),
				html.EscapeString(q.description),
		)))
	}

	// build the <tfoot> footer
	buf.Write([]byte(fmt.Sprintf(query_index_tfoot,
			time.Since(now),
	)))

	//  write full html to client
	_, err := w.Write(buf.Bytes())
	if err != nil {
		ERROR("write(handle_query_json) to %s failed: %s",
			r.RemoteAddr,
			err,
		)
		return
	}
}

func (cf *Config) handle_query_index_tsv(
	w http.ResponseWriter,
	r *http.Request,
) {
	if !cf.check_basic_auth(w, r) {
		return
	}

	w.Header().Set("Content-Type",
				"text/tab-separated-values; charset=utf-8")

	buf := bytes.NewBufferString("name\tsynopsis\tdescription\n")

	// build the row set
	for _, q := range cf.SQLQuerySet {
		buf.Write([]byte(q.name))
		buf.Write([]byte("\t"))
		buf.Write([]byte(q.synopsis))
		buf.Write([]byte("\t"))
		buf.Write([]byte(q.description))
		buf.Write([]byte("\n"))
	}

	//  write full tsv to client
	_, err := w.Write(buf.Bytes())
	if err != nil {
		ERROR("write(handle_query_tsv) to %s failed: %s",
			r.RemoteAddr,
			err,
		)
		return
	}
}

func (cf *Config) handle_query_index_csv(
	w http.ResponseWriter,
	r *http.Request,
) {
	if !cf.check_basic_auth(w, r) {
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")

	var buf bytes.Buffer
	csv := csv.NewWriter(&buf)

	var row [3]string
	row[0] = "name"
	row[1] = "synopsis"
	row[2] = "description"
	csv.Write(row[:])

	for _, q := range cf.SQLQuerySet {
		row[0], row[1], row[2] = q.name, q.synopsis, q.description
		err := csv.Write(row[:])
		if err != nil {
			panic(err)
		}
	}
	csv.Flush()

	//  write full csv to client
	_, err := w.Write(buf.Bytes())
	if err != nil {
		ERROR("write(handle_query_csv) to %s failed: %s",
			r.RemoteAddr,
			err,
		)
		return
	}
}
