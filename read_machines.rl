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
