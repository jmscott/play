/*
 *  Synopsis:
 *	Build an abstract syntax tree for "floq" language.
 *  Note:
 *	In "set" ast duplicates can exist!
 *
 *	Parser appears to be re-entrant.  However, only single static grammar
 *	per process.  for example, array yyToknames['] is global.  ughh.
 *
 *	func lookahead() ignores eof.  that is not correct.
 */

%{
package main

import (
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
	command_ref	*command
	tuple_ref	*tuple
	sysatt		*sysatt
}

//  lowest numbered yytoken.  must be first in list.
%token	__MIN_YYTOK

%token	PARSE_ERROR
%token	ARGV
%token	yy_SET  ARRAY  
%token	RUN  START
%token	COMMAND  COMMAND_REF
%token	TUPLE  TUPLE_REF
%token	DEFINE   AS
%token	EXPAND_ENV
%token	FLOW  STMT_LIST
%token	UINT64  STRING  NAME
%token	PROJECT_OSX  PROJECT_ATT
%token	yy_TRUE  yy_FALSE  yy_AND  yy_OR  NOT  yy_EMPTY
%token	yy_STRING  CAST
%token	EQ  NEQ  GT  GTE  LT  LTE  MATCH  NOMATCH
%token	CONCAT
%token	WHEN
%token	PROJECT_OSX_EXIT_CODE  CAST_BOOL  CAST_UINT64  CAST_STRING
%token	yy_IS  yy_NULL  IS_NULL
%token	IS_NULL  IS_NULL_UINT64  IS_NULL_STRING  IS_NULL_BOOL
%token	IS_NOT_NULL  IS_NOT_NULL_UINT64  IS_NOT_NULL_STRING  IS_NOT_NULL_BOOL

%type	<uint64>	UINT64		
%type	<string>	STRING  name
%type	<ast>		flow
%type	<ast>		arg_list
%type	<ast>		element  element_list
%type	<ast>		array_element  array_element_list
%type	<ast>		set  array
%type	<ast>		constant  expr  qualification
%type	<ast>		stmt  stmt_list
%type	<command_ref>	COMMAND_REF
%type	<tuple_ref>	TUPLE_REF
%type	<sysatt>	PROJECT_OSX

%nonassoc		yy_IS
%left			yy_OR  yy_AND
%left			EQ  NEQ  GT  GTE  LT  LTE  MATCH  NOMATCH
%left			CONCAT
%right			NOT  EXPAND_ENV  CAST

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
	  	$$ = yylex.(*yyLexState).ast(UINT64)
		$$.uint64 = $1
	  }
	|
	  STRING
	  {
	  	$$ = yylex.(*yyLexState).ast(STRING)
	  	$$.string = $1
	  }
	|
	  EXPAND_ENV   STRING
	  {
	  	$$ = yylex.(*yyLexState).ast(STRING)
		$$.string = os.ExpandEnv($2)
	  }
	|
	  yy_TRUE
	  {
	  	$$ = yylex.(*yyLexState).ast(yy_TRUE)
	  }
	|
	  yy_FALSE
	  {
	  	$$ = yylex.(*yyLexState).ast(yy_FALSE)
	  }
	;

