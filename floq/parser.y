/*
 *  Synopsis:
 *	Build an abstract syntax tree for Yacc grammar of "floq" language.
 */

%{
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

const max_name_rune_count = 127

func init() {

	if yyToknames[3] != "__MIN_YYTOK" {
		panic("yyToknames[3] != __MIN_YYTOK: correct yacc command?")
	}
	//yyDebug = 4
}
%}

%union {
	ast		*ast
	string		string
	uint64		uint64
}

//  lowest numbered yytoken.  must be first in list.
%token	__MIN_YYTOK

%token	PARSE_ERROR
%token	CREATE
%token	SCANNER  SCANNER_REF
%token	COMMAND  CMD_REF
%token	OF  LINES
%token	NAME  UINT64  STRING
%token	FLOW  STATEMENT
%token	ATT  ATT_LIST
%token	ATT_ARRAY

%type	<uint64>	UINT64		
%type	<string>	STRING  new_name
%type	<ast>		constant
%type	<ast>		flow
%type	<ast>		att  att_list  att_value
%type	<ast>		stmt  stmt_list
%type	<ast>		att_array 
%type	<ast>		create  scanner  command

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
	;

att_array:
	  /*  empty  */
	  {
	  	$$ = &ast{
			yy_tok:		ATT_ARRAY,
			line_no:        yylex.(*yyLexState).line_no,
			array_ref:	make([]string, 0),
		}
	  }
	|
	  STRING
	  {
	  	ar := make([]string, 1)
		ar[0] = $1 
	  	$$ = &ast{
			yy_tok:		ATT_ARRAY,
			line_no:        yylex.(*yyLexState).line_no,
			array_ref:	ar,
		}
	  }
	|
	  att_array  ','  STRING
	  {
	  	lex := yylex.(*yyLexState)

		ar := $1.array_ref
		ar = append(ar, $3)
		$1.array_ref = ar
		if len(ar) > 127 {
			lex.error("attribute array > 127 elements")
			return 0
		}
		$$ = $1
	  }
	;

att_value:
	  constant
	|
	  '['  att_array  ']'
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

att_list:
	  /*  empty  */
	  {
	  	$$ = &ast{
			yy_tok:         ATT_LIST,
			line_no:        yylex.(*yyLexState).line_no,
		}
	  }
	|
	  att
	  {
		lex := yylex.(*yyLexState)

	  	al := &ast{
			yy_tok:		ATT_LIST,
			line_no:	lex.line_no,
		}
		al.left = $1
		$1.parent = al

		$$ = al
	  }
	|
	  att_list ','  att
	  {
		al := $1
	  	a := $3
		a.parent = al


		var an *ast
		for an = al.left;  an.next != nil;  an = an.next {}
		an.next = a
		a.previous = a

		$$ = $1
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
			yy_tok:		CMD_REF,
			line_no:        lex.line_no,
			cmd_ref:	&command {},
		}
	  }
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

		e := fmt.Sprintf("name already exists as scanner %s", lex.name)
		lex.Error(e)
		return 0
	  }
	|
	  CMD_REF
	  {
	  	lex := yylex.(*yyLexState)

		e := fmt.Sprintf("name already exists as command %s", lex.name)
		lex.Error(e)
		return 0
	  }
	;
stmt:
	  create  scanner  new_name
	  {
	  	lex := yylex.(*yyLexState)

	  	$2.scanner_ref.name = $3
	  	$2.parent = $1
		$1.left = $2
		lex.put_name($3, $2)

		lex.name_is_name = false

		$$ = $1
	  }
	|
	  create  command  new_name  '('
	  {
		yylex.(*yyLexState).name_is_name = true

	  } att_list  ')'
	  {
	  	lex := yylex.(*yyLexState)

		lex.name_is_name = false
		lex.put_name($3, $2)

		al := $6
		acmd := $2
		al.parent = $2
		acmd.left = al

	  	acmd.cmd_ref.name = $3
	  	acmd.parent = $1

		$1.left = $2

		//  frisk the attibutes of command

		cmd := acmd.cmd_ref
		e := cmd.frisk_att(al) 
		if e != "" {
			lex.error("command: %s: %s", cmd.name, e)
			return 0
		}

		$$ = $1
	  }
	;	
	
