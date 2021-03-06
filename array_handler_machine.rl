package rjson

func handleArrayValues(data []byte, handler ArrayValueHandler, stack []int) (int, []int, error) {
  var top, cs, p, pp int
  var err error
  pe := len(data)
  eof := len(data)

%%{
machine handleArrayValues;

include skipper "skip_machine.rl";

prepush {
  if top + 1 >= len(stack) {
    stack = append(stack, make([]int, 1 + top - len(stack))...)
  }
}

action try_handler {
  pp, err = handler.HandleArrayValue(data[p:])
  if err != nil {
    return p + pp, stack, err
  }
  if pp < 0 {
    err = errPOutOfRange
    fbreak;
  }
  if pp != 0 {
    if p + pp - 1 >= pe {
      err = errPOutOfRange
      fbreak;
    }
    fexec p + pp - 1;
  }
}

action try_handler_simple {
  _, err = handler.HandleArrayValue(data[p:])
  if err != nil {
    return p, stack, err
  }
}

skip_array := skip_array_def;
skip_object := skip_object_def;

handled_value =
 json_bool >(try_handler_simple)
 | json_null >(try_handler_simple)
 | json_number >(try_handler_simple)
 | json_string >(try_handler)
 | '[' >(try_handler) @{fcall skip_array;}
 | '{' >(try_handler) @{fcall skip_object;}
;

main :=
  ( json_space* (
  json_null |
  ('['
    (
      json_space* handled_value
      (json_space* ',' json_space* handled_value )*
    )?
    json_space* ']'
  ))) @err{
        return p, stack, errInvalidArray
      };

write data; write init;  write exec;
}%%

return p, stack, err
}
