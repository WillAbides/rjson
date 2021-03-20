package rjson

%%{
machine skipper;

include common "common.rl";

skip_json_value = (json_simple_value
  | '['@{fcall skip_array;}
  | '{'@{fcall skip_object;}
  );

skip_array := (
  json_space*
  ( skip_json_value ( json_space* ',' json_space*  skip_json_value )* )? json_space*
  ']'
  ) @{fret;}
    @eof{err = errUnexpectedEOF; fbreak;}
    @err{err = errInvalidArray; fbreak;};

skip_object := (
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

}%%

func skipValue(data []byte, stack []int) (int, []int ,error) {
  var top int
  cs, p := 0, 0
	pe := len(data)
	eof := len(data)
	var err error

%%{
machine skipValue;

include skipper "skip_machine.rl";

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
