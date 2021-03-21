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

%%{
machine intReader;
include common "common.rl";

read_int = (
 ('-' @{neg = true})?
 json_uint >{start = p}
 [^.0-9eE]  @{fhold; fbreak;}
 ) @eof{
     if p == 0 {
       return 0, p, errInvalidInt
     }
     fhold; fbreak;
   }
   @err{return 0, p, errInvalidInt}
;

read_uint = (
 json_uint >{start = p}
 [^.0-9eE]  @{fhold; fbreak;}
 ) @eof{
     if p == 0 {
       return 0, p, errInvalidUInt
     }
     fhold; fbreak;
   }
   @err{return 0, p, errInvalidUInt}
;


}%%

func readInt64(data []byte) (int64, int, error) {
    cs, p := 0, 0
  	pe := len(data)
  	eof := len(data)
  	var start int
  	var neg bool

%%{

machine readInt64;
include intReader "read_machines.rl";
main := read_int;

write data; write init; write exec;
}%%

  n, err := readInt64Helper(neg, data[start:p])
  return n, p, err
}

func readInt32(data []byte) (int32, int, error) {
    cs, p := 0, 0
  	pe := len(data)
  	eof := len(data)
  	var start int
  	var neg bool

%%{

machine readInt32;
include intReader "read_machines.rl";
main := read_int;

write data; write init; write exec;
}%%

  n, err := readInt32Helper(neg, data[start:p])
  return n, p, err
}

func readUint64(data []byte) (uint64, int, error) {
    cs, p := 0, 0
  	pe := len(data)
  	eof := len(data)
  	var start int

%%{

machine readUint64;
include intReader "read_machines.rl";
main := read_uint;

write data; write init; write exec;
}%%

  n, err := readUint64Helper(data[start:p])
  return n, p, err
}

func readUint32(data []byte) (uint32, int, error) {
    cs, p := 0, 0
  	pe := len(data)
  	eof := len(data)
  	var start int

%%{

machine readUint32;
include intReader "read_machines.rl";
main := read_uint;

write data; write init; write exec;
}%%

  n, err := readUint32Helper(data[start:p])
  return n, p, err
}
