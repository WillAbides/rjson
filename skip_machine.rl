package rjson

%%{
machine skipper;

include common "common.rl";

skip_json_value = (json_simple_value
  | '['@{fcall skip_array;}
  | '{'@{fcall skip_object;}
  );

skip_array_def = (
  json_space*
  ( skip_json_value ( json_space* ',' json_space*  skip_json_value )* )? json_space*
  ']'
  ) @{fret;}
    @eof{err = errUnexpectedEOF; fbreak;}
    @err{err = errInvalidArray; fbreak;};

skip_object_def = (
    json_space*
  (
    json_string json_space* ':' json_space* skip_json_value
    ( json_space* ',' json_space*
        json_string json_space* ':' json_space*
        skip_json_value
    )*
  )?
  json_space*
  '}'
  ) @{fret;}
    @eof{err = errUnexpectedEOF; fbreak;}
    @err{err = errInvalidObject; fbreak;};

skip_json_value_fast = (json_bool | json_null | json_string | json_number
  | '['@{fcall skip_array_fast;}
  | '{'@{fcall skip_object_fast;}
  );

skip_array_fast_def = (
  [^[\]"]* (
    json_string
    | [^[\]"]+
    | '['@{fcall skip_array_fast;}
  )*
  ']'
  ) @{fret;}
    @eof{err = errUnexpectedEOF; fbreak;}
    @err{err = errInvalidArray; fbreak;};

skip_object_fast_def = (
  [^{}"]* (
    json_string
    | [^{}"]+
    | '{'@{fcall skip_object_fast;}
  )*
  '}'
  ) @{fret;}
    @eof{err = errUnexpectedEOF; fbreak;}
    @err{err = errInvalidObject; fbreak;};

}%%

func skipValueFast(data []byte, stack []int) (int, []int ,error) {
  var top int
  cs, p := 0, 0
	pe := len(data)
	eof := len(data)
	var err error

%%{
machine skipValueFast;

include skipper "skip_machine.rl";

skip_array_fast := skip_array_fast_def;
skip_object_fast := skip_object_fast_def;

prepush {
  if top + 1 >= len(stack) {
    stack = append(stack, make([]int, 1 + top - len(stack))...)
  }
}

main := json_space* skip_json_value_fast @err{
  return p, stack, errNoValidToken
};

write data; write init;  write exec;

}%%

  return p,stack, err
}

func skipValue(data []byte, stack []int) (int, []int ,error) {
  var top int
  cs, p := 0, 0
	pe := len(data)
	eof := len(data)
	var err error

%%{
machine skipValue;

include skipper "skip_machine.rl";

skip_array := skip_array_def;
skip_object := skip_object_def;

prepush {
  if top + 1 >= len(stack) {
    stack = append(stack, make([]int, 1 + top - len(stack))...)
  }
}

main := json_space* skip_json_value @err{
  return p, stack, errNoValidToken
};

write data; write init;  write exec;

}%%

  return p,stack, err
}

func skipStringFast(data []byte) (int, error) {
  cs, p := 0, 0
  pe := len(data)
  eof := len(data)
  var err error

%%{
machine skipStringFast;
include common "common.rl";

main := ('"' ([^"\\] | '\\' any)* '"') @err{err = errInvalidString};

write data; write init;  write exec;
}%%

  return p, err
}