stmt_list:
	  /*  empty */
	  {
	  	$$ = &ast{
			yy_tok:         STATEMENT,
			line_no:        yylex.(*yyLexState).line_no,
		}
	  }
	|
	  stmt  ';'
	  {
	  	lex := yylex.(*yyLexState)

		a := &ast{
			yy_tok:		STATEMENT,
			line_no:	$1.line_no,
			left:		$1,	
			parent:		lex.ast_root,
		}
		$1.parent = a

		$$ = a
	  }
	|
	  stmt_list  stmt  ';'
	  {
	  	lex := yylex.(*yyLexState)

		s := &ast{
			yy_tok:		STATEMENT,
			line_no:	$2.line_no,
			left:		$2,	
			parent:         lex.ast_root,
		}
		$2.parent = s

		//  find end statement
		sl := $1
		for ;  sl.next != nil;  sl = sl.next {}
		sl.next = s
		s.previous = sl

		$$ = $1
	  }
	;
%%

var keyword = map[string]int{
	"command":		COMMAND,
	"create":		CREATE,
	"lines":		LINES,
	"of":			OF,
	"scanner":		SCANNER,
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
	name_is_name		bool
}


func (lex *yyLexState) put_name(name string, a *ast) {

	//  cheap sanity test
	if lex.name2ast[name] != nil {
		panic(fmt.Sprintf("put_name: overriding name2ast: %s", name))
	}
	if a.line_no == 0 {
		a.line_no = lex.line_no
	}
	lex.name2ast[name] = a
}

func (lex *yyLexState) pushback(c rune) {

	if lex.peek != 0 {
		panic("pushback(): push before peek")	/* impossible */
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

func lookahead(lex *yyLexState, peek rune, ifyes int, ifno int) (int, error) {
	
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
func (lex *yyLexState) scan_word(yylval *yySymType, c rune) (tok int, err error) {
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

//  lexical scan of a token

func (lex *yyLexState) Lex(yylval *yySymType) (tok int) {

	if lex.err != nil {
		return PARSE_ERROR
	}
	if lex.eof {
		return 0
	}
	c, eof, err := skip_space(lex)
	if err != nil {
		goto PARSE_ERROR
	}
	if eof {
		return 0
	}

	switch {

	//  ascii outside of strings, for time being (why?)
	case c > 127:
		goto PARSE_ERROR

	case unicode.IsLetter(c) || c == '_':
		tok, err = lex.scan_word(yylval, c)
		if err != nil {
			goto PARSE_ERROR
		}
		return tok

	case unicode.IsNumber(c):
		err = lex.scan_uint64(yylval, c)
		if err != nil {
			goto PARSE_ERROR
		}
		return UINT64

	case c == '"':
		lno := lex.line_no	// reset line number on error

		eof, err = lex.scan_string(yylval)
		if err != nil {
			goto PARSE_ERROR
		}
		if eof {
			lex.line_no = lno
			err = lex.mkerror("unexpected end of file in string")
			goto PARSE_ERROR
		}
		return STRING

	case c == '`':
		lno := lex.line_no	// reset line number on error

		eof, err = lex.scan_raw_string(yylval)
		if err != nil {
			goto PARSE_ERROR
		}
		if eof {
			lex.line_no = lno
			err = lex.mkerror("unexpected end of file in raw string")
			goto PARSE_ERROR
		}
		return STRING
	}
	return int(c)

PARSE_ERROR:
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
		ast_root:	&ast{
					yy_tok:		FLOW,
					line_no:	1,
				},
	}
	yyParse(lex)
	return lex.ast_root, lex.err
}
