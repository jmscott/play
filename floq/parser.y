/*
 *  Synopsis:
 *	Build an abstract syntax tree for Yacc grammar of "floq" language.
 */

%{
package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

const max_name_rune_count = 127

func init() {
	if yyToknames[3] != "__MIN_YYTOK" {
		panic("yyToknames[3] != __MIN_YYTOK: yacc may have changed")
	}
}

//  abstract syntax tree that represents the flow

type ast struct {

	yy_tok	int

	//  children
	left	*ast
	right	*ast

	//  siblings
	next	*ast
}
%}

%union {
	ast		*ast
	name		string
			string
			uint64
}

//  lowest numbered yytoken.  must be first in list.
%token	__MIN_YYTOK

%token	PARSE_ERROR
%token	EQ  NEQ  MATCH  NO_MATCH
%token  STRING  yy_STRING
%token  BOOL  yy_BOOL
%token	UINT64
%token	NAME
%token	SYNC MAP LOAD_OR_STORE SYNC_MAP_REF LOADED

%type	<string>	NAME
%type	<string>	STRING

%%
flow:
	  /*  empty */
	|
	  stmt_list
	;

declare_sync_map:
	  SYNC  MAP  NAME  '['  yy_STRING  ']'  yy_BOOL
	;
stmt:
	  declare_sync_map
	;

stmt_list:
	  stmt  ';'
	|
	  stmt_list  stmt  ';'
	;
%%

var keyword = map[string]int{
	"bool":			yy_BOOL,
	"string":		yy_STRING,
	"sync":			SYNC,
	"map":			MAP,
}

type yyLexState struct {
	in				io.RuneReader	//  config source stream
	line_no				uint64	   //  lexical line number
	eof				bool       //  seen eof in token stream
	peek				rune       //  lookahead in lexer
	err				error

	ast_root			*ast

}

func (l *yyLexState) pushback(c rune) {

	if l.peek != 0 {
		panic("pushback(): push before peek")	/* impossible */
	}
	l.peek = c
	if c == '\n' {
		l.line_no--
	}
}

/*
 *  Read next UTF8 rune.
 */
func (l *yyLexState) get() (c rune, eof bool, err error) {

	if l.eof {
		return 0, true, nil
	}
	if l.peek != 0 {		/* returned stashed char */
		c = l.peek
		/*
		 *  Only pushback 1 char.
		 */
		l.peek = 0
		if c == '\n' {
			l.line_no++
		}
		return c, false, nil
	}
	c, _, err = l.in.ReadRune()
	if err != nil {
		if err == io.EOF {
			l.eof = true
			return 0, true, nil
		}
		return 0, false, err
	}

	if c == unicode.ReplacementChar {
		return 0, false, l.mkerror("get: invalid unicode sequence")
	}
	if c == '\n' {
		l.line_no++
	}
	return c, false, nil
}

func lookahead(l *yyLexState, peek rune, ifyes int, ifno int) (int, error) {
	
	next, eof, err := l.get()
	if err != nil {
		return 0, err
	}
	if next == peek {
		return ifyes, err
	}
	if !eof {
		l.pushback(next)
	}
	return ifno, nil
}

/*
 *  Skip '#' comment.
 *  The scan stops on the terminating newline or error
 */
func skip_comment(l *yyLexState) (err error) {
	var c rune
	var eof bool

	/*
	 *  Scan for newline, end of file, or error.
	 */
	for c, eof, err = l.get();  !eof && err == nil;  c, eof, err = l.get() {
		if c == '\n' {
			return
		}
	}
	return err

}

// skip over whitespace in code, complicated by # coments.

