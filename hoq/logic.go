package main

//  A rummy describes temporal + tri-state logic: true, false, null, waiting
//
//  There are known knowns, there are known unknowns ...

type rummy uint8

const (
	//  channel closed
	rum_NIL = rummy(0)

	//  will eventually resolve to true, false or null
	rum_WAIT = rummy(0x1)

	//  known to be null in the sql sense
	rum_NULL = rummy(0x2)

	//  known to be false
	rum_FALSE = rummy(0x4)

	//  known to be true
	rum_TRUE = rummy(0x8)
)

//  state tables for logical and/or with SQL semantics

var and = [137]rummy{}
var or = [137]rummy{}

//  build the state tables for temporal logical AND and OR used by opcodes
//  in method flow.bool2().

func init() {

	//  some shifted constants for left hand bits

	const lw = rummy(rum_WAIT << 4)
	const ln = rummy(rum_NULL << 4)
	const lf = rummy(rum_FALSE << 4)
	const lt = rummy(rum_TRUE << 4)

	//  SQL logical AND semantics with null,
	//  applied in precedence listed below
	//
	//  false and *	=>	false
	//  * and false	=>	false
	//  null and *	=>	null
	//  * and null	=>	null
	//  *		=>	true

	//  left value is false

	and[lf|rum_WAIT] = rum_FALSE
	and[lf|rum_NULL] = rum_FALSE
	and[lf|rum_FALSE] = rum_FALSE
	and[lf|rum_TRUE] = rum_FALSE

	//  right value is false

	and[lw|rum_FALSE] = rum_FALSE
	and[ln|rum_FALSE] = rum_FALSE
	and[lt|rum_FALSE] = rum_FALSE
	and[lf|rum_FALSE] = rum_FALSE

	//  left value is true

	and[lt|rum_WAIT] = rum_WAIT
	and[lt|rum_NULL] = rum_NULL
	and[lt|rum_FALSE] = rum_FALSE
	and[lt|rum_TRUE] = rum_TRUE

	//  right value is true

	and[lw|rum_TRUE] = rum_WAIT
	and[ln|rum_TRUE] = rum_NULL
	and[lt|rum_TRUE] = rum_TRUE
	and[lf|rum_TRUE] = rum_FALSE

	//  left value is null

	and[ln|rum_NULL] = rum_NULL
	and[ln|rum_TRUE] = rum_NULL
	and[ln|rum_FALSE] = rum_FALSE
	and[ln|rum_WAIT] = rum_WAIT

	//  right value is null

	and[lt|rum_NULL] = rum_NULL
	and[lf|rum_NULL] = rum_FALSE
	and[ln|rum_NULL] = rum_NULL
	and[lw|rum_NULL] = rum_WAIT

	//  left value is waiting

	and[lw|rum_NULL] = rum_WAIT
	and[lw|rum_TRUE] = rum_WAIT
	and[lw|rum_FALSE] = rum_FALSE
	and[lw|rum_WAIT] = rum_WAIT

	//  right value is waiting

	and[lt|rum_WAIT] = rum_WAIT
	and[lf|rum_WAIT] = rum_FALSE
	and[ln|rum_WAIT] = rum_WAIT
	and[lw|rum_WAIT] = rum_WAIT

	//  SQL logical OR semantics with null,
	//  applied in precedence listed below.
	//
	//  true or *	=>	true
	//  * or true	=>	true
	//  null or *	=>	null
	//  * or null	=>	null
	//  *		=>	false

	//  left value is true

	or[lt|rum_WAIT] = rum_TRUE
	or[lt|rum_NULL] = rum_TRUE
	or[lt|rum_FALSE] = rum_TRUE
	or[lt|rum_TRUE] = rum_TRUE

	//  right value is true

	or[lw|rum_TRUE] = rum_TRUE
	or[ln|rum_TRUE] = rum_TRUE
	or[lf|rum_TRUE] = rum_TRUE
	or[lt|rum_TRUE] = rum_TRUE

	//  left value is false

	or[lf|rum_WAIT] = rum_WAIT
	or[lf|rum_NULL] = rum_NULL
	or[lf|rum_FALSE] = rum_FALSE
	or[lf|rum_TRUE] = rum_TRUE

	//  right value is false

	or[lw|rum_FALSE] = rum_WAIT
	or[ln|rum_FALSE] = rum_NULL
	or[lf|rum_FALSE] = rum_FALSE
	or[lt|rum_FALSE] = rum_TRUE

	//  left value is null

	or[ln|rum_WAIT] = rum_WAIT
	or[ln|rum_NULL] = rum_NULL
	or[ln|rum_FALSE] = rum_NULL
	or[ln|rum_TRUE] = rum_TRUE

	//  right value is null

	or[ln|rum_NULL] = rum_NULL
	or[lt|rum_NULL] = rum_TRUE
	or[lf|rum_NULL] = rum_NULL
	or[lw|rum_NULL] = rum_WAIT

	//  left value is waiting

	or[lw|rum_WAIT] = rum_WAIT
	or[lw|rum_NULL] = rum_WAIT
	or[lw|rum_TRUE] = rum_TRUE
	or[lw|rum_FALSE] = rum_WAIT

	//  right value is waiting
	or[lw|rum_WAIT] = rum_WAIT
	or[lt|rum_WAIT] = rum_TRUE
	or[lf|rum_WAIT] = rum_WAIT
	or[ln|rum_WAIT] = rum_WAIT
}

func (rum rummy) String() string {

	//  english description of rummy states

	var rum2string = [16]string{
		"NIL",
		"WAIT",
		"NULL",
		"3",
		"FALSE",
		"5", "6", "7",
		"TRUE",
		"9", "10", "11", "12", "13", "14", "15",
	}
	if rum < 16 {
		return rum2string[rum]
	}

	//  two rummy values packed into 8 bits

	return rum2string[rum>>4] + "|" + rum2string[rum&0x0F]
}

func (bv *bool_value) rummy() rummy {

	switch {
	case bv == nil:
		return rum_WAIT
	case bv.is_null:
		return rum_NULL
	case bv.bool:
		return rum_TRUE
	}
	return rum_FALSE
}
