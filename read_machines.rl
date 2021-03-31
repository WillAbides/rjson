package rjson

func readNull(data []byte) (int, error) {
  cs, p := 0, 0
  pe := len(data)
  eof := len(data)

%%{

machine readNull;
include common "common.rl";

main := (json_space* json_null)@err{return p, errNotNull};

write data; write init; write exec;
}%%

  return p, nil
}

func readBool(data []byte) (bool, int, error) {
  cs, p := 0, 0
  pe := len(data)
  eof := len(data)
  var val bool

%%{

machine readBool;
include common "common.rl";

main := (
  json_space*
  (
    json_true @{val = true}
    | json_false @{val = false}
  )
)@err{return false, p, errNotBool}
;

write data; write init; write exec;
}%%

  return val, p, nil
}

func readFloat64(data []byte) (float64, int, error) {
      cs, p := 0, 0
    	pe := len(data)
    	eof := len(data)
    	var start int
    	var hasDecimal bool
    	var hasExp bool

%%{

machine readFloat64;
include common "common.rl";

main := (
  json_int >{start = p}
  ('.'[0-9]+)? @{hasDecimal = true}
  ([eE][+\-]?[0-9]+)? @{hasExp = true}
  ) @err{return 0, p, errInvalidNumber};

write data; write init; write exec;
}%%

  n, err := readFloat64Helper(hasDecimal, hasExp, data[start:p])
  return n, p, err
}