func skip_space(l *yyLexState) (c rune, eof bool, err error) {

	for c, eof, err = l.get();  !eof && err == nil;  c, eof, err = l.get() {
		if unicode.IsSpace(c) {
			continue
		}
		if c != '#' {
			return c, false, nil
		}

		/*
		 *  Skipping over # comment terminated by newline or EOF
		 */
		err = skip_comment(l)
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
func (l *yyLexState) scan_string(yylval *yySymType) (eof bool, err error) {
	var c rune
	s := ""

	for c, eof, err = l.get();  !eof && err == nil;  c, eof, err = l.get(){
		if c == '"' {
			yylval.string = s
			return false, nil
		}
		switch c {
		case '\n':
			return false, l.mkerror("new line in string")
		case '\r':
			return false, l.mkerror("carriage return in string")
		case '\t':
			return false, l.mkerror("tab in string")
		case '\\':
			return false, l.mkerror("backslash in string")
		}
		s += string(c)
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

/*
 *  Scan an almost raw string as defined in golang.
 *  Carriage return is stripped.
 */
func (l *yyLexState) scan_raw_string(yylval *yySymType) (eof bool, err error) {
	var c rune
	s := ""

	/*
	 *  Scan a raw string of unicode letters, accepting all but `
	 */
	for c, eof, err = l.get();  !eof && err == nil;  c, eof, err = l.get() {
		
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
func (l *yyLexState) scan_word(yylval *yySymType, c rune) (tok int, err error) {
	var eof bool

	w := string(c)		//  panic() if cast fails?
	count := 1

	/*
	 *  Scan a string of unicode (?) letters, numbers/digits and '_' 
	 *  characters.
	 */
	for c, eof, err = l.get();  !eof && err == nil;  c, eof, err = l.get() {
		if c > 127 ||
		   (c != '_' &&
		   !unicode.IsLetter(c) &&
		   !unicode.IsNumber(c)) {
			break
		}
		count++
		if count > max_name_rune_count {
			return 0, l.mkerror("word: too many chars: max=%d",
						max_name_rune_count)
		}
		w += string(c)		//  Note: replace with string builder?
	}
	if err != nil {
		return 0, err
	}
	if !eof {
		l.pushback(c)		/* first character after word */
	}

	if keyword[w] > 0 {		/* got a keyword */
		return keyword[w], nil	/* return yacc generated token */
	}

	yylval.string = w
	return NAME, nil
}

func (l *yyLexState) scan_uint64(yylval *yySymType, c rune) (err error) {
	var eof bool

	ui64 := string(c)
	count := 1

	/*
	 *  Scan a string of unicode numbers/digits and let Scanf parse the
	 *  actual digit string.
	 */
	for c, eof, err = l.get();  !eof && err == nil;  c, eof, err = l.get() {
		count++
		if count > 20 {
			return l.mkerror("uint64 > 20 digits")
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
		l.pushback(c)		//  first character after ui64
	}

	yylval.uint64, err = strconv.ParseUint(ui64, 10, 64)
	return
}

func (l *yyLexState) Lex(yylval *yySymType) (tok int) {

	if l.err != nil {
		return PARSE_ERROR
	}
	if l.eof {
		return 0
	}
	c, eof, err := skip_space(l)
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
		tok, err = l.scan_word(yylval, c)
		if err != nil {
			goto PARSE_ERROR
		}
		return tok

	case unicode.IsNumber(c):
		err = l.scan_uint64(yylval, c)
		if err != nil {
			goto PARSE_ERROR
		}
		return UINT64

	case c == '"':
		lno := l.line_no	// reset line number on error

		eof, err = l.scan_string(yylval)
		if err != nil {
			goto PARSE_ERROR
		}
		if eof {
			l.line_no = lno
			err = l.mkerror("unexpected end of file in string")
			goto PARSE_ERROR
		}
		return STRING

	case c == '`':
		lno := l.line_no	// reset line number on error

		eof, err = l.scan_raw_string(yylval)
		if err != nil {
			goto PARSE_ERROR
		}
		if eof {
			l.line_no = lno
			err = l.mkerror("unexpected end of file in raw string")
			goto PARSE_ERROR
		}
		return STRING

	case c == '=':

		//  string equality: ==

		tok, err = lookahead(l, '=', EQ, 0)
		if err != nil {
			goto PARSE_ERROR
		}
		if tok == EQ {
			return tok
		}

		//  match regular expression: =~

		tok, err = lookahead(l, '~', MATCH, '=')
		if err != nil {
			goto PARSE_ERROR
		}
		return tok

	case c == '!':

		//  string inequality: !=

		tok, err = lookahead(l, '=', NEQ, 0)
		if err != nil {
			goto PARSE_ERROR
		}
		if tok == NEQ {
			return tok
		}

		//  regular expression not matches: !~

		tok, err = lookahead(l, '~', NO_MATCH, '=')
		if err != nil {
			goto PARSE_ERROR
		}
		return tok
	}
	return int(c)

PARSE_ERROR:
	l.err = err
	return PARSE_ERROR
}

func (l *yyLexState) mkerror(format string, args...interface{}) error {

	return errors.New(fmt.Sprintf("%s, near line %d",
		fmt.Sprintf(format, args...),
		l.line_no,
	))
}

func (l *yyLexState) error(format string, args...interface{}) {

	l.Error(fmt.Sprintf(format, args...))
}

func (l *yyLexState) Error(msg string) {

	if l.err == nil {			//  only report first error
		l.err = l.mkerror("%s", msg)
	}
}

func parse(in io.RuneReader) (*ast, error) {

	lex := &yyLexState{
		in:		in,
		line_no:	1,
	}
	yyParse(lex)
	return lex.ast_root, lex.err
}
