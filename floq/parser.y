/*
 *  Synopsis:
 *	Build an abstract syntax tree for "floq" language.
 *  Note:
 *	func lookahead() ignores eof.  that is not correct.
 */

%{
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"unicode"
)

const max_name_rune_count = 127

func init() {

	//  sanity test for mapping yy tokens to name
	if yyToknames[3] != "__MIN_YYTOK" {
		corrupt("yyToknames[3]!=__MIN_YYTOK: correct yacc command?")
	}
	//  simple sanity test
	for i, nm := range yyToknames[4:] {
		if yy_name(yy_name2tok(nm)) != nm {
			corrupt("yy_name != yy_tok: %s@%d", nm, i + 4)
		}
	}

	//yyDebug = 4
}
%}

%union {
	ast		*ast
	name		string
	string
	uint64
	int
}

//  lowest numbered yytoken.  must be first in list.
%token	__MIN_YYTOK

%token	PARSE_ERROR
%token	ARG  ARG_LIST
%token	ATT  ATT_TUPLE
%token	ATT_ARRAY
%token	RUN
%token	COMMAND  COMMAND_REF
%token	CREATE
%token	EXPAND_ENV
%token	FLOW  STMT_LIST  STMT
%token	NAME  UINT64  STRING
%token	OF  LINES
%token	SCANNER  SCANNER_REF
%token	TRACER  TRACER_REF
%token	yy_TRUE  yy_FALSE  yy_AND  yy_OR  NOT
%token	EQ  NEQ  GT  GTE  LT  LTE  MATCH  NOMATCH
%token	CONCAT
%token	WHEN

%type	<uint64>	UINT64		
%type	<string>	STRING  new_name
%type	<ast>		flow

%type	<ast>		arg_list
%type	<ast>		att  atts  att_tuple  att_value  att_expr att_array_list
%type	<ast>		constant  expr
%type	<ast>		create  create_tuple
%type	<ast>		flow_stmt  create_stmt  stmt  stmt_list
%type	<ast>		scanner  command  tracer

%left		yy_AND  yy_OR
%left		EQ  NEQ  GT  GTE  LT  LTE
%left		MATCH  NOMATCH
%left		ADD  SUB
%left		MUL  DIV
%left		CONCAT
%right		NOT  EXPAND_ENV

%%

flow:
	  stmt_list
	  {
		lex := yylex.(*yyLexState)
		$1.parent = lex.ast_root
		lex.ast_root.left = $1
	  }
	;

constant:
	  UINT64
	  {
	  	$$ = &ast{
			yy_tok:		UINT64,
			uint64:		$1,
			line_no:	yylex.(*yyLexState).line_no,
		}
	  }
	|
	  STRING
	  {
	  	$$ = &ast{
			yy_tok:		STRING,
			string:		$1,
			line_no:        yylex.(*yyLexState).line_no,
		}
	  }
	|
	  EXPAND_ENV   STRING
	  {
	  	$$ = &ast{
			yy_tok:		STRING,
			string:		os.ExpandEnv($2),
			line_no:        yylex.(*yyLexState).line_no,
		}
	  }
	|
	  yy_TRUE
	  {
	  	$$ = &ast{
			yy_tok:		yy_TRUE,
			line_no:        yylex.(*yyLexState).line_no,
		}
	  }
	|
	  yy_FALSE
	  {
	  	$$ = &ast{
			yy_tok:		yy_FALSE,
			line_no:        yylex.(*yyLexState).line_no,
		}
	  }
	;