expr:
	  constant
	|
	  expr  CAST  yy_STRING  %prec CAST
	  {
	  	lex := yylex.(*yyLexState)

		expr := $1

	  	$$ = lex.ast(CAST, expr, lex.ast(yy_STRING))
	  }
	|
	  COMMAND_REF  '$'  {
	  	yylex.(*yyLexState).name_is_name = true
	  }  name  {
		lex := yylex.(*yyLexState)
		cmd := $1
		name := $4

		if cmd.is_sysatt(name) == false {
			lex.error(
				"command: %s, unknown system attribute: %s",
				cmd.name,
				name,
			)
			return 0
		}

		csa := lex.ast(PROJECT_OSX)
		csa.name = name
		csa.sysatt_ref = &sysatt{
					name:           name,
					command_ref:    cmd,
				}
		$$ = csa
	  }
	|
	  COMMAND_REF  '.'  {
	  	yylex.(*yyLexState).name_is_name = true
	  }  name  {
		lex := yylex.(*yyLexState)
		cmd := $1
		name := $4
		tup := cmd.tuple_ref

		if tup == nil {
			lex.error(
				"command: %s: no tuple defined: %s",
				cmd.name,
				name,
			)
			return 0
		}

		if tup.atts[name] == nil {
			lex.error(
				"command: %s: no attribute defined: %s",
				cmd.name,
				name,
			)
			return 0
		}
		csa := lex.ast(PROJECT_ATT)
		csa.name = name
		$$ = csa
	  }
	|
	  expr  yy_AND  expr
	  {
		$$ = yylex.(*yyLexState).new_rel_op(yy_AND, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  yy_OR  expr
	  {
		$$ = yylex.(*yyLexState).new_rel_op(yy_OR, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  LT  expr
	  {
		$$ = yylex.(*yyLexState).new_rel_op(LT, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  LTE  expr
	  {
		$$ = yylex.(*yyLexState).new_rel_op(LTE, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  EQ  expr
	  {
		$$ = yylex.(*yyLexState).new_rel_op(EQ, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  NEQ  expr
	  {
		$$ = yylex.(*yyLexState).new_rel_op(NEQ, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  GTE  expr
	  {
		$$ = yylex.(*yyLexState).new_rel_op(GTE, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  GT  expr
	  {
		$$ = yylex.(*yyLexState).new_rel_op(GT, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  MATCH  expr
	  {
		$$ = yylex.(*yyLexState).new_rel_op(MATCH, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  NOMATCH  expr
	  {
		$$ = yylex.(*yyLexState).new_rel_op(NOMATCH, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  CONCAT  expr
	  {
		$$ = yylex.(*yyLexState).new_rel_op(CONCAT, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  NOT  expr  %prec NOT
	  {
		$$ = yylex.(*yyLexState).new_rel_op(NOT, $2, nil)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  yy_IS  yy_NULL  %prec yy_IS
	  {
	  	$$ = yylex.(*yyLexState).new_rel_op(IS_NULL, $1, nil)
		if $$ == nil {
			return 0
		}
	  }
	|
	  expr  yy_IS  NOT  yy_NULL  %prec yy_IS
	  {
	  	$$ = yylex.(*yyLexState).new_rel_op(IS_NOT_NULL, $1, nil)
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

qualification:
	  /*  empty  */
	  {
	  	$$ = nil
	  }
	|
	  WHEN  expr
	  {
	  	lex := yylex.(*yyLexState)

		//  Note:  move to ast.frisk()
	  	if $2.is_bool() == false {
			lex.error("when: qualification not boolean")
			return 0
		}
		$$ = lex.ast(WHEN, $2)
	  }
	;

name:
	  NAME
	  {
	  	lex := yylex.(*yyLexState)

		lex.name_is_name = false
	  	$$ = lex.name 
	  }
	;

array_element:
	  constant
	|
	  array
	|
	  set
	;
array:
	  '['  array_element_list  ']'
	  {
		$$ = $2
	  }
	;

array_element_list:
	  /*  empty  */
	  {
	  	$$ = yylex.(*yyLexState).ast(ARRAY)
	  }
	|
	  array_element
	  {
	  	$$ = yylex.(*yyLexState).ast(ARRAY, $1)
	  }
	|
	  array_element_list  ','  array_element
	  {
		$1.push_left($3)
	  }
	;

element:
	  array_element
	|
	  name  ':'  array_element
	  {
	  	$3.name = $1
		$$ = $3
	  }
	;

//  Note: checking for duplicates?!!

element_list:
	  /*  empty  */
	  {
	  	$$ = yylex.(*yyLexState).ast(yy_SET)
	  }
	|
	  element
	  {
	  	$$ = yylex.(*yyLexState).ast(yy_SET, $1)
	  }
	|
	  element_list  ','  element
	  {
		$1.push_left($3)
	  }
	;

set:
	  '{'  element_list  '}'
	  {
	  	$$ = $2
	  }
	;

stmt:
	  DEFINE  TUPLE  name  AS  set
	  {
	  	var err error

	  	lex := yylex.(*yyLexState)

	  	define := lex.ast(DEFINE, lex.ast(TUPLE), $5)

		tup := define.left
		tup.tuple_ref, err = new_tuple($3, $5)
		if err != nil {
			lex.error("tuple %s: %s", $3, err)
			return 0
		}
		lex.name2ast[$<string>3] = tup
		lex.name2tuple[$<string>3] = tup.tuple_ref
		tup.tuple_ref.atts = make(map[string]*attribute)

		$$ = define
	  }
	|
	  DEFINE  COMMAND  name  AS  set
	  {
	  	lex := yylex.(*yyLexState)
		name := $3
		set := $5

	  	define := lex.ast(DEFINE, lex.ast(COMMAND), set)

		cmd := define.left
		cmd.name = name
		cmd.command_ref = &command{
					name: name,
					path: set.string_element("path"),
					args: set.array_string_element("args"),
					env: set.array_string_element("env"),
				}
		cf := cmd.command_ref
		lex.name2cmd[name] = cf
		lex.name2ast[name] = cmd

		$$ = define
	  }
	|
	  DEFINE  COMMAND  name  '.'  TUPLE_REF  AS  set
	  {
	  	lex := yylex.(*yyLexState)
		name := $3
		set := $7

	  	define := lex.ast(DEFINE, lex.ast(COMMAND), set)

		cmd := define.left
		cmd.name = name
		cmd.command_ref = &command{
					name: name,
					path: set.string_element("path"),
					args: set.array_string_element("args"),
					env: set.array_string_element("env"),
					tuple_ref: $<tuple_ref>5,
				}
		cr := cmd.command_ref
		lex.name2cmd[name] = cr
		lex.name2ast[name] = cmd
		cr.tuple_ref = $5
		$$ = define
	  }
	|
	  RUN  name
	  {
	  	lex := yylex.(*yyLexState)

	  	lex.error("run: command not defined: %s", lex.name)
		return 0
	  }
	|
	  RUN  COMMAND_REF  '('  arg_list  ')'  qualification  {
	  	lex := yylex.(*yyLexState)

		run := lex.ast(RUN, $4, $6)
		run.command_ref = $2
		run.name = run.command_ref.name
		$$ = run
	  }
	|
	  START  name
	  {
	  	lex := yylex.(*yyLexState)

	  	lex.error("start: command not defined: %s", lex.name)
		return 0
	  }
	|
	  START  COMMAND_REF  '('  arg_list  ')'  qualification  {
	  	lex := yylex.(*yyLexState)

		start := lex.ast(START, $4, $6)
		start.command_ref = $2
		start.name = start.command_ref.name
		$$ = start
	  }
	;

stmt_list:
	  stmt  ';'
	  {
		$$ = yylex.(*yyLexState).ast(STMT_LIST, $1)
	  }
	|
	  stmt_list  stmt  ';'
	  {
		$1.push_left($2)
	  }
	;

arg_list:
	  /*  empty */
	  {
		$$ = nil
	  }
	|
	  expr
	  {
		$$ = yylex.(*yyLexState).ast(ARGV, $1)
	  }
	|
	  arg_list  ','  expr
	  {
		$1.push_left($3)
	  }
	;
%%

var keyword = map[string]int{
	"and":			yy_AND,
	"as":			AS,
	"command":		COMMAND,
	"define":		DEFINE,
	"ExpandEnv":		EXPAND_ENV,
	"false":		yy_FALSE,
	"is":			yy_IS,
	"not":			NOT,
	"null":			yy_NULL,
	"or":			yy_OR,
	"run":			RUN,
	"string":		yy_STRING,
	"true":			yy_TRUE,
	"tuple":		TUPLE,
	"when":			WHEN,
}

type yyLexState struct {
	in			io.RuneReader	//  source stream
	line_no			uint32	   	//  lexical line number
	eof			bool       	//  seen eof in token stream
	peek			rune       	//  lookahead in lexer
	err			error

	ast_root		*ast
	name			string
	string
	uint64

	name2ast		map[string]*ast
	name2cmd		map[string]*command
	name2tuple		map[string]*tuple
	name2satt		map[string]*sysatt
	name2run		map[string]*ast
	depends			map[string]string

	command_ref		*command
	tuple_ref		*tuple
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

func (lex *yyLexState) ast(yy_tok int, args...*ast) *ast {
	an := &ast{
		yy_tok:		yy_tok,
		line_no:	lex.line_no,
	}
	for i, a := range args {
		if a == nil {
			continue
		}
		if i == 0 {
			an.push_left(a)
		} else if i == 1 {
			an.push_right(a)
		} else {
			an.corrupt("ast: args range > 1: %d", i)
		}
	}
	return an
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
func (lex *yyLexState) skip_comment() (err error) {
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

func (lex *yyLexState) skip_space() (c rune, eof bool, err error) {

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
		err = lex.skip_comment()
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
func (lex *yyLexState) scan_string(yylval *yySymType) (eof bool, err error) {
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
func (lex *yyLexState) scan_raw_string(yylval *yySymType) (eof bool, err error) {
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
func (lex *yyLexState) scan_word(
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
	yylval.name = w

	//  COMMAND_REF?

	if lex.name2cmd[w] != nil {
		lex.command_ref = lex.name2cmd[w]
		yylval.command_ref = lex.command_ref
		return COMMAND_REF, nil
	}

	//  TUPLE REFERENCE IN COMMAND  DEFINE

	if lex.name2tuple[w] != nil {
		lex.tuple_ref = lex.name2tuple[w]
		yylval.tuple_ref = lex.tuple_ref
		return TUPLE_REF, nil
	}

	if lex.name_is_name == false && lex.name2ast[w] != nil {
		return lex.name2ast[w].yy_tok, nil
	}

	return NAME, nil
}

func (lex *yyLexState) scan_uint64(yylval *yySymType, c rune) (err error) {
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

//  Note: move to pass2?

func (lex *yyLexState) new_rel_op(tok int, left, right *ast) (a *ast) {

	switch tok {
	case NOT:
		if left.is_bool() == false {
			lex.line_no = left.line_no
			lex.error("NOT: can not negate %s", left.yy_name())
			return nil
		}
	case yy_AND, yy_OR:
		if left.is_bool() == false {
			lex.line_no = left.line_no
			lex.error(
				"%s: left expr not bool: got %s",
				yy_name(tok),
				left.yy_name(),
			)
			return nil
		}
		if right.is_bool() == false {
			lex.line_no = right.line_no
			lex.error(
				"%s: right expr not bool: got %s",
				yy_name(tok),
				right.yy_name(),
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
				left.yy_name(),
				right.yy_name(),
			)
			return nil
		}
	case CONCAT, MATCH, NOMATCH:
		if left.is_string() == false {
			lex.line_no = left.line_no
			lex.error("%s: left is not string", left.yy_name())
			return nil
		}
		if right.is_string() == false {
			lex.line_no = right.line_no
			lex.error("%s: right is not string", right.yy_name())
			return nil
		}
	case IS_NULL, IS_NOT_NULL:
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
	c, eof, err := lex.skip_space()
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

	case c == ':':
		tok, err = lex.lookahead(':', CAST, ':')
		if err != nil {
			goto LEX_ERROR
		}
		return tok
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
		tok, err = lex.scan_word(yylval, c)
		if err != nil {
			goto LEX_ERROR
		}
		return tok

	case unicode.IsNumber(c):
		err = lex.scan_uint64(yylval, c)
		if err != nil {
			goto LEX_ERROR
		}
		return UINT64

	case c == '"':
		lno := lex.line_no	// reset line number on error

		eof, err = lex.scan_string(yylval)
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

		eof, err = lex.scan_raw_string(yylval)
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
		name2cmd:	make(map[string]*command),
		name2tuple:	make(map[string]*tuple),
		name2satt:	make(map[string]*sysatt),
		depends:	make(map[string]string),
		ast_root:	&ast{
					yy_tok:		FLOW,
					line_no:	1,
				},
	}
	yyParse(lex)

	//  check for cyclic dependencies
	var depends []string
	for key, val := range lex.depends {
		depends = append(depends, key + " " + val)
	}
	if len(depends) > 0 && tsort(depends) == nil {
		return nil, lex.mkerror("cyclic dependncy")
	}
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