expr:
	  constant
	|
	  expr  yy_AND  expr
	  {
	  	lex := yylex.(*yyLexState)
		$$ = lex.new_rel_op(yy_AND, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  yy_OR  expr
	  {
	  	lex := yylex.(*yyLexState)
		$$ = lex.new_rel_op(yy_OR, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  LT  expr
	  {
	  	lex := yylex.(*yyLexState)
		$$ = lex.new_rel_op(LT, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  LTE  expr
	  {
	  	lex := yylex.(*yyLexState)
		$$ = lex.new_rel_op(LTE, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  EQ  expr
	  {
	  	lex := yylex.(*yyLexState)
		$$ = lex.new_rel_op(EQ, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  NEQ  expr
	  {
	  	lex := yylex.(*yyLexState)
		$$ = lex.new_rel_op(NEQ, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  GTE  expr
	  {
	  	lex := yylex.(*yyLexState)
		$$ = lex.new_rel_op(GTE, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  GT  expr
	  {
	  	lex := yylex.(*yyLexState)
		$$ = lex.new_rel_op(GT, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  MATCH  expr
	  {
	  	lex := yylex.(*yyLexState)
		$$ = lex.new_rel_op(MATCH, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  NOMATCH  expr
	  {
	  	lex := yylex.(*yyLexState)
		$$ = lex.new_rel_op(NOMATCH, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  CONCAT  expr
	  {
	  	lex := yylex.(*yyLexState)
		$$ = lex.new_rel_op(CONCAT, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  NOT  expr  %prec NOT
	  {
	  	lex := yylex.(*yyLexState)
		$$ = lex.new_rel_op(NOT, $2, nil)
		if $$ == nil {
			return 0
		}
	  }
	|
	  '('  expr  ')'
	  {
	  	$$ = $2
	  }
	;

att_expr:
	  constant
	;

att_array_list:
	  /*  empty  */
	  {
	  	$$ = &ast{
			yy_tok:		ATT_ARRAY,
			line_no:        yylex.(*yyLexState).line_no,
			array_ref:	make([]string, 0),
		}
	  }
	|
	  att_expr
	  {
	  	lex := yylex.(*yyLexState)

	  	if $1.yy_tok != STRING {
			lex.error("attribute array element not string")
			return 0
		}
	  	ar := make([]string, 1)
		ar[0] = $1.string 
	  	$$ = &ast{
			yy_tok:		ATT_ARRAY,
			line_no:        lex.line_no,
			array_ref:	ar,
		}
	  }
	|
	  att_array_list  ','  att_expr
	  {
	  	lex := yylex.(*yyLexState)

	  	if $3.yy_tok != STRING {
			lex.error("attribute array element not string")
			return 0
		}

		ar := $1.array_ref
		ar = append(ar, $3.string)
		$1.array_ref = ar
		if len(ar) > 127 {
			lex.error("attribute array > 127 elements")
			return 0
		}
		$$ = $1
	  }
	;

att_value:
	  att_expr
	|
	  '['  att_array_list  ']'
	  {
	  	$$ = $2
	  }
	;

att:
	  new_name  ':'  att_value
	  {
	  	lex := yylex.(*yyLexState)

	  	a := &ast{
			yy_tok:		ATT,
			line_no:	lex.line_no,
		}
		a.left = &ast{
				parent:		a,
				yy_tok:		NAME,
				string:		$1,
				line_no:        lex.line_no,
		}

		a.right = $3
		if a.right.yy_tok == STRING {
			c := len(a.right.string)
			format := a.left.string + ": string attribute: %s"
			if c == 0 {
				lex.error(format, "is empty")
				return 0
			}
			if c > 127 {
				lex.error(format, fmt.Sprintf("%d > 127", c))
				return 0
			}
		}
		$3.parent = a

		$$ = a
	  }
	;

atts:
	  /*  empty */
	  {
	  	$$ = &ast{
			yy_tok:         ATT_TUPLE,
			line_no:        yylex.(*yyLexState).line_no,
		}
	  }
	|
	  att
	  {
		lex := yylex.(*yyLexState)

	  	al := &ast{
			yy_tok:		ATT_TUPLE,
			line_no:	lex.line_no,
		}
		al.left = $1
		$1.parent = al

		$$ = al
	  }
	|
	  atts ','  att
	  {
		al := $1
	  	a := $3
		a.parent = al

		var an *ast
		for an = al.left;  an.next != nil;  an = an.next {}
		an.next = a
		a.prev = a

		$$ = $1
	  }
	;

att_tuple:
	  '{'  atts  '}'
	  {
	  	$$ = $2;
	  }
	;

create:
	  CREATE
	  {
		lex := yylex.(*yyLexState)
	  	$$ = &ast{
			yy_tok:		CREATE,
			line_no:        lex.line_no,
		}
	  }
	;

scanner:
	  SCANNER  OF  LINES
	  {
		lex := yylex.(*yyLexState)
	  	$$ = &ast{
			yy_tok:		SCANNER_REF,
			line_no:        lex.line_no,
			scanner_ref:	&scanner {
						split:	bufio.ScanLines,
					},
		}
	  }
	;

command:
  	  COMMAND
	  {
	  	lex := yylex.(*yyLexState)
		$$ = &ast{
			yy_tok:		COMMAND_REF,
			line_no:        lex.line_no,
			command_ref:	&command{},
		}
	  }
	;
tracer:
  	  TRACER
	  {
	  	lex := yylex.(*yyLexState)
		$$ = &ast{
			yy_tok:		TRACER_REF,
			line_no:        lex.line_no,
			tracer_ref:	&tracer{},
		}
	  }
	;
new_name:
	  NAME
	  {
	  	lex := yylex.(*yyLexState)

	  	$$ = lex.name 
	  }
	|
	  SCANNER_REF
	  {
	  	lex := yylex.(*yyLexState)

		lex.error("name exists as scanner: %s", lex.name)
		return 0
	  }
	|
	  COMMAND_REF
	  {
	  	lex := yylex.(*yyLexState)

		lex.error("name exists as command: %s", lex.name)
		return 0
	  }
	;

create_tuple:
	  /*  empty */
	  {
	  	$$ = &ast{
			yy_tok: ATT_TUPLE,
			line_no:        yylex.(*yyLexState).line_no,
		}
	  }
	|
	  att_tuple
	;

create_stmt:
	  create  tracer  new_name
	  {
		//  Note:  could production "new_name" set "name_is_name"?
	  	yylex.(*yyLexState).name_is_name = true

	  } create_tuple {
	  	
	  	lex := yylex.(*yyLexState)

		lex.name_is_name = false
		lex.name2ast[$3] = $2

		atup := $5
		atra := $2
		atup.parent = $2
		atra.left = atup

	  	atra.tracer_ref.name = $3
	  	atra.parent = $1

		$1.left = $2

		//  frisk the attibutes of tracer

		tra := atra.tracer_ref
		lex.err = atup.frisk_att("tracer:" + tra.name) 
		if lex.err != nil {
			return 0
		}

		$$ = $1
	  }
	|
	  create  scanner  new_name
	  {
		//  Note:  could production "new_name" set "name_is_name"?
	  	yylex.(*yyLexState).name_is_name = true

	  } create_tuple {
	  	
	  	lex := yylex.(*yyLexState)

		lex.name_is_name = false
		lex.name2ast[$3] = $2

		al := $5
		ascan := $2
		al.parent = $2
		ascan.left = al

	  	ascan.scanner_ref.name = $3
	  	ascan.parent = $1

		$1.left = $2

		//  frisk the attibutes of command

		scan := ascan.scanner_ref
		lex.err = al.frisk_att("scanner: " + scan.name) 
		if lex.err != nil {
			return 0
		}

		$$ = $1
	  }
	|
	  create  command  new_name
	  {
		yylex.(*yyLexState).name_is_name = true

	  } create_tuple {
	  	lex := yylex.(*yyLexState)
		lex.name_is_name = false

		c := $1
		cmd := $2
		nm := $3
		ctup := $5
		cref := cmd.command_ref

		cmd.parent = c
		c.left = cmd
		ctup.parent = cmd
		cmd.left = ctup

		lex.name2ast[nm] = cmd
		cref.name = nm

		lex.err = cref.frisk_att(ctup)
		if lex.err != nil {
			return 0
		}
		$$ = $1
	  }
	;	

flow_stmt:
	  RUN  COMMAND_REF  {
	  	lex := yylex.(*yyLexState)
		cmd := lex.name2ast[lex.name].command_ref

		if cmd == nil {
			corrupt("command: nil command ref: " + lex.name)
		}

		rc := lex.run_cmd2ast[cmd]
		if rc != nil {
			lex.error("command run twice: \"%s\"", cmd.name)
			return 0
		}

	  	$$ = &ast{
			yy_tok:		RUN,
			line_no:        yylex.(*yyLexState).line_no,
			command_ref:	cmd,
		}
	  }  '('  arg_list  ')' {
	  	lex := yylex.(*yyLexState)

	  	arun := $<ast>3
		cmd := arun.command_ref
		argv := $5

		arun.left = argv
		argv.parent = arun

		lex.run_cmd2ast[cmd] = arun 
		$$ = arun
	  }
	;

stmt:
	  /* empty */
	  {
	  	$$ = &ast{
			yy_tok:		STMT,
			line_no:        yylex.(*yyLexState).line_no,
		}
	  }
	|
	  create_stmt
	  {
	  	$$ = &ast{
			yy_tok:		STMT,
			left:		$1,
			line_no:        yylex.(*yyLexState).line_no,
		}
		$1.parent = $$
	  }
	|
	  flow_stmt
	  {
	  	$$ = &ast{
			yy_tok:		STMT,
			left:		$1,
			line_no:        yylex.(*yyLexState).line_no,
		}
		$1.parent = $$
	  }
	|
	  flow_stmt  WHEN  {
	  	$<ast>$ = &ast{
				yy_tok:		WHEN,
				line_no:	yylex.(*yyLexState).line_no,
		}
	  }  expr  {
	  	lex := yylex.(*yyLexState)
		flo := $1
	  	when := $<ast>3
		cond := $4

	  	if cond.is_bool() == false {
			lex.error("when qualification not boolean")
			return 0
		}

		flo.right = when
			
		stmt := &ast{
			yy_tok:		STMT,
			left:		flo,
			line_no:	yylex.(*yyLexState).line_no,
		}
		flo.parent = stmt
		when.parent = flo
		when.left = cond
		cond.parent = when

		$$ = stmt
	  }
	;

stmt_list:
	  stmt  ';'
	  {
	  	lex := yylex.(*yyLexState)

		stmt := $1
		sl := &ast{
			yy_tok:		STMT_LIST,
			line_no:	lex.line_no,
			left:		stmt,
			parent:		lex.ast_root,
			uint64:		1,
		}
		stmt.uint64 = 1
		stmt.parent = sl

		$$ = sl
	  }
	|
	  stmt_list  stmt  ';'
	  {
		sl := $1
		if sl.yy_tok != STMT_LIST {
			sl.corrupt("stmt_list not STMT_LIST: %s", sl.name())
		}
		stmt := $2
		stmt.parent = sl

		//  add stmt to end of STMT.
		s_tail := sl.left
		if s_tail == nil {
			sl.corrupt("left STMT is nil")
		}
		for ;  s_tail.next != nil;  s_tail = s_tail.next {}

		//  order number in list
		stmt.uint64 = s_tail.uint64 + 1

		s_tail.next = stmt
		stmt.prev = s_tail

		sl.uint64++		//  count the # stmt

		$$ = $1
	  }
	;

arg_list:
	  /*  empty */
	  {
	  	$$ = &ast{
			yy_tok:         ARG_LIST,
			line_no:        yylex.(*yyLexState).line_no,
		}
	  }
	|
	  expr
	  {
		lex := yylex.(*yyLexState)

	  	al := &ast{
			yy_tok:		ARG_LIST,
			line_no:	lex.line_no,
			uint64:		1,
		}
		al.left = $1
		$1.parent = al

		$$ = al
	  }
	|
	  arg_list  ','  expr
	  {
		al := $1
	  	e := $3
		e.parent = al

		//  find the tail of arg list
		var an *ast
		for an = al.left;  an.next != nil;  an = an.next {}
		an.next = e
		e.prev = an

		al.uint64++

		$$ = $1
	  }
	;
%%

var keyword = map[string]int{
	"and":			yy_AND,
	"Command":		COMMAND,
	"create":		CREATE,
	"ExpandEnv":		EXPAND_ENV,
	"false":		yy_FALSE,
	"lines":		LINES,
	"not":			NOT,
	"of":			OF,
	"or":			yy_OR,
	"run":			RUN,
	"Scanner":		SCANNER,
	"Tracer":		TRACER,
	"true":			yy_TRUE,
	"when":			WHEN,
}

type yyLexState struct {
	in			io.RuneReader	//  source stream
	line_no			int	   	//  lexical line number
	eof			bool       	//  seen eof in token stream
	peek			rune       	//  lookahead in lexer
	err			error

	ast_root		*ast
	name			string
	string
	uint64

	name2ast		map[string]*ast
	run_cmd2ast		map[*command]*ast
	name_is_name		bool
}

func (lex *yyLexState) pushback(c rune) {

	if lex.peek != 0 {
		corrupt("pushback(): push before peek")
	}
	lex.peek = c
	if c == '\n' {
		lex.line_no--
	}
}

/*
 *  Read next UTF8 rune.
 */
func (lex *yyLexState) get() (c rune, eof bool, err error) {

	if lex.eof {
		return 0, true, nil
	}
	if lex.peek != 0 {		/* returned stashed char */
		c = lex.peek
		/*
		 *  Only pushback 1 char.
		 */
		lex.peek = 0
		if c == '\n' {
			lex.line_no++
		}
		return c, false, nil
	}
	c, _, err = lex.in.ReadRune()
	if err != nil {
		if err == io.EOF {
			lex.eof = true
			return 0, true, nil
		}
		return 0, false, err
	}

	if c == unicode.ReplacementChar {
		return 0, false, lex.mkerror("get: invalid unicode sequence")
	}
	if c == '\n' {
		lex.line_no++
	}
	return c, false, nil
}

//  Note: end of file ignored!!

func (lex *yyLexState) lookahead(peek rune, ifyes int, ifno int) (int, error) {
	
	next, eof, err := lex.get()
	if err != nil {
		return 0, err
	}
	if next == peek {
		return ifyes, err
	}
	if !eof {
		lex.pushback(next)
	}
	return ifno, nil
}

/*
 *  Skip '#' comment.
 *  The scan stops on the terminating newline or error
 */
func skip_comment(lex *yyLexState) (err error) {
	var c rune
	var eof bool

	/*
	 *  Scan for newline, end of file, or error.
	 */
	for c, eof, err = lex.get();  !eof && err == nil;  c, eof, err = lex.get() {
		if c == '\n' {
			return
		}
	}
	return err

}

// skip over whitespace in code, complicated by # coments.

func skip_space(lex *yyLexState) (c rune, eof bool, err error) {

	for c, eof, err = lex.get();
	    !eof && err == nil;
	    c, eof, err = lex.get() {
		if unicode.IsSpace(c) {
			continue
		}
		if c != '#' {
			return c, false, nil
		}

		/*
		 *  Skipping over # comment terminated by newline or EOF
		 */
		err = skip_comment(lex)
		if err != nil {
			return 0, false, err
		}
	}
	return 0, eof, err
}

/*
 *  Very simple utf8 string scanning, with no proper escapes for characters.
 *  Expect this module to be replaced with correct text.Scanner.
 */
func (lex *yyLexState) scanner_string(yylval *yySymType) (eof bool, err error) {
	var c rune
	s := ""

	for c, eof, err = lex.get();
	    !eof && err == nil;
	    c, eof, err = lex.get() {
		if c == '"' {
			yylval.string = s
			return false, nil
		}
		switch c {
		case '\n':
			return false, lex.mkerror("new line in string")
		case '\r':
			return false, lex.mkerror("carriage return in string")
		case '\t':
			return false, lex.mkerror("tab in string")
		case '\\':
			return false, lex.mkerror("backslash in string")
		}
		s += string(c)
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

/*
 *  Scan an almost raw `...`  string as defined in golang.
 *  Carriage return is stripped.
 */
func (lex *yyLexState) scanner_raw_string(yylval *yySymType) (eof bool, err error) {
	var c rune
	s := ""

	/*
	 *  Scan a raw string of unicode letters, accepting all but `
	 */
	for c, eof, err = lex.get();
	    !eof && err == nil;
	    c, eof, err = lex.get() {
		
		switch c {
		case '\r':
			//  why does go skip carriage return?  raw is not so raw
			continue
		case '`':
			yylval.string = s
			return false, nil
		}
		s += string(c)
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

/*
 *  Scan a word consisting of a sequence of unicode Letters, Numbers and '_'
 *  characters.
 */
func (lex *yyLexState) scanner_word(
	yylval *yySymType,
	c rune,
) (tok int, err error) {
	var eof bool

	w := string(c)		//  panic() if cast fails?
	count := 1

	/*
	 *  Scan a string of unicode (?) letters, numbers/digits and '_' 
	 *  characters.
	 */
	for c, eof, err = lex.get();
	    !eof && err == nil;
	    c, eof, err = lex.get() {
		if c > 127 ||
		   (c != '_' &&
		   !unicode.IsLetter(c) &&
		   !unicode.IsNumber(c)) {
			break
		}
		count++
		if count > max_name_rune_count {
			return 0, lex.mkerror("word: too many chars: max=%d",
						max_name_rune_count)
		}
		w += string(c)		//  Note: replace with string builder?
	}
	if err != nil {
		return 0, err
	}
	if !eof {
		lex.pushback(c)		/* first character after word */
	}

	if keyword[w] > 0 {		/* got a keyword */
		return keyword[w], nil	/* return yacc generated token */
	}

	lex.name = w
	if lex.name_is_name == false && lex.name2ast[w] != nil {
		yylval.name = w
		return lex.name2ast[w].yy_tok, nil
	}
	return NAME, nil
}

func (lex *yyLexState) scanner_uint64(yylval *yySymType, c rune) (err error) {
	var eof bool

	ui64 := string(c)
	count := 1

	/*
	 *  Scan a string of unicode numbers/digits and let Scanf parse the
	 *  actual digit string.
	 */
	for c, eof, err = lex.get();
	    !eof && err == nil;
	    c, eof, err = lex.get() {
		count++
		if count > 20 {
			return lex.mkerror("uint64 > 20 digits")
		}
		if c > 127 || !unicode.IsNumber(c) {
			break
		}
		ui64 += string(c)
	}
	if err != nil {
		return
	}
	if !eof {
		lex.pushback(c)		//  first character after ui64
	}

	yylval.uint64, err = strconv.ParseUint(ui64, 10, 64)
	return
}

func (lex *yyLexState) new_rel_op(tok int, left, right *ast) (a *ast) {

	switch tok {
	case NOT:
		if left.is_bool() == false {
			lex.line_no = left.line_no
			lex.error("NOT: can not negate %s", left.name())
			return nil
		}
	case yy_AND, yy_OR:
		if left.is_bool() == false {
			lex.line_no = left.line_no
			lex.error(
				"%s: left expr not bool: got %s, want BOOL",
				yy_name(tok),
				left.name(),
			)
			return nil
		}
		if right.is_bool() == false {
			lex.line_no = right.line_no
			lex.error(
				"%s: right expr not bool: got %s, want BOOL",
				yy_name(tok),
				right.name(),
			)
			return nil
		}
	case EQ, NEQ, LT, LTE, GTE, GT:
		can_compare := (left.is_string() && right.is_string()) ||
		               (left.is_uint64() && right.is_uint64()) ||
		               (left.is_bool() && right.is_bool())
		if !can_compare {
			lex.line_no = right.line_no
			lex.error(
				"%s: can not compare %s and %s",
				yy_name(tok),
				left.name(),
				right.name(),
			)
			return nil
		}
	case CONCAT, MATCH, NOMATCH:
		if left.is_string() == false {
			lex.line_no = left.line_no
			lex.error("%s: left is not string", left.name())
			return nil
		}
		if right.is_string() == false {
			lex.line_no = right.line_no
			lex.error("%s: right is not string", right.name())
			return nil
		}
	default:
		msg := fmt.Sprintf("new_rel_op:  yy token: %s", yy_name(tok))
		corrupt(msg)
		return nil	//  NOTREACHED
	}

	a = &ast{
		yy_tok:		tok,
		left:		left,
		right:		right,
		line_no:	left.line_no,	//  ought to be op line no
	}
	left.parent = a
	if right != nil {
		right.parent = a
	}
	return a
}

//  lexical scan of a token

func (lex *yyLexState) Lex(yylval *yySymType) (tok int) {

	if lex.err != nil {
		return PARSE_ERROR
	}
	if lex.eof {
		return 0
	}
	yylval.name = ""
	c, eof, err := skip_space(lex)
	if err != nil {
		goto LEX_ERROR
	}
	if eof {
		return 0
	}

	switch {

	//  ascii outside of strings, for time being (why?)
	case c > 127:
		err = lex.mkerror("char not ascii: 0x%x", c)
		goto LEX_ERROR

	case c == '=':
		//  clang "==" equality

		tok, err = lex.lookahead('=', EQ, 0)
		if err != nil {
			goto LEX_ERROR
		}
		if tok != 0 {
			return tok
		}

		//  expr regexp operator "=~"

		tok, err = lex.lookahead('~', MATCH, '=')
		if err != nil {
			goto LEX_ERROR
		}
		return tok

	case c == '!':
		//  clang inequality "!="

		tok, err = lex.lookahead('=', NEQ, 0)
		if err != nil {
			goto LEX_ERROR
		}
		if tok != 0 {
			return tok
		}

		//  expr negate match regexp operator "!~"

		tok, err = lex.lookahead('~', NOMATCH, '!')
		if err != nil {
			goto LEX_ERROR
		}
		return tok

	case c == '|':
		tok, err = lex.lookahead('|', CONCAT, '|')
		if err != nil {
			goto LEX_ERROR
		}
		return tok

	case c == '>':
		tok, err = lex.lookahead('=', GTE, GT)
		if err != nil {
			goto LEX_ERROR
		}
		return tok

	case c == '<':

		tok, err = lex.lookahead('=', LTE, LT)
		if err != nil {
			goto LEX_ERROR
		}
		return tok

	case unicode.IsLetter(c) || c == '_':
		tok, err = lex.scanner_word(yylval, c)
		if err != nil {
			goto LEX_ERROR
		}
		return tok

	case unicode.IsNumber(c):
		err = lex.scanner_uint64(yylval, c)
		if err != nil {
			goto LEX_ERROR
		}
		return UINT64

	case c == '"':
		lno := lex.line_no	// reset line number on error

		eof, err = lex.scanner_string(yylval)
		if err != nil {
			goto LEX_ERROR
		}
		if eof {
			lex.line_no = lno
			err = lex.mkerror("unexpected end of file in string")
			goto LEX_ERROR
		}
		return STRING

	case c == '`':
		lno := lex.line_no	// reset line number on error

		eof, err = lex.scanner_raw_string(yylval)
		if err != nil {
			goto LEX_ERROR
		}
		if eof {
			lex.line_no = lno
			err = lex.mkerror("end of file in raw string")
			goto LEX_ERROR
		}
		return STRING
	}

	return int(c)

LEX_ERROR:
	lex.err = err
	return PARSE_ERROR
}

func (lex *yyLexState) mkerror(format string, args...interface{}) error {

	return errors.New(fmt.Sprintf("%s, near line %d",
		fmt.Sprintf(format, args...),
		lex.line_no,
	))
}

func (lex *yyLexState) error(format string, args...interface{}) {

	lex.Error(fmt.Sprintf(format, args...))
}

func (lex *yyLexState) Error(msg string) {

	if lex.err == nil {			//  only report first error
		lex.err = lex.mkerror("%s", msg)
	}
}

func parse(in io.RuneReader) (*ast, error) {

	lex := &yyLexState{
		in:		in,
		line_no:	1,
		name2ast:	make(map[string]*ast),
		run_cmd2ast:	make(map[*command]*ast),
		ast_root:	&ast{
					yy_tok:		FLOW,
					line_no:	1,
				},
	}
	yyParse(lex)
	return lex.ast_root, lex.err
}

func yy_name(tok int) (name string) {
	//  print token name or int value of yy token
	offset := tok - __MIN_YYTOK + 3
	if (offset < len(yyToknames) && tok > __MIN_YYTOK) {
		name = yyToknames[offset]
	} else {
		name = fmt.Sprintf( "UNKNOWN(%d)", tok)
	}
	return
}
func yy_names(toks ...int) (names string) {

       for _, tok := range toks {
               if names == "" {
                       names = yy_name(tok)
               } else {
                       names = names + "," + yy_name(tok)
               }
       }
       return names
}

func yy_name2tok(name string) int {

	for i, nm := range yyToknames {
		if nm == name {
			return i + __MIN_YYTOK - 3
		}
	}
	return __MIN_YYTOK - 2	// == "error" in yyToknames
}

func WTF(format string, args ...interface{}) {

	format = "WTF: " + format
	fmt.Fprintf(os.Stderr, fmt.Sprintf(format, args...) + "\n")
}
